package repository

import (
	"errors"
	"log"
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
	err := r.Db.Preload("Sender").Preload("Files").Preload("Readers").
		First(&message, messageID).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *Repository) GetUserChats(userID uint) ([]models.Chat, error) {
	var chats []models.Chat

	err := r.Db.Raw(`
		SELECT c.* FROM chats c
		JOIN chat_members cm ON c.id = cm.chat_id
		WHERE cm.user_id = ?
		GROUP BY c.id
		ORDER BY c.updated_at DESC
	`, userID).Scan(&chats).Error

	if err != nil {
		return nil, err
	}

	for i := range chats {
		err := r.Db.Raw(`
			SELECT u.* FROM users u
			JOIN chat_members cm ON u.id = cm.user_id
			WHERE cm.chat_id = ?
		`, chats[i].ID).Scan(&chats[i].Members).Error

		if err != nil {
			continue
		}

		lastMessage, _ := r.GetLastChatMessage(chats[i].ID)
		chats[i].LastMessage = lastMessage

		unreadCount, _ := r.GetUnreadCount(chats[i].ID, userID)
		chats[i].UnreadCount = int(unreadCount)
	}

	return chats, nil
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
	err := r.Db.Preload("Sender").Preload("Files").Preload("Readers").
		Where("chat_id = ? AND is_deleted = false", chatID).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	return messages, err
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
		WHERE c.type = 'private' 
		AND c.id IN (
			SELECT cm1.chat_id 
			FROM chat_members cm1
			WHERE cm1.user_id = ?
			INTERSECT
			SELECT cm2.chat_id 
			FROM chat_members cm2
			WHERE cm2.user_id = ?
		)
		AND (
			SELECT COUNT(*) 
			FROM chat_members 
			WHERE chat_id = c.id
		) = 2
		LIMIT 1
	`, userID1, userID2).Scan(&chat).Error

	if errors.Is(err, gorm.ErrRecordNotFound) || chat.ID == 0 {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &chat, nil
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

func (r *Repository) GetUnreadCount(chatID, userID uint) (int64, error) {
	var count int64

	err := r.Db.Raw(`
        SELECT COUNT(*) FROM messages m
        WHERE m.chat_id = ? 
        AND m.sender_id != ?
        AND m.is_deleted = false
        AND NOT EXISTS (
            SELECT 1 FROM message_readers mr 
            WHERE mr.message_id = m.id AND mr.user_id = ?
        )
    `, chatID, userID, userID).Scan(&count).Error

	return count, err
}

func (r *Repository) MarkMessageAsRead(messageID, userID uint) error {
	var existing models.MessageReader
	err := r.Db.Where("message_id = ? AND user_id = ?", messageID, userID).First(&existing).Error

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

func (r *Repository) UpdateChatInviteStatus(inviteID uint, status string) error {
	return r.Db.Model(&models.ChatInvite{}).
		Where("id = ?", inviteID).
		Update("status", status).Error
}

func (r *Repository) GetChatInvitesByUserID(userID uint) ([]models.ChatInvite, error) {
	var invites []models.ChatInvite
	err := r.Db.Preload("Chat").Preload("Inviter").
		Where("user_id = ?", userID).
		Find(&invites).Error
	return invites, err
}

func (r *Repository) GetChatInviteByID(id uint) (*models.ChatInvite, error) {
	var invite models.ChatInvite
	err := r.Db.Preload("Chat").Preload("Inviter").Preload("User").
		First(&invite, id).Error
	return &invite, err
}

func (r *Repository) GetChatInvite(chatID, userID uint) (*models.ChatInvite, error) {
	var invite models.ChatInvite
	err := r.Db.Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&invite).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &invite, err
}

func (r *Repository) CreateChatJoinRequest(request *models.ChatJoinRequest) error {
	return r.Db.Create(request).Error
}

func (r *Repository) GetChatJoinRequest(id uint) (*models.ChatJoinRequest, error) {
	var request models.ChatJoinRequest
	err := r.Db.Preload("Chat").Preload("User").First(&request, id).Error
	return &request, err
}

func (r *Repository) GetChatJoinRequests(chatID uint) ([]models.ChatJoinRequest, error) {
	var requests []models.ChatJoinRequest
	err := r.Db.Preload("User").Where("chat_id = ? AND status = 'pending'", chatID).Find(&requests).Error
	return requests, err
}

func (r *Repository) UpdateChatJoinRequestStatus(id uint, status string) error {
	return r.Db.Model(&models.ChatJoinRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) GetUserChatJoinRequests(userID uint) ([]models.ChatJoinRequest, error) {
	var requests []models.ChatJoinRequest
	err := r.Db.Preload("Chat").Where("user_id = ?", userID).Find(&requests).Error
	return requests, err
}

func (r *Repository) SearchPublicChats(search string) ([]models.Chat, error) {
	var chats []models.Chat
	query := r.Db.Where("type = ? AND is_searchable = true", "group")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ?", searchPattern)
	}

	err := query.Preload("Members").Find(&chats).Error
	return chats, err
}

func (r *Repository) UpdateChatVisibility(chatID uint, isSearchable bool) error {
	return r.Db.Model(&models.Chat{}).Where("id = ?", chatID).Update("is_searchable", isSearchable).Error
}

func (r *Repository) GetChatAdmin(chatID uint) (uint, error) {
	var member models.ChatMember
	err := r.Db.Where("chat_id = ? AND is_admin = true", chatID).First(&member).Error
	if err != nil {
		return 0, err
	}
	return member.UserID, nil
}

func (r *Repository) DeleteChatJoinRequest(id uint) error {
	return r.Db.Where("id = ?", id).Delete(&models.ChatJoinRequest{}).Error
}

func (r *Repository) GetChatJoinRequestByUserAndChat(userID, chatID uint) (*models.ChatJoinRequest, error) {
	var request models.ChatJoinRequest
	err := r.Db.Where("user_id = ? AND chat_id = ?", userID, chatID).First(&request).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &request, err
}

func (r *Repository) UpdateMessage(messageID uint, content string) error {
	return r.Db.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"content":    content,
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) GetMessageByID(messageID uint) (*models.Message, error) {
	var message models.Message
	err := r.Db.First(&message, messageID).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *Repository) MarkMessageAsDeleted(messageID uint) error {
	return r.Db.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_deleted": true,
			"content":    "[Сообщение удалено]",
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) UpdateMessageContent(messageID uint, newContent string) error {
	return r.Db.Model(&models.Message{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"content":    newContent,
			"is_edited":  true,
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) GetChatInfo(chatID uint) (*models.Chat, error) {
	var chat models.Chat
	err := r.Db.First(&chat, chatID).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *Repository) IsChatAdmin(chatID, userID uint) (bool, error) {
	var count int64
	err := r.Db.Model(&models.ChatMember{}).
		Where("chat_id = ? AND user_id = ? AND is_admin = true", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *Repository) DeleteChat(chatID uint) error {
	return r.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("chat_id = ?", chatID).Delete(&models.Message{}).Error; err != nil {
			return err
		}

		if err := tx.Where("message_id IN (SELECT id FROM messages WHERE chat_id = ?)", chatID).
			Delete(&models.MessageFile{}).Error; err != nil {
			return err
		}

		if err := tx.Where("message_id IN (SELECT id FROM messages WHERE chat_id = ?)", chatID).
			Delete(&models.MessageReader{}).Error; err != nil {
			return err
		}

		if err := tx.Where("chat_id = ?", chatID).Delete(&models.ChatMember{}).Error; err != nil {
			return err
		}

		if err := tx.Where("chat_id = ?", chatID).Delete(&models.ChatJoinRequest{}).Error; err != nil {
			return err
		}

		if err := tx.Where("chat_id = ?", chatID).Delete(&models.ChatInvite{}).Error; err != nil {
			return err
		}

		return tx.Delete(&models.Chat{}, chatID).Error
	})
}

func (r *Repository) GetChatCreator(chatID uint) (uint, error) {
	var chat models.Chat
	err := r.Db.Select("created_by").First(&chat, chatID).Error
	if err != nil {
		return 0, err
	}
	return chat.CreatedBy, nil
}

func (r *Repository) GetLastChatMessage(chatID uint) (*models.Message, error) {
	log.Printf("Получаем последнее сообщение для чата %d", chatID)

	var message models.Message
	err := r.Db.Preload("Sender").
		Where("chat_id = ? AND is_deleted = false", chatID).
		Order("created_at DESC").
		First(&message).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Сообщений не найдено для чата %d", chatID)
		return nil, nil
	}

	if err != nil {
		log.Printf("Ошибка при получении последнего сообщения для чата %d: %v", chatID, err)
		return nil, err
	}

	log.Printf("Найдено последнее сообщение для чата %d: ID=%d", chatID, message.ID)
	return &message, nil
}
