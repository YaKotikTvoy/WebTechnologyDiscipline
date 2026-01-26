package models

import (
	"time"
)

type User struct {
	ID           string     `json:"id"`
	Phone        string     `json:"phone"`
	Username     *string    `json:"username,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	Role         string     `json:"role"`
	Email        *string    `json:"email,omitempty"`
	IsActive     bool       `json:"is_active"`
	IsOnline     bool       `json:"is_online"`
	LastSeenAt   time.Time  `json:"last_seen_at"`
	StorageUsed  int64      `json:"storage_used"`
	StorageLimit int64      `json:"storage_limit"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"-"`
}

type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required,min=10,max=20"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
	Device   string `json:"device,omitempty"`
}

type UpdateProfileRequest struct {
	Username  *string `json:"username,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Email     *string `json:"email,omitempty"`
}

type DeleteRequest struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code,omitempty"`
}

type UserSession struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	DeviceInfo     *string   `json:"device_info,omitempty"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	LastActivityAt time.Time `json:"last_activity_at"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type ContactRequest struct {
	ID          string    `json:"id"`
	RequesterID string    `json:"requester_id"`
	RecipientID string    `json:"recipient_id"`
	Status      string    `json:"status"`
	Message     *string   `json:"message,omitempty"`
	Requester   *User     `json:"requester,omitempty"`
	Recipient   *User     `json:"recipient,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateContactRequest struct {
	Status string `json:"status" validate:"required,oneof=accepted rejected"`
}

type Contact struct {
	UserID    string    `json:"user_id"`
	ContactID string    `json:"contact_id"`
	Alias     *string   `json:"alias,omitempty"`
	User      *User     `json:"user,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type SearchUserResponse struct {
	User           *User  `json:"user"`
	IsContact      bool   `json:"is_contact"`
	HasPendingReq  bool   `json:"has_pending_request"`
	CanSendRequest bool   `json:"can_send_request"`
	RequestID      string `json:"request_id,omitempty"`
}
