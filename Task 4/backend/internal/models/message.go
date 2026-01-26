package models

import (
	"time"
)

type Message struct {
	ID                 string    `json:"id"`
	ChatID             *string   `json:"chat_id,omitempty"`
	SenderID           string    `json:"sender_id"`
	RecipientID        *string   `json:"recipient_id,omitempty"`
	SenderUsername     *string   `json:"sender_username,omitempty"`
	SenderAvatar       *string   `json:"sender_avatar,omitempty"`
	Content            string    `json:"content"`
	IsEdited           bool      `json:"is_edited"`
	IsDeleted          bool      `json:"is_deleted"`
	DeletedBySender    bool      `json:"deleted_by_sender,omitempty"`
	DeletedByRecipient bool      `json:"deleted_by_recipient,omitempty"`
	ReadByRecipient    bool      `json:"read_by_recipient,omitempty"`
	Files              []File    `json:"files,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type File struct {
	ID         string    `json:"id"`
	PublicURL  string    `json:"url"`
	Name       string    `json:"name"`
	Size       int       `json:"size"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at,omitempty"`
}

type SendMessageRequest struct {
	ChatID      *string `json:"chat_id,omitempty"`
	RecipientID *string `json:"recipient_id,omitempty"`
	Content     string  `json:"content" validate:"required,min=1,max=5000"`
}

type EditMessageRequest struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

type Pagination struct {
	Page     int `json:"page" query:"page" validate:"min=1"`
	PageSize int `json:"page_size" query:"page_size" validate:"min=1,max=100"`
}

type MessageResponse struct {
	Messages []Message `json:"messages"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	Pages    int       `json:"pages"`
}

type UploadFileResponse struct {
	FileID    string `json:"file_id"`
	PublicURL string `json:"public_url"`
	FileName  string `json:"file_name"`
	FileSize  int    `json:"file_size"`
	MimeType  string `json:"mime_type"`
}
