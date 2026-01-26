package repository

import (
	"database/sql"
	"time"
	"webchat/internal/models"
	"webchat/pkg/database"
)

type UserRepository interface {
	GetDB() *sql.DB
	CreateUser(user *models.User, passwordHash string) error
	GetUserByPhone(phone string) (*models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(id string, update *models.UpdateProfileRequest) error
	DeleteUser(id string) error
	AddContact(userID, contactID string, alias *string) error
	RemoveContact(userID, contactID string) error
	GetContacts(userID string) ([]models.User, error)
	AddToBlacklist(blockerID, blockedID string) error
	RemoveFromBlacklist(blockerID, blockedID string) error
	GetBlacklist(userID string) ([]models.User, error)
	CreateDeletionCode(userID, email, code string) error
	VerifyDeletionCode(userID, email, code string) (bool, error)
}

type userRepository struct {
	db *database.DB
}

func (r *userRepository) GetDB() *sql.DB {
	return r.db.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User, passwordHash string) error {
	query := `
		INSERT INTO users (phone, username, password_hash, avatar_url, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return r.db.QueryRow(query,
		user.Phone,
		user.Username,
		passwordHash,
		user.AvatarURL,
		user.Role,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *userRepository) GetUserByPhone(phone string) (*models.User, error) {
	query := `
		SELECT id, phone, username, avatar_url, role, email, is_active, created_at, updated_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRow(query, phone).Scan(
		&user.ID,
		&user.Phone,
		&user.Username,
		&user.AvatarURL,
		&user.Role,
		&user.Email,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetUserByID(id string) (*models.User, error) {
	query := `
		SELECT id, phone, username, avatar_url, role, email, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Phone,
		&user.Username,
		&user.AvatarURL,
		&user.Role,
		&user.Email,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(id string, update *models.UpdateProfileRequest) error {
	query := `
		UPDATE users
		SET username = COALESCE($1, username),
			avatar_url = COALESCE($2, avatar_url),
			email = COALESCE($3, email),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND deleted_at IS NULL
		RETURNING updated_at
	`

	var updatedAt time.Time
	err := r.db.QueryRow(query,
		update.Username,
		update.AvatarURL,
		update.Email,
		id,
	).Scan(&updatedAt)

	return err
}

func (r *userRepository) DeleteUser(id string) error {
	query := `
		UPDATE users
		SET deleted_at = CURRENT_TIMESTAMP,
			is_active = FALSE
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id)
	return err
}

func (r *userRepository) AddContact(userID, contactID string, alias *string) error {
	query := `
		INSERT INTO contacts (user_id, contact_id, alias)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, contact_id) DO UPDATE
		SET alias = EXCLUDED.alias
	`

	_, err := r.db.Exec(query, userID, contactID, alias)
	return err
}

func (r *userRepository) RemoveContact(userID, contactID string) error {
	query := `
		DELETE FROM contacts
		WHERE user_id = $1 AND contact_id = $2
	`

	_, err := r.db.Exec(query, userID, contactID)
	return err
}

func (r *userRepository) GetContacts(userID string) ([]models.User, error) {
	query := `
		SELECT u.id, u.phone, u.username, u.avatar_url, u.role, u.created_at
		FROM contacts c
		JOIN users u ON c.contact_id = u.id
		WHERE c.user_id = $1 AND u.deleted_at IS NULL
		ORDER BY u.username NULLS LAST, u.phone
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Phone,
			&user.Username,
			&user.AvatarURL,
			&user.Role,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, user)
	}

	return contacts, nil
}

func (r *userRepository) AddToBlacklist(blockerID, blockedID string) error {
	query := `
		INSERT INTO blacklist (blocker_id, blocked_id)
		VALUES ($1, $2)
		ON CONFLICT (blocker_id, blocked_id) DO NOTHING
	`

	_, err := r.db.Exec(query, blockerID, blockedID)
	return err
}

func (r *userRepository) RemoveFromBlacklist(blockerID, blockedID string) error {
	query := `
		DELETE FROM blacklist
		WHERE blocker_id = $1 AND blocked_id = $2
	`

	_, err := r.db.Exec(query, blockerID, blockedID)
	return err
}

func (r *userRepository) GetBlacklist(userID string) ([]models.User, error) {
	query := `
		SELECT u.id, u.phone, u.username, u.avatar_url, u.role, u.created_at
		FROM blacklist b
		JOIN users u ON b.blocked_id = u.id
		WHERE b.blocker_id = $1 AND u.deleted_at IS NULL
		ORDER BY b.created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blockedUsers []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Phone,
			&user.Username,
			&user.AvatarURL,
			&user.Role,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		blockedUsers = append(blockedUsers, user)
	}

	return blockedUsers, nil
}

func (r *userRepository) CreateDeletionCode(userID, email, code string) error {
	query := `
		INSERT INTO deletion_codes (user_id, email, code, expires_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP + INTERVAL '15 minutes')
	`

	_, err := r.db.Exec(query, userID, email, code)
	return err
}

func (r *userRepository) VerifyDeletionCode(userID, email, code string) (bool, error) {
	query := `
		UPDATE deletion_codes
		SET is_used = TRUE
		WHERE user_id = $1 
			AND email = $2 
			AND code = $3
			AND is_used = FALSE
			AND expires_at > CURRENT_TIMESTAMP
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(query, userID, email, code).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
