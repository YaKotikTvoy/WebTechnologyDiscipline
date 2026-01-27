package service

import (
	"errors"
	"fmt"
	"io"
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

	request := &models.FriendRequest{
		SenderID:    senderID,
		RecipientID: recipient.ID,
		Status:      "pending",
	}

	if err := s.Repo.CreateFriendRequest(request); err != nil {
		return nil, errors.New("не удалось отправить запрос")
	}

	return request, nil
}

func (s *Service) GetFriendRequests(userID uint) ([]models.FriendRequest, error) {
	return s.Repo.GetFriendRequests(userID)
}

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

func (s *Service) GetFriends(userID uint) ([]models.Friend, error) {
	return s.Repo.GetFriends(userID)
}

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

	areFriends, err := s.Repo.AreFriends(userID1, userID2)
	if err != nil {
		return nil, err
	}

	if !areFriends {
		request := &models.FriendRequest{
			SenderID:    userID1,
			RecipientID: userID2,
			Status:      "pending",
		}
		if err := s.Repo.CreateFriendRequest(request); err != nil {
			return nil, errors.New("не удалось отправить запрос в друзья")
		}
		return nil, errors.New("необходимо быть друзьями для создания чата. запрос в друзья отправлен")
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

	return s.Repo.GetChatByID(chat.ID)
}

func (s *Service) CreateGroupChat(creatorID uint, name string, memberPhones []string) (*models.Chat, error) {
	chat := &models.Chat{
		Name:      name,
		Type:      "group",
		CreatedBy: creatorID,
	}

	if err := s.Repo.CreateChat(chat); err != nil {
		return nil, err
	}

	if err := s.Repo.AddChatMember(chat.ID, creatorID, true); err != nil {
		return nil, err
	}

	users, err := s.Repo.FindUsersByPhones(memberPhones)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID != creatorID {
			if err := s.Repo.AddChatMember(chat.ID, user.ID, false); err != nil {
				return nil, err
			}
		}
	}

	return s.Repo.GetChatByID(chat.ID)
}
func (s *Service) GetChat(chatID uint) (*models.Chat, error) {
	return s.Repo.GetChatByID(chatID)
}

func (s *Service) AddChatMember(chatID, adderID, userID uint) error {
	isMember, err := s.Repo.IsChatMember(chatID, adderID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("не являетесь участником чата")
	}

	members, err := s.Repo.GetChatMembers(chatID)
	if err != nil {
		return err
	}

	var isAdderAdmin bool
	for _, m := range members {
		if m.UserID == adderID {
			isAdderAdmin = m.IsAdmin
			break
		}
	}

	if !isAdderAdmin {
		return errors.New("только администратор может добавлять участников")
	}

	existingInvite, err := s.Repo.GetChatInvite(chatID, userID)
	if err == nil && existingInvite != nil && existingInvite.Status == "pending" {
		return errors.New("приглашение уже отправлено")
	}

	invite := &models.ChatInvite{
		ChatID:    chatID,
		InviterID: adderID,
		UserID:    userID,
		Status:    "pending",
	}

	if err := s.Repo.CreateChatInvite(invite); err != nil {
		return err
	}

	chat, _ := s.Repo.GetChatByID(chatID)
	inviter, _ := s.Repo.GetUserByID(adderID)

	s.hub.SendToUser(userID, models.WSMessage{
		Type: "chat_invite",
		Data: map[string]interface{}{
			"invite_id": invite.ID,
			"chat_id":   chatID,
			"chat_name": chat.Name,
			"inviter": map[string]interface{}{
				"id":       inviter.ID,
				"phone":    inviter.Phone,
				"username": inviter.Username,
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

	chat, err := s.Repo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	if chat.Type == "private" {
		members, err := s.Repo.GetChatMembers(chatID)
		if err != nil {
			return nil, err
		}

		if len(members) == 2 {
			var otherMemberID uint
			for _, member := range members {
				if member.UserID != senderID {
					otherMemberID = member.UserID
					break
				}
			}

			areFriends, err := s.Repo.AreFriends(senderID, otherMemberID)
			if err != nil {
				return nil, err
			}

			if !areFriends {
				return nil, errors.New("необходимо быть друзьями для отправки сообщений")
			}
		}
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

	return s.Repo.GetMessageWithDetails(message.ID)
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

func (s *Service) GetUnreadCount(chatID, userID uint) (int, error) {
	unreadMessages, err := s.Repo.GetUnreadMessages(chatID, userID)
	if err != nil {
		return 0, err
	}

	return len(unreadMessages), nil
}

func (s *Service) MarkMessageAsRead(messageID, userID uint) error {
	return s.Repo.MarkMessageAsRead(messageID, userID)
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
