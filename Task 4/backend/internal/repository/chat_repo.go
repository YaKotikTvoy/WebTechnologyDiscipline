package repository

import (
	"database/sql"
	"fmt"
	"time"

	"webchat/internal/models"
	"webchat/pkg/database"
)

type ChatRepository interface {
	CreateChat(chat *models.Chat) error
	GetChatByID(id string) (*models.Chat, error)
	GetPublicChats() ([]models.Chat, error)
	GetUserChats(userID string) ([]models.Chat, error)
	UpdateChat(id string, update *models.UpdateChatRequest) error
	AddChatMember(chatID, userID string) error
	RemoveChatMember(chatID, userID string) error
	GetChatMembers(chatID string) ([]models.ChatMember, error)
	SetChatRole(chatID, userID string, role *models.ChatRole) error
	GetChatRole(chatID, userID string) (*models.ChatRole, error)
	CreateInvitation(invite *models.Invitation) error
	GetInvitationByCode(code string) (*models.Invitation, error)
	UseInvitation(code, userID string) error
	CheckUserInChat(chatID, userID string) (bool, error)
	UserHasPermission(chatID, userID string, permission string) (bool, error)
	GetPublicChat(chatID string) (*models.Chat, error)
	CheckPublicChat(chatID string) (bool, error)
}

type chatRepository struct {
	db *database.DB
}

func (r *chatRepository) GetPublicChat(chatID string) (*models.Chat, error) {
	query := `
        SELECT c.id, c.name, c.description, c.avatar_url, c.is_public,
               c.creator_id, c.only_admin_invite, c.created_at, c.updated_at,
               COUNT(DISTINCT cm.user_id) as member_count
        FROM chats c
        LEFT JOIN chat_members cm ON c.id = cm.chat_id
        WHERE c.id = $1 AND c.is_public = TRUE
        GROUP BY c.id
    `

	var chat models.Chat
	err := r.db.QueryRow(query, chatID).Scan(
		&chat.ID,
		&chat.Name,
		&chat.Description,
		&chat.AvatarURL,
		&chat.IsPublic,
		&chat.CreatorID,
		&chat.OnlyAdminInvite,
		&chat.CreatedAt,
		&chat.UpdatedAt,
		&chat.MemberCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepository) CheckPublicChat(chatID string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM chats 
            WHERE id = $1 AND is_public = TRUE
        )
    `

	var exists bool
	err := r.db.QueryRow(query, chatID).Scan(&exists)
	return exists, err
}

func NewChatRepository(db *database.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChat(chat *models.Chat) error {
	query := `
		INSERT INTO chats (name, description, avatar_url, is_public, creator_id, only_admin_invite)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(query,
		chat.Name,
		chat.Description,
		chat.AvatarURL,
		chat.IsPublic,
		chat.CreatorID,
		chat.OnlyAdminInvite,
	).Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt)

	if err != nil {
		return err
	}

	err = r.AddChatMember(chat.ID, chat.CreatorID)
	if err != nil {
		return err
	}

	ownerRole := &models.ChatRole{
		UserID:            chat.CreatorID,
		RoleName:          "owner",
		CanDeleteMessages: true,
		CanRemoveUsers:    true,
		CanAssignRoles:    true,
	}

	return r.SetChatRole(chat.ID, chat.CreatorID, ownerRole)
}

