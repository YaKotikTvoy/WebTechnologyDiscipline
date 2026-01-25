package models

import (
	"time"
)

type Chat struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     *string   `json:"description,omitempty"`
	AvatarURL       *string   `json:"avatar_url,omitempty"`
	IsPublic        bool      `json:"is_public"`
	CreatorID       string    `json:"creator_id"`
	OnlyAdminInvite bool      `json:"only_admin_invite"`
	MemberCount     int       `json:"member_count,omitempty"`
	UnreadCount     int       `json:"unread_count,omitempty"`
	LastMessage     *string   `json:"last_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateChatRequest struct {
	Name            string  `json:"name" validate:"required,min=1,max=255"`
	Description     *string `json:"description,omitempty"`
	IsPublic        bool    `json:"is_public"`
	OnlyAdminInvite bool    `json:"only_admin_invite"`
}

type UpdateChatRequest struct {
	Name            *string `json:"name,omitempty" validate:"max=255"`
	Description     *string `json:"description,omitempty"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	OnlyAdminInvite *bool   `json:"only_admin_invite,omitempty"`
}

type ChatMember struct {
	UserID    string    `json:"user_id"`
	Username  *string   `json:"username,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	RoleName  string    `json:"role_name"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ChatRole struct {
	UserID            string `json:"user_id"`
	RoleName          string `json:"role_name"`
	CanDeleteMessages bool   `json:"can_delete_messages"`
	CanRemoveUsers    bool   `json:"can_remove_users"`
	CanAssignRoles    bool   `json:"can_assign_roles"`
}

type AssignRoleRequest struct {
	UserID            string `json:"user_id" validate:"required"`
	RoleName          string `json:"role_name" validate:"required,oneof=member moderator admin owner"`
	CanDeleteMessages bool   `json:"can_delete_messages"`
	CanRemoveUsers    bool   `json:"can_remove_users"`
	CanAssignRoles    bool   `json:"can_assign_roles"`
}

type Invitation struct {
	ID         string    `json:"id"`
	ChatID     string    `json:"chat_id"`
	InviteCode string    `json:"invite_code"`
	InviterID  *string   `json:"inviter_id,omitempty"`
	IsUsed     bool      `json:"is_used"`
	UsedBy     *string   `json:"used_by,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateInviteRequest struct {
	ExpiresInHours int `json:"expires_in_hours" validate:"required,min=1,max=720"`
}
