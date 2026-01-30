package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
	"webchat/internal/config"
	"webchat/internal/models"
	"webchat/internal/repository"
	. "webchat/internal/utils"
	"webchat/internal/ws"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo      *repository.Repository
	config    *config.Config
	jwtSecret string
	hub       *ws.Hub
}

func NewService(repo *repository.Repository, config *config.Config, jwtSecret string, hub *ws.Hub) *Service {
	return &Service{
		Repo:      repo,
		config:    config,
		jwtSecret: jwtSecret,
		hub:       hub,
	}
}

func (s *Service) Register(phone, password string) (string, error) {
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return "", errors.New("неверный формат телефона")
	}

	existing, _ := s.Repo.GetUserByPhone(normalizedPhone)
	if existing != nil {
		return "", errors.New("пользователь уже существует")
	}

	tempPassword := &models.TempPassword{
		Phone:     normalizedPhone,
		Password:  password,
		CreatedAt: time.Now(),
	}

	if err := s.Repo.CreateTempPassword(tempPassword); err != nil {
		return "", errors.New("не удалось сохранить данные")
	}

	code := GenerateCode()

	registrationCode := &models.RegistrationCode{
		Phone:     normalizedPhone,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.Repo.CreateRegistrationCode(registrationCode); err != nil {
		return "", errors.New("не удалось создать код подтверждения")
	}

	fmt.Printf("Код регистрации для %s: %s\n", normalizedPhone, code)
	return "", nil
}