func (r *chatRepository) GetChatByID(id string) (*models.Chat, error) {
	query := `
		SELECT c.id, c.name, c.description, c.avatar_url, c.is_public, 
		       c.creator_id, c.only_admin_invite, c.created_at, c.updated_at,
		       COUNT(DISTINCT cm.user_id) as member_count
		FROM chats c
		LEFT JOIN chat_members cm ON c.id = cm.chat_id
		WHERE c.id = $1
		GROUP BY c.id
	`

	var chat models.Chat
	err := r.db.QueryRow(query, id).Scan(
		&chat.ID,
		&chat.Name,
		&chat.Description,
		&chat.AvatarURL,
		&chat.IsPublic,
		&chat.CreatorID,
		&chat.OnlyAdminInvite,
		&chat.CreatedAt,
		&chat.UpdatedAt,
		&chat.MemberCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepository) GetPublicChats() ([]models.Chat, error) {
	query := `
		SELECT c.id, c.name, c.description, c.avatar_url, c.is_public,
		       c.creator_id, c.only_admin_invite, c.created_at, c.updated_at,
		       COUNT(DISTINCT cm.user_id) as member_count
		FROM chats c
		LEFT JOIN chat_members cm ON c.id = cm.chat_id
		WHERE c.is_public = TRUE
		GROUP BY c.id
		ORDER BY c.updated_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		err := rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.Description,
			&chat.AvatarURL,
			&chat.IsPublic,
			&chat.CreatorID,
			&chat.OnlyAdminInvite,
			&chat.CreatedAt,
			&chat.UpdatedAt,
			&chat.MemberCount,
		)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	return chats, nil
}

func (r *chatRepository) GetUserChats(userID string) ([]models.Chat, error) {
	query := `
		SELECT c.id, c.name, c.description, c.avatar_url, c.is_public,
		       c.creator_id, c.only_admin_invite, c.created_at, c.updated_at,
		       COUNT(DISTINCT cm2.user_id) as member_count,
		       m.content as last_message,
		       (SELECT COUNT(*) FROM messages m2 
		        WHERE m2.chat_id = c.id 
		        AND m2.created_at > COALESCE(
		            (SELECT last_seen FROM user_chat_activity 
		             WHERE chat_id = c.id AND user_id = $1),
		            '1970-01-01'::timestamp
		        )) as unread_count
		FROM chats c
		INNER JOIN chat_members cm ON c.id = cm.chat_id AND cm.user_id = $1
		LEFT JOIN chat_members cm2 ON c.id = cm2.chat_id
		LEFT JOIN LATERAL (
			SELECT content FROM messages 
			WHERE chat_id = c.id 
			ORDER BY created_at DESC 
			LIMIT 1
		) m ON TRUE
		GROUP BY c.id, m.content
		ORDER BY c.updated_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		var lastMessage sql.NullString
		err := rows.Scan(
			&chat.ID,
			&chat.Name,
			&chat.Description,
			&chat.AvatarURL,
			&chat.IsPublic,
			&chat.CreatorID,
			&chat.OnlyAdminInvite,
			&chat.CreatedAt,
			&chat.UpdatedAt,
			&chat.MemberCount,
			&lastMessage,
			&chat.UnreadCount,
		)
		if err != nil {
			return nil, err
		}

		if lastMessage.Valid {
			chat.LastMessage = &lastMessage.String
		}

		chats = append(chats, chat)
	}

	return chats, nil
}

