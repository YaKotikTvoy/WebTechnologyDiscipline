package models

import (
	"time"
)

type User struct {
	ID        string     `json:"id"`
	Phone     string     `json:"phone"`
	Username  *string    `json:"username,omitempty"`
	AvatarURL *string    `json:"avatar_url,omitempty"`
	Role      string     `json:"role"`
	Email     *string    `json:"email,omitempty"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type RegisterRequest struct {
	Phone    string `json:"phone" validate:"required,min=10,max=20"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type LoginRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
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