func (s *Service) Login(phone, password string) (string, error) {
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return "", errors.New("неверный формат телефона")
	}

	user, err := s.Repo.GetUserByPhone(normalizedPhone)
	if err != nil {
		return "", errors.New("неверные телефон или пароль")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("неверные телефон или пароль")
	}

	if err := s.Repo.UpdateUserLastSeen(user.ID); err != nil {
		return "", err
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return "", err
	}

	if err := s.Repo.CreateSession(&models.UserSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Logout(token string) error {
	return s.Repo.DeleteSession(token)
}

func (s *Service) LogoutAll(userID uint) error {
	return s.Repo.DeleteAllUserSessions(userID)
}

func (s *Service) GetCurrentUser(token string) (*models.User, error) {
	session, err := s.Repo.GetSessionByToken(token)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	user, err := s.Repo.GetUserByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *Service) AddContact(userID uint, contactPhone string) (*models.Chat, error) {
	normalizedPhone, err := NormalizePhone(contactPhone)
	if err != nil {
		return nil, errors.New("неверный формат телефона")
	}

	contactUser, err := s.Repo.GetUserByPhone(normalizedPhone)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if userID == contactUser.ID {
		return nil, errors.New("нельзя добавить самого себя")
	}

	existingChat, _ := s.Repo.GetPrivateChat(userID, contactUser.ID)
	if existingChat != nil {
		return existingChat, nil
	}

	return s.CreatePrivateChat(userID, contactUser.ID)
}

func (s *Service) SearchUserByPhone(phone string) (*models.User, error) {
	return s.Repo.GetUserByPhone(phone)
}

func (s *Service) GetChats(userID uint) ([]models.Chat, error) {
	log.Printf("Загружаем чаты для пользователя %d", userID)

	chats, err := s.Repo.GetUserChats(userID)
	if err != nil {
		log.Printf("Ошибка при загрузке чатов: %v", err)
		return nil, err
	}

	log.Printf("Найдено %d чатов", len(chats))

	for i := range chats {
		lastMessage, err := s.Repo.GetLastChatMessage(chats[i].ID)
		if err != nil {
			log.Printf("Ошибка при загрузке последнего сообщения для чата %d: %v", chats[i].ID, err)
			chats[i].LastMessage = nil
		} else {
			log.Printf("Последнее сообщение для чата %d: %v", chats[i].ID, lastMessage)
			chats[i].LastMessage = lastMessage
		}
	}

	return chats, nil
}

func (s *Service) CreatePrivateChat(userID1, userID2 uint) (*models.Chat, error) {
	if userID1 == userID2 {
		return nil, errors.New("нельзя создать чат с самим собой")
	}

	existingChat, err := s.Repo.GetPrivateChat(userID1, userID2)
	if err != nil {
		return nil, err
	}

	if existingChat != nil {
		return existingChat, nil
	}

	chat := &models.Chat{
		Type:      "private",
		CreatedBy: userID1,
	}

	if err := s.Repo.CreateChat(chat); err != nil {
		return nil, errors.New("ошибка создания чата")
	}

	if err := s.Repo.AddChatMember(chat.ID, userID1, false); err != nil {
		return nil, errors.New("ошибка добавления в чат")
	}

	if err := s.Repo.AddChatMember(chat.ID, userID2, false); err != nil {
		return nil, errors.New("ошибка добавления в чат")
	}

	user1, _ := s.Repo.GetUserByID(userID1)
	user2, _ := s.Repo.GetUserByID(userID2)

	// Системное сообщение о создании приватного чата
	systemMessage := &models.Message{
		ChatID:   chat.ID,
		SenderID: userID1,
		Content:  fmt.Sprintf("Приватный чат создан между %s и %s", user1.Username, user2.Username),
		Type:     "system_chat_created",
	}

	if err := s.Repo.CreateMessage(systemMessage); err != nil {
		return nil, err
	}

	// Отправляем уведомление второму пользователю
	s.hub.SendToUser(userID2, models.WSMessage{
		Type: "chat_created",
		Data: map[string]interface{}{
			"chat_id": chat.ID,
			"type":    "private",
			"with_user": map[string]interface{}{
				"id":       userID1,
				"phone":    user1.Phone,
				"username": user1.Username,
			},
			"unread_count": 1,
			"message":      fmt.Sprintf("%s создал приватный чат с вами", user1.Username),
			"timestamp":    time.Now().Unix(),
		},
	})

	return s.Repo.GetChatByID(chat.ID)
}
func (s *Service) CreateGroupChat(creatorID uint, name string, memberPhones []string, isSearchable bool) (*models.Chat, error) {
	chat := &models.Chat{
		Name:         name,
		Type:         "group",
		CreatedBy:    creatorID,
		IsSearchable: isSearchable,
	}

	if err := s.Repo.CreateChat(chat); err != nil {
		return nil, err
	}

	if err := s.Repo.AddChatMember(chat.ID, creatorID, true); err != nil {
		return nil, err
	}

	creator, _ := s.Repo.GetUserByID(creatorID)

	systemMessage := &models.Message{
		ChatID:   chat.ID,
		SenderID: creatorID,
		Content:  fmt.Sprintf("%s создал группу '%s'", creator.Username, chat.Name),
		Type:     "system_chat_created",
	}
	s.Repo.CreateMessage(systemMessage)

	if len(memberPhones) > 0 {
		users, err := s.Repo.FindUsersByPhones(memberPhones)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			if user.ID != creatorID {
				if err := s.Repo.AddChatMember(chat.ID, user.ID, false); err != nil {
					continue
				}
				s.hub.SendToUser(user.ID, models.WSMessage{
					Type: "chat_created",
					Data: map[string]interface{}{
						"chat_id":   chat.ID,
						"chat_name": chat.Name,
						"chat_type": "group",
						"creator": map[string]interface{}{
							"id":       creator.ID,
							"phone":    creator.Phone,
							"username": creator.Username,
						},
						"unread_count": 1,
						"message":      fmt.Sprintf("%s создал группу '%s'", creator.Username, chat.Name),
						"timestamp":    time.Now().Unix(),
					},
				})
			}
		}
	}

	return s.Repo.GetChatByID(chat.ID)
}

func (s *Service) sendWelcomeMessage(chatID, senderID uint, welcomeText string) error {
	message := &models.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  welcomeText,
	}

	return s.Repo.CreateMessage(message)
}

func (s *Service) GetChat(chatID uint) (*models.Chat, error) {
	return s.Repo.GetChatByID(chatID)
}

