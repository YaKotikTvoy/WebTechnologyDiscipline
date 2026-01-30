package models

import (
	"regexp"
	"time"
)

type User struct {
	ID           uint      `json:"id"`
	Phone        string    `json:"phone" gorm:"unique;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Username     string    `json:"username"`
	CreatedAt    time.Time `json:"created_at"`
	LastSeenAt   time.Time `json:"last_seen_at"`
}

type UpdateProfileRequest struct {
	Username string `json:"username" validate:"max=50"`
}

type Chat struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	CreatedBy    uint      `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	IsSearchable bool      `json:"is_searchable"`
	Members      []User    `json:"members" gorm:"many2many:chat_members;"`
}

type ChatMember struct {
	ID      uint `json:"id"`
	ChatID  uint `json:"chat_id"`
	UserID  uint `json:"user_id"`
	IsAdmin bool `json:"is_admin"`
}

type ChatJoinRequest struct {
	ID        uint      `json:"id"`
	ChatID    uint      `json:"chat_id"`
	UserID    uint      `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Chat      Chat      `json:"chat" gorm:"foreignKey:ChatID"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

type Message struct {
	ID        uint          `json:"id"`
	ChatID    uint          `json:"chat_id"`
	SenderID  uint          `json:"sender_id"`
	Content   string        `json:"content"`
	Type      string        `json:"type" gorm:"default:'regular'"`
	IsDeleted bool          `json:"is_deleted"`
	IsEdited  bool          `json:"is_edited" gorm:"default:false"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt *time.Time    `json:"updated_at,omitempty"`
	Sender    User          `json:"sender" gorm:"foreignKey:SenderID"`
	Files     []MessageFile `json:"files"`
	Readers   []User        `json:"readers" gorm:"many2many:message_readers;"`
}

type MessageReader struct {
	ID        uint      `json:"id"`
	MessageID uint      `json:"message_id"`
	UserID    uint      `json:"user_id"`
	ReadAt    time.Time `json:"read_at"`
}

type MessageFile struct {
	ID         uint      `json:"id"`
	MessageID  uint      `json:"message_id"`
	Filename   string    `json:"filename"`
	Filepath   string    `json:"filepath"`
	Filesize   int64     `json:"filesize"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type UserSession struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required,min=10"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type CreateChatRequest struct {
	Name         string   `json:"name"`
	Type         string   `json:"type" validate:"required,oneof=private group"`
	MemberPhones []string `json:"member_phones"`
}

type SendMessageRequest struct {
	Content string `json:"content" validate:"required,max=5000"`
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Notification struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func ValidatePhone(phone string) bool {
	re := regexp.MustCompile(`^7\d{10}$`)
	return re.MatchString(phone)
}

type RegistrationCode struct {
	ID        uint      `json:"id"`
	Phone     string    `json:"phone"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type VerifyCodeRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Code     string `json:"code" validate:"required,min=6,max=6"`
	Password string `json:"password" validate:"required,min=6"`
}

type ChatInvite struct {
	ID        uint      `json:"id"`
	ChatID    uint      `json:"chat_id"`
	InviterID uint      `json:"inviter_id"`
	UserID    uint      `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Chat      Chat      `json:"chat" gorm:"foreignKey:ChatID"`
	Inviter   User      `json:"inviter" gorm:"foreignKey:InviterID"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}

type TempPassword struct {
	ID        uint      `json:"id"`
	Phone     string    `json:"phone"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}