func (r *chatRepository) UpdateChat(id string, update *models.UpdateChatRequest) error {
	query := `
		UPDATE chats
		SET name = COALESCE($1, name),
			description = COALESCE($2, description),
			avatar_url = COALESCE($3, avatar_url),
			only_admin_invite = COALESCE($4, only_admin_invite),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(query,
		update.Name,
		update.Description,
		update.AvatarURL,
		update.OnlyAdminInvite,
		id,
	).Scan(&updatedAt)

	return err
}

func (r *chatRepository) AddChatMember(chatID, userID string) error {
	query := `
		INSERT INTO chat_members (chat_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`

	_, err := r.db.Exec(query, chatID, userID)
	if err != nil {
		return err
	}

	memberRole := &models.ChatRole{
		UserID:            userID,
		RoleName:          "member",
		CanDeleteMessages: false,
		CanRemoveUsers:    false,
		CanAssignRoles:    false,
	}

	return r.SetChatRole(chatID, userID, memberRole)
}

func (r *chatRepository) RemoveChatMember(chatID, userID string) error {
	query := `
		DELETE FROM chat_members
		WHERE chat_id = $1 AND user_id = $2
	`

	_, err := r.db.Exec(query, chatID, userID)
	return err
}

func (r *chatRepository) GetChatMembers(chatID string) ([]models.ChatMember, error) {
	query := `
		SELECT u.id, u.username, u.avatar_url, cr.role_name, cm.joined_at
		FROM chat_members cm
		JOIN users u ON cm.user_id = u.id AND u.deleted_at IS NULL
		LEFT JOIN chat_roles cr ON cm.chat_id = cr.chat_id AND cm.user_id = cr.user_id
		WHERE cm.chat_id = $1
		ORDER BY 
			CASE cr.role_name 
				WHEN 'owner' THEN 1
				WHEN 'admin' THEN 2
				WHEN 'moderator' THEN 3
				ELSE 4
			END,
			cm.joined_at
	`

	rows, err := r.db.Query(query, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.ChatMember
	for rows.Next() {
		var member models.ChatMember
		err := rows.Scan(
			&member.UserID,
			&member.Username,
			&member.AvatarURL,
			&member.RoleName,
			&member.JoinedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	return members, nil
}

func (r *chatRepository) SetChatRole(chatID, userID string, role *models.ChatRole) error {
	query := `
		INSERT INTO chat_roles (chat_id, user_id, role_name, can_delete_messages, can_remove_users, can_assign_roles)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (chat_id, user_id) DO UPDATE SET
			role_name = EXCLUDED.role_name,
			can_delete_messages = EXCLUDED.can_delete_messages,
			can_remove_users = EXCLUDED.can_remove_users,
			can_assign_roles = EXCLUDED.can_assign_roles,
			granted_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Exec(query,
		chatID,
		userID,
		role.RoleName,
		role.CanDeleteMessages,
		role.CanRemoveUsers,
		role.CanAssignRoles,
	)

	return err
}

func (r *chatRepository) GetChatRole(chatID, userID string) (*models.ChatRole, error) {
	query := `
		SELECT role_name, can_delete_messages, can_remove_users, can_assign_roles
		FROM chat_roles
		WHERE chat_id = $1 AND user_id = $2
	`

	var role models.ChatRole
	err := r.db.QueryRow(query, chatID, userID).Scan(
		&role.RoleName,
		&role.CanDeleteMessages,
		&role.CanRemoveUsers,
		&role.CanAssignRoles,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	role.UserID = userID
	return &role, nil
}

func (r *chatRepository) CreateInvitation(invite *models.Invitation) error {
	query := `
		INSERT INTO invitations (chat_id, inviter_id, invite_code, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(query,
		invite.ChatID,
		invite.InviterID,
		invite.InviteCode,
		invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt)

	return err
}

func (r *chatRepository) GetInvitationByCode(code string) (*models.Invitation, error) {
	query := `
		SELECT id, chat_id, inviter_id, invite_code, is_used, used_by, expires_at, created_at
		FROM invitations
		WHERE invite_code = $1 AND expires_at > CURRENT_TIMESTAMP
	`

	var invite models.Invitation
	var inviterID, usedBy sql.NullString

	err := r.db.QueryRow(query, code).Scan(
		&invite.ID,
		&invite.ChatID,
		&inviterID,
		&invite.InviteCode,
		&invite.IsUsed,
		&usedBy,
		&invite.ExpiresAt,
		&invite.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if inviterID.Valid {
		invite.InviterID = &inviterID.String
	}

	if usedBy.Valid {
		invite.UsedBy = &usedBy.String
	}

	return &invite, nil
}

func (r *chatRepository) UseInvitation(code, userID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var chatID string
	err = tx.QueryRow(`
		SELECT chat_id FROM invitations 
		WHERE invite_code = $1 AND is_used = FALSE AND expires_at > CURRENT_TIMESTAMP
		FOR UPDATE
	`, code).Scan(&chatID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO chat_members (chat_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (chat_id, user_id) DO NOTHING
	`, chatID, userID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		UPDATE invitations
		SET is_used = TRUE, used_by = $2
		WHERE invite_code = $1
	`, code, userID)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *chatRepository) CheckUserInChat(chatID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM chat_members 
			WHERE chat_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(query, chatID, userID).Scan(&exists)
	return exists, err
}

func (r *chatRepository) UserHasPermission(chatID, userID string, permission string) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1 FROM chat_roles
			WHERE chat_id = $1 AND user_id = $2 AND %s = TRUE
		)
	`, permission)

	var hasPermission bool
	err := r.db.QueryRow(query, chatID, userID).Scan(&hasPermission)
	return hasPermission, err
}