func (s *Service) AddChatMember(chatID, adderID uint, phone string) error {
	isMember, err := s.Repo.IsChatMember(chatID, adderID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("не являетесь участником чата")
	}

	user, err := s.Repo.GetUserByPhone(phone)
	if err != nil {
		return errors.New("пользователь не найден")
	}

	if user.ID == adderID {
		return errors.New("нельзя добавить самого себя")
	}

	existingMember, _ := s.Repo.IsChatMember(chatID, user.ID)
	if existingMember {
		return errors.New("пользователь уже в чате")
	}

	if err := s.Repo.AddChatMember(chatID, user.ID, false); err != nil {
		return errors.New("не удалось добавить пользователя")
	}

	chat, _ := s.Repo.GetChatByID(chatID)
	adder, _ := s.Repo.GetUserByID(adderID)

	systemMessage := &models.Message{
		ChatID:   chatID,
		SenderID: adderID,
		Content:  fmt.Sprintf("%s добавил %s в чат", adder.Username, user.Username),
		Type:     "system_user_added",
	}

	if err := s.Repo.CreateMessage(systemMessage); err != nil {
		return err
	}

	for _, member := range chat.Members {
		if member.ID != adderID {
			unreadCount, _ := s.Repo.GetUnreadCount(chatID, member.ID)

			s.hub.SendToUser(member.ID, models.WSMessage{
				Type: "new_message",
				Data: map[string]interface{}{
					"chat_id":   chatID,
					"chatName":  chat.Name,
					"chat_type": chat.Type,
					"message": map[string]interface{}{
						"id":         systemMessage.ID,
						"chat_id":    chatID,
						"sender_id":  adderID,
						"content":    systemMessage.Content,
						"type":       systemMessage.Type,
						"created_at": time.Now(),
					},
					"sender": map[string]interface{}{
						"id":       adder.ID,
						"phone":    adder.Phone,
						"username": adder.Username,
					},
					"unread_count": unreadCount,
					"timestamp":    time.Now().Unix(),
				},
			})
		}
	}

	s.hub.SendToUser(user.ID, models.WSMessage{
		Type: "chat_created",
		Data: map[string]interface{}{
			"chat_id":   chatID,
			"chat_name": chat.Name,
			"chat_type": chat.Type,
			"adder": map[string]interface{}{
				"id":       adder.ID,
				"phone":    adder.Phone,
				"username": adder.Username,
			},
			"unread_count": 1,
			"message":      fmt.Sprintf("%s добавил вас в чат '%s'", adder.Username, chat.Name),
			"timestamp":    time.Now().Unix(),
		},
	})

	return nil
}

func (s *Service) RemoveChatMember(chatID, removerID, userID uint) error {
	isMember, err := s.Repo.IsChatMember(chatID, removerID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("not a chat member")
	}

	members, err := s.Repo.GetChatMembers(chatID)
	if err != nil {
		return err
	}

	var isRemoverAdmin bool
	for _, m := range members {
		if m.UserID == removerID {
			isRemoverAdmin = m.IsAdmin
			break
		}
	}

	if !isRemoverAdmin {
		return errors.New("only admin can remove members")
	}

	if removerID == userID {
		return errors.New("cannot remove yourself")
	}

	return s.Repo.RemoveChatMember(chatID, userID)
}

func (s *Service) SendMessage(chatID, senderID uint, content string, files []*multipart.FileHeader) (*models.Message, error) {
	isMember, err := s.Repo.IsChatMember(chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("не являетесь участником чата")
	}

	if len(content) > 5000 {
		return nil, errors.New("сообщение слишком длинное")
	}

	message := &models.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  content,
	}

	if err := s.Repo.CreateMessage(message); err != nil {
		return nil, err
	}

	for _, file := range files {
		if file != nil {
			if file.Size > s.config.MaxFileSize {
				continue
			}

			uploadDir := s.config.UploadDir
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				continue
			}

			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
			filepath := filepath.Join(uploadDir, filename)

			src, err := file.Open()
			if err != nil {
				continue
			}
			defer src.Close()

			dst, err := os.Create(filepath)
			if err != nil {
				continue
			}
			defer dst.Close()

			if _, err := io.Copy(dst, src); err != nil {
				continue
			}

			messageFile := &models.MessageFile{
				MessageID: message.ID,
				Filename:  file.Filename,
				Filepath:  filename,
				Filesize:  file.Size,
				MimeType:  file.Header.Get("Content-Type"),
			}

			s.Repo.AttachFileToMessage(messageFile)
		}
	}

	messageWithDetails, err := s.Repo.GetMessageWithDetails(message.ID)
	if err != nil {
		return nil, err
	}

	chat, err := s.Repo.GetChatByID(chatID)
	if err != nil {
		return messageWithDetails, nil
	}

	sender, _ := s.Repo.GetUserByID(senderID)

	for _, member := range chat.Members {
		if member.ID != senderID {
			unreadCount, _ := s.Repo.GetUnreadCount(chatID, member.ID)

			s.hub.SendToUser(member.ID, models.WSMessage{
				Type: "new_message",
				Data: map[string]interface{}{
					"chat_id":   chatID,
					"chatName":  chat.Name,
					"chat_type": chat.Type,
					"message":   messageWithDetails,
					"sender": map[string]interface{}{
						"id":       sender.ID,
						"phone":    sender.Phone,
						"username": sender.Username,
					},
					"unread_count": unreadCount,
					"timestamp":    time.Now().Unix(),
				},
			})
		}
	}

	return messageWithDetails, nil
}
func (s *Service) IsChatMember(chatID, userID uint) (bool, error) {
	return s.Repo.IsChatMember(chatID, userID)
}

