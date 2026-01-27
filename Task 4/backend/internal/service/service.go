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

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      *repository.Repository
	config    *config.Config
	jwtSecret string
}

func NewService(repo *repository.Repository, config *config.Config, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		config:    config,
		jwtSecret: jwtSecret,
	}
}

func (s *Service) Register(phone, password string) (string, error) {
	existing, _ := s.repo.GetUserByPhone(phone)
	if existing != nil {
		return "", errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &models.User{
		Phone:        phone,
		PasswordHash: string(hashedPassword),
		Username:     phone,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return "", err
	}

	fmt.Printf("Registration code for %s: 123456\n", phone)

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return "", err
	}

	if err := s.repo.CreateSession(&models.UserSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Login(phone, password string) (string, error) {
	user, err := s.repo.GetUserByPhone(phone)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := s.repo.UpdateUserLastSeen(user.ID); err != nil {
		return "", err
	}

	token, err := s.generateJWT(user.ID)
	if err != nil {
		return "", err
	}

	if err := s.repo.CreateSession(&models.UserSession{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) Logout(token string) error {
	return s.repo.DeleteSession(token)
}

func (s *Service) LogoutAll(userID uint) error {
	return s.repo.DeleteAllUserSessions(userID)
}

func (s *Service) GetCurrentUser(token string) (*models.User, error) {
	session, err := s.repo.GetSessionByToken(token)
	if err != nil {
		return nil, errors.New("invalid session")
	}

	user, err := s.repo.GetUserByID(session.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (s *Service) SendFriendRequest(senderID uint, recipientPhone string) (*models.FriendRequest, error) {
	if recipientPhone == "" {
		return nil, errors.New("recipient phone is required")
	}

	recipient, err := s.repo.GetUserByPhone(recipientPhone)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if senderID == recipient.ID {
		return nil, errors.New("cannot send friend request to yourself")
	}

	areFriends, err := s.repo.AreFriends(senderID, recipient.ID)
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

	if err := s.repo.CreateFriendRequest(request); err != nil {
		return nil, err
	}

	return request, nil
}

func (s *Service) GetFriendRequests(userID uint) ([]models.FriendRequest, error) {
	return s.repo.GetFriendRequests(userID)
}

func (s *Service) RespondToFriendRequest(requestID uint, userID uint, status string) error {
	request, err := s.repo.GetFriendRequestByID(requestID)
	if err != nil {
		return errors.New("friend request not found")
	}

	if request.RecipientID != userID {
		return errors.New("unauthorized")
	}

	if err := s.repo.UpdateFriendRequestStatus(requestID, status); err != nil {
		return err
	}

	if status == "accepted" {
		if err := s.repo.AddFriend(request.SenderID, request.RecipientID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) GetFriends(userID uint) ([]models.Friend, error) {
	return s.repo.GetFriends(userID)
}

func (s *Service) RemoveFriend(userID, friendID uint) error {
	areFriends, err := s.repo.AreFriends(userID, friendID)
	if err != nil {
		return err
	}
	if !areFriends {
		return errors.New("not friends")
	}

	return s.repo.RemoveFriend(userID, friendID)
}

func (s *Service) SearchUserByPhone(phone string) (*models.User, error) {
	return s.repo.GetUserByPhone(phone)
}

func (s *Service) GetChats(userID uint) ([]models.Chat, error) {
	return s.repo.GetUserChats(userID)
}

func (s *Service) CreatePrivateChat(userID1, userID2 uint) (*models.Chat, error) {
	areFriends, err := s.repo.AreFriends(userID1, userID2)
	if err != nil {
		return nil, err
	}
	if !areFriends {
		return nil, errors.New("users must be friends")
	}

	existingChat, _ := s.repo.GetPrivateChat(userID1, userID2)
	if existingChat != nil {
		return existingChat, nil
	}

	chat := &models.Chat{
		Type:      "private",
		CreatedBy: userID1,
	}

	if err := s.repo.CreateChat(chat); err != nil {
		return nil, err
	}

	if err := s.repo.AddChatMember(chat.ID, userID1, false); err != nil {
		return nil, err
	}
	if err := s.repo.AddChatMember(chat.ID, userID2, false); err != nil {
		return nil, err
	}

	return s.repo.GetChatByID(chat.ID)
}

func (s *Service) CreateGroupChat(creatorID uint, name string, memberPhones []string) (*models.Chat, error) {
	chat := &models.Chat{
		Name:      name,
		Type:      "group",
		CreatedBy: creatorID,
	}

	if err := s.repo.CreateChat(chat); err != nil {
		return nil, err
	}

	if err := s.repo.AddChatMember(chat.ID, creatorID, true); err != nil {
		return nil, err
	}

	users, err := s.repo.FindUsersByPhones(memberPhones)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.ID != creatorID {
			areFriends, _ := s.repo.AreFriends(creatorID, user.ID)
			if !areFriends {
				continue
			}
			if err := s.repo.AddChatMember(chat.ID, user.ID, false); err != nil {
				return nil, err
			}
		}
	}

	return s.repo.GetChatByID(chat.ID)
}

func (s *Service) GetChat(chatID uint) (*models.Chat, error) {
	return s.repo.GetChatByID(chatID)
}

func (s *Service) AddChatMember(chatID, adderID, userID uint) error {
	isMember, err := s.repo.IsChatMember(chatID, adderID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("not a chat member")
	}

	members, err := s.repo.GetChatMembers(chatID)
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

	areFriends, err := s.repo.AreFriends(adderID, userID)
	if err != nil {
		return err
	}
	if !areFriends {
		return errors.New("can only add friends")
	}

	return s.repo.AddChatMember(chatID, userID, false)
}

func (s *Service) RemoveChatMember(chatID, removerID, userID uint) error {
	isMember, err := s.repo.IsChatMember(chatID, removerID)
	if err != nil {
		return err
	}
	if !isMember {
		return errors.New("not a chat member")
	}

	members, err := s.repo.GetChatMembers(chatID)
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

	return s.repo.RemoveChatMember(chatID, userID)
}

func (s *Service) SendMessage(chatID, senderID uint, content string, file *multipart.FileHeader) (*models.Message, error) {
	isMember, err := s.repo.IsChatMember(chatID, senderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("not a chat member")
	}

	chat, err := s.repo.GetChatByID(chatID)
	if err != nil {
		return nil, err
	}

	if chat.Type == "private" {
		var otherMemberID uint
		for _, member := range chat.Members {
			if member.ID != senderID {
				otherMemberID = member.ID
				break
			}
		}
		areFriends, err := s.repo.AreFriends(senderID, otherMemberID)
		if err != nil {
			return nil, err
		}
		if !areFriends {
			return nil, errors.New("not friends with chat member")
		}
	}

	if len(content) > 5000 {
		return nil, errors.New("message too long")
	}

	message := &models.Message{
		ChatID:   chatID,
		SenderID: senderID,
		Content:  content,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, err
	}

	if file != nil {
		if file.Size > s.config.MaxFileSize {
			return nil, errors.New("file too large")
		}

		uploadDir := s.config.UploadDir
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		filepath := filepath.Join(uploadDir, filename)

		src, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		dst, err := os.Create(filepath)
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return nil, err
		}

		messageFile := &models.MessageFile{
			MessageID: message.ID,
			Filename:  file.Filename,
			Filepath:  filename,
			Filesize:  file.Size,
			MimeType:  file.Header.Get("Content-Type"),
		}

		if err := s.repo.AttachFileToMessage(messageFile); err != nil {
			return nil, err
		}
	}

	return s.repo.GetMessageWithDetails(message.ID)
}

func (s *Service) IsChatMember(chatID, userID uint) (bool, error) {
	return s.repo.IsChatMember(chatID, userID)
}

func (s *Service) GetChatMessages(chatID, userID uint, limit int) ([]models.Message, error) {
	isMember, err := s.repo.IsChatMember(chatID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, errors.New("not a chat member")
	}

	return s.repo.GetChatMessages(chatID, limit)
}

func (s *Service) DeleteMessage(messageID, userID uint) error {
	return s.repo.DeleteMessage(messageID, userID)
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

	return s.repo.UpdateUser(userID, updates)
}
