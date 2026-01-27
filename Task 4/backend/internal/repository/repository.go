package repository

import (
	"errors"
	"time"
	"webchat/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
	Db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{Db: db}
}

func (r *Repository) CreateUser(user *models.User) error {
	return r.Db.Create(user).Error
}

func (r *Repository) GetUserByPhone(phone string) (*models.User, error) {
	var user models.User

	err := r.Db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(userID uint, updates map[string]interface{}) error {
	return r.Db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *Repository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.Db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserLastSeen(userID uint) error {
	return r.Db.Model(&models.User{}).Where("id = ?", userID).Update("last_seen_at", time.Now()).Error
}

func (r *Repository) CreateFriendRequest(request *models.FriendRequest) error {
	return r.Db.Create(request).Error
}

func (r *Repository) GetFriendRequestByID(id uint) (*models.FriendRequest, error) {
	var request models.FriendRequest
	err := r.Db.Preload("Sender").Preload("Recipient").First(&request, id).Error
	if err != nil {
		return nil, err
	}
	return &request, nil
}

func (r *Repository) GetFriendRequests(recipientID uint) ([]models.FriendRequest, error) {
	var requests []models.FriendRequest
	err := r.Db.Preload("Sender").Where("recipient_id = ? AND status = 'pending'", recipientID).Find(&requests).Error
	return requests, err
}

func (r *Repository) UpdateFriendRequestStatus(id uint, status string) error {
	return r.Db.Model(&models.FriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) AreFriends(userID, friendID uint) (bool, error) {
	var count int64
	err := r.Db.Model(&models.Friend{}).
		Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userID, friendID, friendID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) AddFriend(userID, friendID uint) error {
	friends := []models.Friend{
		{UserID: userID, FriendID: friendID},
		{UserID: friendID, FriendID: userID},
	}
	return r.Db.Create(&friends).Error
}

func (r *Repository) RemoveFriend(userID, friendID uint) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
			userID, friendID, friendID, userID).Delete(&models.Friend{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) GetFriends(userID uint) ([]models.Friend, error) {
	var friends []models.Friend
	err := r.Db.Preload("Friend").Where("user_id = ?", userID).Find(&friends).Error
	return friends, err
}

func (r *Repository) CreateChat(chat *models.Chat) error {
	return r.Db.Create(chat).Error
}

func (r *Repository) GetChatByID(id uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.Db.Preload("Members").First(&chat, id).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *Repository) GetMessageWithDetails(messageID uint) (*models.Message, error) {
	var message models.Message
	err := r.Db.Preload("Sender").Preload("Files").
		First(&message, messageID).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *Repository) GetUserChats(userID uint) ([]models.Chat, error) {
	var chats []models.Chat
	err := r.Db.Joins("JOIN chat_members ON chats.id = chat_members.chat_id").
		Where("chat_members.user_id = ?", userID).
		Preload("Members").
		Group("chats.id").
		Find(&chats).Error
	return chats, err
}

func (r *Repository) AddChatMember(chatID, userID uint, isAdmin bool) error {
	member := models.ChatMember{
		ChatID:  chatID,
		UserID:  userID,
		IsAdmin: isAdmin,
	}
	return r.Db.Create(&member).Error
}

func (r *Repository) IsChatMember(chatID, userID uint) (bool, error) {
	var count int64
	err := r.Db.Model(&models.ChatMember{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) RemoveChatMember(chatID, userID uint) error {
	return r.Db.Where("chat_id = ? AND user_id = ?", chatID, userID).Delete(&models.ChatMember{}).Error
}

func (r *Repository) GetChatMembers(chatID uint) ([]models.ChatMember, error) {
	var members []models.ChatMember
	err := r.Db.Where("chat_id = ?", chatID).Find(&members).Error
	return members, err
}

func (r *Repository) CreateMessage(message *models.Message) error {
	return r.Db.Create(message).Error
}

func (r *Repository) GetChatMessages(chatID uint, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.Db.Preload("Sender").Preload("Files").
		Where("chat_id = ? AND is_deleted = false", chatID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
}

func (r *Repository) DeleteMessage(messageID, userID uint) error {
	return r.Db.Model(&models.Message{}).
		Where("id = ? AND sender_id = ?", messageID, userID).
		Update("is_deleted", true).Error
}

func (r *Repository) AttachFileToMessage(file *models.MessageFile) error {
	return r.Db.Create(file).Error
}

func (r *Repository) CreateSession(session *models.UserSession) error {
	return r.Db.Create(session).Error
}

func (r *Repository) GetSessionByToken(token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.Db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *Repository) DeleteSession(token string) error {
	return r.Db.Where("token = ?", token).Delete(&models.UserSession{}).Error
}

func (r *Repository) DeleteAllUserSessions(userID uint) error {
	return r.Db.Where("user_id = ?", userID).Delete(&models.UserSession{}).Error
}

func (r *Repository) FindUsersByPhones(phones []string) ([]models.User, error) {
	var users []models.User
	err := r.Db.Where("phone IN ?", phones).Find(&users).Error
	return users, err
}

func (r *Repository) GetPrivateChat(userID1, userID2 uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.Db.Raw(`
		SELECT c.* FROM chats c
		JOIN chat_members cm1 ON c.id = cm1.chat_id AND cm1.user_id = ?
		JOIN chat_members cm2 ON c.id = cm2.chat_id AND cm2.user_id = ?
		WHERE c.type = 'private' AND (
			SELECT COUNT(*) FROM chat_members WHERE chat_id = c.id
		) = 2
	`, userID1, userID2).First(&chat).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &chat, err
}
func (r *Repository) FindGroupChats(search string) ([]models.Chat, error) {
	var chats []models.Chat
	query := r.Db.Where("type = ?", "group")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ?", searchPattern)
	}

	err := query.Preload("Members").Find(&chats).Error
	return chats, err
}

func (r *Repository) GetUnreadMessages(chatID, userID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.Db.Raw(`
		SELECT m.* FROM messages m
		WHERE m.chat_id = ? 
		AND m.sender_id != ?
		AND m.is_deleted = false
		AND NOT EXISTS (
			SELECT 1 FROM message_readers mr 
			WHERE mr.message_id = m.id AND mr.user_id = ?
		)
		ORDER BY m.created_at DESC
	`, chatID, userID, userID).Scan(&messages).Error
	return messages, err
}

func (r *Repository) MarkMessageAsRead(messageID, userID uint) error {
	existing := &models.MessageReader{}
	err := r.Db.Where("message_id = ? AND user_id = ?", messageID, userID).First(existing).Error

	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		reader := &models.MessageReader{
			MessageID: messageID,
			UserID:    userID,
			ReadAt:    time.Now(),
		}
		return r.Db.Create(reader).Error
	}

	return err
}

func (r *Repository) MarkChatMessagesAsRead(chatID, userID uint) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		var unreadMessages []uint
		err := tx.Raw(`
			SELECT m.id FROM messages m
			WHERE m.chat_id = ? 
			AND m.sender_id != ?
			AND m.is_deleted = false
			AND NOT EXISTS (
				SELECT 1 FROM message_readers mr 
				WHERE mr.message_id = m.id AND mr.user_id = ?
			)
		`, chatID, userID, userID).Pluck("id", &unreadMessages).Error

		if err != nil {
			return err
		}

		for _, msgID := range unreadMessages {
			reader := &models.MessageReader{
				MessageID: msgID,
				UserID:    userID,
				ReadAt:    time.Now(),
			}
			if err := tx.Create(reader).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repository) CreateRegistrationCode(code *models.RegistrationCode) error {
	return r.Db.Create(code).Error
}

func (r *Repository) GetRegistrationCode(phone, code string) (*models.RegistrationCode, error) {
	var regCode models.RegistrationCode
	err := r.Db.Where("phone = ? AND code = ?", phone, code).First(&regCode).Error
	if err != nil {
		return nil, err
	}
	return &regCode, nil
}

func (r *Repository) DeleteRegistrationCode(id uint) error {
	return r.Db.Where("id = ?", id).Delete(&models.RegistrationCode{}).Error
}

func (r *Repository) GetRegistrationCodeByPhone(phone string) (*models.RegistrationCode, error) {
	var regCode models.RegistrationCode
	err := r.Db.Where("phone = ?", phone).First(&regCode).Error
	if err != nil {
		return nil, err
	}
	return &regCode, nil
}

func (r *Repository) CreateTempPassword(temp *models.TempPassword) error {
	return r.Db.Create(temp).Error
}

func (r *Repository) GetTempPassword(phone string) (*models.TempPassword, error) {
	var temp models.TempPassword
	err := r.Db.Where("phone = ?", phone).Order("created_at DESC").First(&temp).Error
	if err != nil {
		return nil, err
	}
	return &temp, nil
}

func (r *Repository) DeleteTempPassword(id uint) error {
	return r.Db.Where("id = ?", id).Delete(&models.TempPassword{}).Error
}

func (r *Repository) CreateChatInvite(invite *models.ChatInvite) error {
	return r.Db.Create(invite).Error
}

func (r *Repository) GetTempPasswordsByPhone(phone string) ([]models.TempPassword, error) {
	var temps []models.TempPassword
	err := r.Db.Where("phone = ?", phone).Find(&temps).Error
	return temps, err
}

func (r *Repository) GetRegistrationCodesByPhone(phone string) ([]models.RegistrationCode, error) {
	var codes []models.RegistrationCode
	err := r.Db.Where("phone = ?", phone).Find(&codes).Error
	return codes, err
}

func (r *Repository) GetFriendRequestsForUser(senderID, recipientID uint) ([]models.FriendRequest, error) {
	var requests []models.FriendRequest
	err := r.Db.Where("(sender_id = ? AND recipient_id = ?) OR (sender_id = ? AND recipient_id = ?)",
		senderID, recipientID, recipientID, senderID).
		Find(&requests).Error
	return requests, err
}

func (r *Repository) DeleteFriendRequest(id uint) error {
	return r.Db.Where("id = ?", id).Delete(&models.FriendRequest{}).Error
}

func (r *Repository) UpdateChatInviteStatus(chatID, userID uint, status string) error {
	return r.Db.Model(&models.ChatInvite{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Update("status", status).Error
}