func (s *Service) GetChatMessages(chatID, userID uint, limit int) ([]models.Message, error) {
	isMember, err := s.Repo.IsChatMember(chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("not a chat member")
	}

	return s.Repo.GetChatMessages(chatID, limit)
}

func (s *Service) generateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *Service) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(float64); ok {
			return uint(userID), nil
		}
	}

	return 0, errors.New("invalid token")
}

func (s *Service) UpdateProfile(userID uint, username string) error {
	updates := make(map[string]interface{})

	if username != "" {
		if len(username) > 50 {
			return errors.New("username too long")
		}
		updates["username"] = username
	}

	if len(updates) == 0 {
		return errors.New("nothing to update")
	}

	return s.Repo.UpdateUser(userID, updates)
}

func (s *Service) GetUnreadCount(chatID, userID uint) (int64, error) {
	count, err := s.Repo.GetUnreadCount(chatID, userID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Service) MarkMessageAsRead(chatID, messageID, userID uint) error {
	isMember, err := s.Repo.IsChatMember(chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("не являетесь участником чата")
	}

	err = s.Repo.MarkMessageAsRead(messageID, userID)
	if err != nil {
		return err
	}

	chat, err := s.Repo.GetChatByID(chatID)
	if err != nil {
		return err
	}

	for _, member := range chat.Members {
		s.hub.SendToUser(member.ID, models.WSMessage{
			Type: "message_read",
			Data: map[string]interface{}{
				"chat_id":    chatID,
				"message_id": messageID,
				"reader_id":  userID,
				"timestamp":  time.Now().Unix(),
			},
		})
	}

	return nil
}
func (s *Service) MarkChatMessagesAsRead(chatID, userID uint) error {
	return s.Repo.MarkChatMessagesAsRead(chatID, userID)
}

func (s *Service) CreatePublicChat(creatorID uint, name string) (*models.Chat, error) {
	chat := &models.Chat{
		Name:      name,
		Type:      "public",
		CreatedBy: creatorID,
	}

	if err := s.Repo.CreateChat(chat); err != nil {
		return nil, err
	}

	if err := s.Repo.AddChatMember(chat.ID, creatorID, true); err != nil {
		return nil, err
	}

	return s.Repo.GetChatByID(chat.ID)
}

func (s *Service) GetPublicChats() ([]models.Chat, error) {
	var chats []models.Chat
	err := s.Repo.Db.Where("type = ?", "public").Preload("Members").Find(&chats).Error
	return chats, err
}

func (s *Service) VerifyCode(phone, code, password string) (string, error) {
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return "", errors.New("неверный формат телефона")
	}

	registrationCode, err := s.Repo.GetRegistrationCode(normalizedPhone, code)
	if err != nil {
		return "", errors.New("неверный код или срок действия истек")
	}

	if time.Now().After(registrationCode.ExpiresAt) {
		return "", errors.New("срок действия кода истек")
	}

	tempPassword, err := s.Repo.GetTempPassword(normalizedPhone)
	if err != nil {
		return "", errors.New("данные регистрации не найдены")
	}

	if password != tempPassword.Password {
		return "", errors.New("неверный пароль")
	}

	s.Repo.DeleteRegistrationCode(registrationCode.ID)
	s.Repo.DeleteTempPassword(tempPassword.ID)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ошибка создания пользователя")
	}

	user := &models.User{
		Phone:        normalizedPhone,
		PasswordHash: string(hashedPassword),
		Username:     normalizedPhone,
	}

	if err := s.Repo.CreateUser(user); err != nil {
		return "", errors.New("ошибка создания пользователя")
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return "", err
	}

	if err := s.Repo.CreateSession(&models.UserSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) ResendCode(phone string) error {
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return errors.New("неверный формат телефона")
	}

	oldCode, _ := s.Repo.GetRegistrationCodeByPhone(normalizedPhone)
	if oldCode != nil {
		s.Repo.DeleteRegistrationCode(oldCode.ID)
	}

	code := GenerateCode()

	registrationCode := &models.RegistrationCode{
		Phone:     normalizedPhone,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.Repo.CreateRegistrationCode(registrationCode); err != nil {
		return errors.New("не удалось отправить код")
	}

	fmt.Printf("New registration code for %s: %s\n", normalizedPhone, code)
	return nil
}

