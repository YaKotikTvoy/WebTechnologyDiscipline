package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
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
	return s.Repo.GetUserChats(userID)
}

func (s *Service) CreatePrivateChat(userID1, userID2 uint) (*models.Chat, error) {
	if userID1 == userID2 {
		return nil, errors.New("нельзя создать чат с самим собой")
	}

	existingChat, _ := s.Repo.GetPrivateChat(userID1, userID2)
	if existingChat != nil {
		return existingChat, nil
	}

	chat := &models.Chat{
		Type:      "private",
		CreatedBy: userID1,
	}

	if err := s.Repo.CreateChat(chat); err != nil {
		return nil, err
	}

	if err := s.Repo.AddChatMember(chat.ID, userID1, false); err != nil {
		return nil, err
	}
	if err := s.Repo.AddChatMember(chat.ID, userID2, false); err != nil {
		return nil, err
	}

	s.hub.SendToUser(userID2, models.WSMessage{
		Type: "chat_created",
		Data: map[string]interface{}{
			"chat_id": chat.ID,
			"type":    "private",
			"with_user": map[string]interface{}{
				"id": userID1,
			},
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

	if len(memberPhones) > 0 {
		users, err := s.Repo.FindUsersByPhones(memberPhones)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			if user.ID != creatorID {
				invite := &models.ChatInvite{
					ChatID:    chat.ID,
					InviterID: creatorID,
					UserID:    user.ID,
					Status:    "pending",
				}
				if err := s.Repo.CreateChatInvite(invite); err != nil {
					return nil, err
				}
			}
		}
	}

	return s.Repo.GetChatByID(chat.ID)
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

	existingInvite, _ := s.Repo.GetChatInvite(chatID, user.ID)
	if existingInvite != nil && existingInvite.Status == "pending" {
		return errors.New("приглашение уже отправлено")
	}

	invite := &models.ChatInvite{
		ChatID:    chatID,
		InviterID: adderID,
		UserID:    user.ID,
		Status:    "pending",
	}

	if err := s.Repo.CreateChatInvite(invite); err != nil {
		if strings.Contains(err.Error(), "chat_invites_unique_pending") {
			return errors.New("приглашение уже отправлено")
		}
		return errors.New("не удалось отправить приглашение")
	}

	chat, _ := s.Repo.GetChatByID(chatID)

	s.hub.SendToUser(user.ID, models.WSMessage{
		Type: "chat_invite",
		Data: map[string]interface{}{
			"invite_id": invite.ID,
			"chat_id":   chatID,
			"chat_name": chat.Name,
			"inviter": map[string]interface{}{
				"id":       adderID,
				"phone":    phone,
				"username": user.Username,
			},
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

func (s *Service) DeleteMessage(messageID, userID uint) error {
	return s.Repo.DeleteMessage(messageID, userID)
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
	fmt.Printf("DEBUG: MarkMessageAsRead - chat: %d, message: %d, user: %d\n",
		chatID, messageID, userID)

	isMember, err := s.Repo.IsChatMember(chatID, userID)
	if err != nil {
		fmt.Printf("DEBUG: IsChatMember error: %v\n", err)
		return err
	}
	if !isMember {
		fmt.Printf("DEBUG: User %d is not a member of chat %d\n", userID, chatID)
		return errors.New("не являетесь участником чата")
	}

	fmt.Printf("DEBUG: Calling Repo.MarkMessageAsRead\n")
	err = s.Repo.MarkMessageAsRead(messageID, userID)
	if err != nil {
		fmt.Printf("DEBUG: Repo.MarkMessageAsRead error: %v\n", err)
		return err
	}

	message, err := s.Repo.GetMessageByID(messageID)
	if err != nil {
		fmt.Printf("DEBUG: GetMessageByID error: %v\n", err)
		return err
	}

	if message.SenderID != userID {
		fmt.Printf("DEBUG: Sending WebSocket notification to sender %d\n", message.SenderID)
		s.hub.SendToUser(message.SenderID, models.WSMessage{
			Type: "message_read",
			Data: map[string]interface{}{
				"chat_id":    chatID,
				"message_id": messageID,
				"reader_id":  userID,
				"timestamp":  time.Now().Unix(),
			},
		})
	}

	fmt.Printf("DEBUG: MarkMessageAsRead completed successfully\n")
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

//func (s *Service) GetFriendRequests(userID uint) ([]models.FriendRequest, error) {
//	return s.Repo.GetFriendRequests(userID)
//}

/*
func (s *Service) RespondToFriendRequest(requestID uint, userID uint, status string) error {
	request, err := s.Repo.GetFriendRequestByID(requestID)
	if err != nil {
		return errors.New("friend request not found")
	}

	if request.RecipientID != userID {
		return errors.New("unauthorized")
	}

	if err := s.Repo.UpdateFriendRequestStatus(requestID, status); err != nil {
		return err
	}

	if status == "accepted" {
		if err := s.Repo.AddFriend(request.SenderID, request.RecipientID); err != nil {
			return err
		}
	}

	return nil
}
*/

//func (s *Service) GetFriends(userID uint) ([]models.Friend, error) {
//	return s.Repo.GetFriends(userID)
//}

/*
	func (s *Service) RemoveFriend(userID, friendID uint) error {
		requests, err := s.Repo.GetFriendRequestsForUser(userID, friendID)
		if err == nil {
			for _, req := range requests {
				s.Repo.DeleteFriendRequest(req.ID)
			}
		}

		areFriends, err := s.Repo.AreFriends(userID, friendID)
		if err != nil {
			return err
		}
		if !areFriends {
			return errors.New("не являетесь друзьями")
		}

		return s.Repo.RemoveFriend(userID, friendID)
	}
*/

/*
func (s *Service) SendFriendRequest(senderID uint, recipientPhone string) (*models.FriendRequest, error) {
	normalizedPhone, err := NormalizePhone(recipientPhone)
	if err != nil {
		return nil, errors.New("неверный формат телефона")
	}

	recipient, err := s.Repo.GetUserByPhone(normalizedPhone)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if senderID == recipient.ID {
		return nil, errors.New("нельзя отправить запрос самому себе")
	}

	existingRequests, err := s.Repo.GetFriendRequestsForUser(senderID, recipient.ID)
	if err != nil {
		return nil, err
	}

	for _, req := range existingRequests {
		if req.Status == "pending" {
			return nil, errors.New("запрос уже отправлен")
		}
		if req.Status == "rejected" {
			s.Repo.DeleteFriendRequest(req.ID)
		}
	}

	//request := &models.FriendRequest{
	//	SenderID:    senderID,
	//	RecipientID: recipient.ID,
	//	Status:      "pending",
	//}

	//if err := s.Repo.CreateFriendRequest(request); err != nil {
	//	return nil, errors.New("не удалось отправить запрос")
	//}

	return request, nil
}
*/
