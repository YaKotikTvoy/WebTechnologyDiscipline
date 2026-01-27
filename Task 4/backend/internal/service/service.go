package service

import (
	"crypto/rand"
	"encoding/hex"
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

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo      *repository.Repository
	config    *config.Config
	jwtSecret string
}

func NewService(repo *repository.Repository, config *config.Config, jwtSecret string) *Service {
	return &Service{
		Repo:      repo,
		config:    config,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(phone, password string) (string, error) {
	normalizedPhone, err := NormalizePhone(phone)
	if err != nil {
		return "", err
	}

	existing, _ := s.Repo.GetUserByPhone(normalizedPhone)
	if existing != nil {
		return "", errors.New("пользователь уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &models.User{
		Phone:        normalizedPhone,
		PasswordHash: string(hashedPassword),
		Username:     normalizedPhone,
	}

	if err := s.Repo.CreateUser(user); err != nil {
		return "", err
	}

	codeBytes := make([]byte, 3)
	rand.Read(codeBytes)
	code := hex.EncodeToString(codeBytes)[:6]

	fmt.Printf("Registration code for %s: %s\n", normalizedPhone, code)

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
	if recipientPhone == "" {
		return nil, errors.New("recipient phone is required")
	}

	normalizedPhone, err := NormalizePhone(recipientPhone)
	if err != nil {
		return nil, err
	}

	recipient, err := s.Repo.GetUserByPhone(normalizedPhone)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if senderID == recipient.ID {
		return nil, errors.New("cannot send friend request to yourself")
	}

	areFriends, err := s.Repo.AreFriends(senderID, recipient.ID)
	if err != nil {
		return nil, err
	}
	if areFriends {
		return nil, errors.New("already friends")
	}

	request := &models.FriendRequest{
		SenderID:    senderID,
		RecipientID: recipient.ID,
		Status:      "pending",
	}

	if err := s.Repo.CreateFriendRequest(request); err != nil {
		return nil, err
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
	areFriends, err := s.Repo.AreFriends(userID, friendID)
	if err != nil {
		return err
	}
	if !areFriends {
		return errors.New("not friends")
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
		return errors.New("not a chat member")
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
		return errors.New("only admin can add members")
	}

	return s.Repo.AddChatMember(chatID, userID, false)
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
		return nil, errors.New("not a chat member")
	}

	if len(content) > 5000 {
		return nil, errors.New("message too long")
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