func (s *Service) RespondToChatInvite(inviteID, userID uint, status string) error {
	invite, err := s.Repo.GetChatInviteByID(inviteID)
	if err != nil {
		return errors.New("приглашение не найдено")
	}

	if invite.UserID != userID {
		return errors.New("неавторизованный доступ")
	}

	if err := s.Repo.UpdateChatInviteStatus(inviteID, status); err != nil {
		return err
	}

	if status == "accepted" {
		if err := s.Repo.AddChatMember(invite.ChatID, userID, false); err != nil {
			return err
		}
	}

	chat, _ := s.Repo.GetChatByID(invite.ChatID)
	user, _ := s.Repo.GetUserByID(userID)

	wsMessage := models.WSMessage{
		Type: "chat_invite_" + status,
		Data: map[string]interface{}{
			"chat_id":   invite.ChatID,
			"chat_name": chat.Name,
			"user_id":   userID,
			"user": map[string]interface{}{
				"id":       user.ID,
				"phone":    user.Phone,
				"username": user.Username,
			},
		},
	}

	s.hub.SendToUser(invite.InviterID, wsMessage)

	return nil
}

func (s *Service) SearchPublicChats(search string) ([]models.Chat, error) {
	return s.Repo.SearchPublicChats(search)
}

func (s *Service) SendChatJoinRequest(userID, chatID uint) error {
	isMember, err := s.Repo.IsChatMember(chatID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("вы уже участник этого чата")
	}

	chat, err := s.Repo.GetChatByID(chatID)
	if err != nil {
		return errors.New("чат не найден")
	}

	if !chat.IsSearchable {
		return errors.New("этот чат не принимает заявки на вступление")
	}

	existingRequest, _ := s.Repo.GetChatJoinRequestByUserAndChat(userID, chatID)
	if existingRequest != nil {
		if existingRequest.Status == "pending" {
			return errors.New("заявка уже отправлена")
		}
		if existingRequest.Status == "rejected" {
			s.Repo.DeleteChatJoinRequest(existingRequest.ID)
		}
	}

	joinRequest := &models.ChatJoinRequest{
		ChatID: chatID,
		UserID: userID,
		Status: "pending",
	}

	if err := s.Repo.CreateChatJoinRequest(joinRequest); err != nil {
		return errors.New("не удалось отправить заявку")
	}
	adminID := chat.CreatedBy
	user, _ := s.Repo.GetUserByID(userID)

	s.hub.SendToUser(adminID, models.WSMessage{
		Type: "chat_join_request",
		Data: map[string]interface{}{
			"request_id": joinRequest.ID,
			"chat_id":    chatID,
			"chat_name":  chat.Name,
			"user": map[string]interface{}{
				"id":       userID,
				"phone":    user.Phone,
				"username": user.Username,
			},
		},
	})

	return nil
}

func (s *Service) RespondToChatJoinRequest(requestID, adminID uint, status string) error {
	request, err := s.Repo.GetChatJoinRequest(requestID)
	if err != nil {
		return errors.New("заявка не найдена")
	}

	isAdmin, err := s.Repo.IsChatAdmin(request.ChatID, adminID)
	if err != nil || !isAdmin {
		return errors.New("недостаточно прав")
	}

	if err := s.Repo.UpdateChatJoinRequestStatus(requestID, status); err != nil {
		return err
	}

	if status == "accepted" {
		if err := s.Repo.AddChatMember(request.ChatID, request.UserID, false); err != nil {
			return err
		}
	}

	chat, _ := s.Repo.GetChatByID(request.ChatID)
	user, _ := s.Repo.GetUserByID(request.UserID)

	wsMessage := models.WSMessage{
		Type: "chat_join_request_" + status,
		Data: map[string]interface{}{
			"chat_id":    request.ChatID,
			"chat_name":  chat.Name,
			"request_id": request.ID,
			"user": map[string]interface{}{
				"id":       user.ID,
				"phone":    user.Phone,
				"username": user.Username,
			},
		},
	}

	s.hub.SendToUser(request.UserID, wsMessage)

	return nil
}

func (s *Service) UpdateChatVisibility(chatID, userID uint, isSearchable bool) error {
	isAdmin, err := s.Repo.IsChatAdmin(chatID, userID)
	if err != nil || !isAdmin {
		return errors.New("недостаточно прав")
	}

	return s.Repo.UpdateChatVisibility(chatID, isSearchable)
}

