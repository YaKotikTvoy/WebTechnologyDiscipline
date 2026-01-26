package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"webchat/internal/models"
	"webchat/pkg/database"
)

type MessageRepository interface {
	CreateMessage(msg *models.Message) error
	GetMessageByID(id string) (*models.Message, error)
	GetMessages(chatID *string, userID string, page, pageSize int) ([]models.Message, int, error)
	UpdateMessage(id, content string) error
	DeleteMessageForUser(messageID, userID string) error
	DeleteMessageForAll(messageID string) error
	AddFileToMessage(messageID string, file *models.File) error
	GetMessageFiles(messageID string) ([]models.File, error)
	CheckBlocked(senderID, recipientID string) (bool, error)
	GetLastSeen(chatID, userID string) (*time.Time, error)
	UpdateLastSeen(chatID, userID string) error
	GetDirectMessages(userID, contactID string, page, pageSize int) ([]models.Message, int, error)
}

type messageRepository struct {
	db *database.DB
}

func NewMessageRepository(db *database.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) CreateMessage(msg *models.Message) error {
	query := `
		INSERT INTO messages (chat_id, sender_id, recipient_id, content)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query,
		msg.ChatID,
		msg.SenderID,
		msg.RecipientID,
		msg.Content,
	).Scan(&msg.ID, &msg.CreatedAt, &msg.UpdatedAt)

	return err
}

func (r *messageRepository) GetMessageByID(id string) (*models.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, recipient_id, content, 
		       is_edited, is_deleted, created_at, updated_at
		FROM messages
		WHERE id = $1 AND is_deleted = FALSE
	`

	var msg models.Message
	err := r.db.QueryRow(query, id).Scan(
		&msg.ID,
		&msg.ChatID,
		&msg.SenderID,
		&msg.RecipientID,
		&msg.Content,
		&msg.IsEdited,
		&msg.IsDeleted,
		&msg.CreatedAt,
		&msg.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &msg, nil
}

func (r *messageRepository) GetMessages(chatID *string, userID string, page, pageSize int) ([]models.Message, int, error) {
	offset := (page - 1) * pageSize

	var total int
	countQuery := `
        SELECT COUNT(*) FROM messages m
        WHERE m.is_deleted = FALSE
    `

	args := []interface{}{}
	argIndex := 1

	if userID != "" {
		countQuery += ` AND NOT EXISTS(
            SELECT 1 FROM blacklist bl 
            WHERE (m.sender_id = bl.blocked_id AND bl.blocker_id = $1)
               OR (bl.blocker_id = m.sender_id AND bl.blocked_id = $1)
        )`
		args = append(args, userID)
		argIndex++
	}

	if chatID != nil {
		countQuery += fmt.Sprintf(" AND m.chat_id = $%d", argIndex)
		args = append(args, *chatID)
	} else if userID != "" {
		countQuery += fmt.Sprintf(" AND m.recipient_id = $%d", argIndex)
		args = append(args, userID)
	}

	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
        SELECT m.id, m.chat_id, m.sender_id, m.recipient_id, m.content,
               m.is_edited, m.is_deleted, m.deleted_by_sender, m.deleted_by_recipient,
               m.created_at, m.updated_at,
               u.username as sender_username, u.avatar_url as sender_avatar
        FROM messages m
        LEFT JOIN users u ON m.sender_id = u.id AND u.deleted_at IS NULL
        WHERE m.is_deleted = FALSE
    `

	queryArgs := []interface{}{}
	queryArgIndex := 1

	if userID != "" {
		query += fmt.Sprintf(` AND NOT EXISTS(
            SELECT 1 FROM blacklist bl 
            WHERE (m.sender_id = bl.blocked_id AND bl.blocker_id = $%d)
               OR (bl.blocker_id = m.sender_id AND bl.blocked_id = $%d)
        )`, queryArgIndex, queryArgIndex)
		queryArgs = append(queryArgs, userID)
		queryArgIndex++
	}

	if chatID != nil {
		query += fmt.Sprintf(" AND m.chat_id = $%d", queryArgIndex)
		queryArgs = append(queryArgs, *chatID)
		queryArgIndex++
	} else if userID != "" {
		query += fmt.Sprintf(" AND m.recipient_id = $%d", queryArgIndex)
		queryArgs = append(queryArgs, userID)
		queryArgIndex++
	}

	query += fmt.Sprintf(`
        ORDER BY m.created_at DESC
        LIMIT $%d OFFSET $%d
    `, queryArgIndex, queryArgIndex+1)

	queryArgs = append(queryArgs, pageSize, offset)

	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var senderUsername, senderAvatar sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.IsEdited,
			&msg.IsDeleted,
			&msg.DeletedBySender,
			&msg.DeletedByRecipient,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&senderUsername,
			&senderAvatar,
		)

		if err != nil {
			return nil, 0, err
		}

		if senderUsername.Valid {
			msg.SenderUsername = &senderUsername.String
		}

		if senderAvatar.Valid {
			msg.SenderAvatar = &senderAvatar.String
		}

		files, err := r.GetMessageFiles(msg.ID)
		if err == nil {
			msg.Files = files
		}

		messages = append(messages, msg)
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, nil
}

func (r *messageRepository) UpdateMessage(id, content string) error {
	query := `
		UPDATE messages
		SET content = $1, is_edited = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(query, content, id).Scan(&updatedAt)
	return err
}

func (r *messageRepository) DeleteMessageForUser(messageID, userID string) error {
	query := `
		UPDATE messages
		SET 
			deleted_by_sender = CASE 
				WHEN sender_id = $1 THEN TRUE 
				ELSE deleted_by_sender 
			END,
			deleted_by_recipient = CASE 
				WHEN recipient_id = $1 THEN TRUE 
				ELSE deleted_by_recipient 
			END,
			is_deleted = (deleted_by_sender OR deleted_by_recipient)
		WHERE id = $2
		RETURNING is_deleted
	`

	var isDeleted bool
	err := r.db.QueryRow(query, userID, messageID).Scan(&isDeleted)
	return err
}

func (r *messageRepository) DeleteMessageForAll(messageID string) error {
	query := `
		UPDATE messages
		SET is_deleted = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err := r.db.Exec(query, messageID)
	return err
}

func (r *messageRepository) AddFileToMessage(messageID string, file *models.File) error {
	query := `
		INSERT INTO message_files (message_id, file_url, file_name, file_size, mime_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, uploaded_at
	`

	err := r.db.QueryRow(query,
		messageID,
		file.URL,
		file.Name,
		file.Size,
		file.MimeType,
	).Scan(&file.ID, &file.UploadedAt)

	return err
}

func (r *messageRepository) GetMessageFiles(messageID string) ([]models.File, error) {
	query := `
		SELECT id, file_url, file_name, file_size, mime_type
		FROM message_files
		WHERE message_id = $1
		ORDER BY uploaded_at
	`

	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(
			&file.ID,
			&file.URL,
			&file.Name,
			&file.Size,
			&file.MimeType,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

func (r *messageRepository) CheckBlocked(senderID, recipientID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM blacklist
			WHERE (blocker_id = $1 AND blocked_id = $2)
			   OR (blocker_id = $2 AND blocked_id = $1)
		)
	`

	var blocked bool
	err := r.db.QueryRow(query, senderID, recipientID).Scan(&blocked)
	return blocked, err
}

func (r *messageRepository) GetLastSeen(chatID, userID string) (*time.Time, error) {
	query := `
		SELECT MAX(created_at) FROM messages
		WHERE chat_id = $1 AND sender_id = $2
	`

	var lastSeen sql.NullTime
	err := r.db.QueryRow(query, chatID, userID).Scan(&lastSeen)

	if err != nil {
		return nil, err
	}

	if lastSeen.Valid {
		return &lastSeen.Time, nil
	}

	return nil, nil
}

func (r *messageRepository) UpdateLastSeen(chatID, userID string) error {
	return nil
}

func (r *messageRepository) GetDirectMessages(userID, contactID string, page, pageSize int) ([]models.Message, int, error) {
	offset := (page - 1) * pageSize

	blocked, err := r.CheckBlocked(userID, contactID)
	if err != nil {
		return nil, 0, err
	}
	if blocked {
		return nil, 0, errors.New("пользователи заблокированы")
	}

	countQuery := `
        SELECT COUNT(*) FROM messages m
        WHERE m.is_deleted = FALSE 
        AND ((m.sender_id = $1 AND m.recipient_id = $2) 
             OR (m.sender_id = $2 AND m.recipient_id = $1))
    `

	var total int
	err = r.db.QueryRow(countQuery, userID, contactID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
        SELECT m.id, m.chat_id, m.sender_id, m.recipient_id, m.content,
               m.is_edited, m.is_deleted, m.deleted_by_sender, m.deleted_by_recipient,
               m.created_at, m.updated_at,
               u.username as sender_username, u.avatar_url as sender_avatar
        FROM messages m
        LEFT JOIN users u ON m.sender_id = u.id AND u.deleted_at IS NULL
        WHERE m.is_deleted = FALSE 
        AND ((m.sender_id = $1 AND m.recipient_id = $2) 
             OR (m.sender_id = $2 AND m.recipient_id = $1))
        ORDER BY m.created_at ASC
        LIMIT $3 OFFSET $4
    `

	rows, err := r.db.Query(query, userID, contactID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		var senderUsername, senderAvatar sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.ChatID,
			&msg.SenderID,
			&msg.RecipientID,
			&msg.Content,
			&msg.IsEdited,
			&msg.IsDeleted,
			&msg.DeletedBySender,
			&msg.DeletedByRecipient,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&senderUsername,
			&senderAvatar,
		)

		if err != nil {
			return nil, 0, err
		}

		if senderUsername.Valid {
			msg.SenderUsername = &senderUsername.String
		}

		if senderAvatar.Valid {
			msg.SenderAvatar = &senderAvatar.String
		}

		files, err := r.GetMessageFiles(msg.ID)
		if err == nil {
			msg.Files = files
		}

		messages = append(messages, msg)
	}

	return messages, total, nil
}