func (s *Service) GetChatJoinRequests(chatID, userID uint) ([]models.ChatJoinRequest, error) {
	isAdmin, err := s.Repo.IsChatAdmin(chatID, userID)
	if err != nil || !isAdmin {
		return nil, errors.New("недостаточно прав")
	}

	return s.Repo.GetChatJoinRequests(chatID)
}

func (s *Service) GetUserChatJoinRequests(userID uint) ([]models.ChatJoinRequest, error) {
	return s.Repo.GetUserChatJoinRequests(userID)
}

func (s *Service) UpdateMessage(messageID, userID uint, content string) error {
	message, err := s.Repo.GetMessageByID(messageID)
	if err != nil {
		return errors.New("сообщение не найдено")
	}

	if message.SenderID != userID {
		return errors.New("нельзя редактировать чужое сообщение")
	}

	if message.IsDeleted {
		return errors.New("сообщение удалено")
	}

	if len(content) > 5000 {
		return errors.New("сообщение слишком длинное")
	}

	return s.Repo.UpdateMessage(messageID, content)
}

func (s *Service) GetUnreadMessages(chatID, userID uint) ([]models.Message, error) {
	return s.Repo.GetUnreadMessages(chatID, userID)
}

func (s *Service) DeleteMessage(chatID, messageID, userID uint, isAdmin bool) error {
	isMember, err := s.Repo.IsChatMember(chatID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("не являетесь участником чата")
	}

	message, err := s.Repo.GetMessageByID(messageID)
	if err != nil {
		return errors.New("сообщение не найдено")
	}

	if message.ChatID != chatID {
		return errors.New("сообщение не из этого чата")
	}

	chat, _ := s.Repo.GetChatByID(chatID)

	canDelete := false

	if message.SenderID == userID {
		canDelete = true
	} else if chat.Type == "private" {
		canDelete = true
	} else if isAdmin {
		canDelete = true
	}

	if !canDelete {
		return errors.New("недостаточно прав для удаления сообщения")
	}

	if err := s.Repo.MarkMessageAsDeleted(messageID); err != nil {
		return errors.New("не удалось удалить сообщение")
	}

	chat, _ = s.Repo.GetChatByID(chatID)
	deleter, _ := s.Repo.GetUserByID(userID)

	for _, member := range chat.Members {
		s.hub.SendToUser(member.ID, models.WSMessage{
			Type: "message_deleted",
			Data: map[string]interface{}{
				"chat_id":    chatID,
				"message_id": messageID,
				"deleted_by": userID,
				"deleter": map[string]interface{}{
					"id":       deleter.ID,
					"phone":    deleter.Phone,
					"username": deleter.Username,
				},
				"timestamp": time.Now().Unix(),
			},
		})
	}

	return nil
}

func (s *Service) EditMessage(chatID, messageID, userID uint, newContent string) (*models.Message, error) {
	isMember, err := s.Repo.IsChatMember(chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("не являетесь участником чата")
	}

	message, err := s.Repo.GetMessageByID(messageID)
	if err != nil {
		return nil, errors.New("сообщение не найдено")
	}

	if message.ChatID != chatID {
		return nil, errors.New("сообщение не из этого чата")
	}

	if message.SenderID != userID {
		return nil, errors.New("нельзя редактировать чужое сообщение")
	}

	if message.IsDeleted {
		return nil, errors.New("нельзя редактировать удаленное сообщение")
	}

	if len(newContent) > 5000 {
		return nil, errors.New("сообщение слишком длинное")
	}

	if err := s.Repo.UpdateMessageContent(messageID, newContent); err != nil {
		return nil, errors.New("не удалось обновить сообщение")
	}

	updatedMessage, err := s.Repo.GetMessageWithDetails(messageID)
	if err != nil {
		return nil, err
	}

	chat, _ := s.Repo.GetChatByID(chatID)
	editor, _ := s.Repo.GetUserByID(userID)

	for _, member := range chat.Members {
		s.hub.SendToUser(member.ID, models.WSMessage{
			Type: "message_edited",
			Data: map[string]interface{}{
				"chat_id":    chatID,
				"message_id": messageID,
				"message":    updatedMessage,
				"edited_by":  userID,
				"editor": map[string]interface{}{
					"id":       editor.ID,
					"phone":    editor.Phone,
					"username": editor.Username,
				},
				"timestamp": time.Now().Unix(),
			},
		})
	}

	return updatedMessage, nil
}

func (s *Service) IsChatAdmin(chatID, userID uint) (bool, error) {
	return s.Repo.IsChatAdmin(chatID, userID)
}
