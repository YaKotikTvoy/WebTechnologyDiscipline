package service

import (
	"errors"
	"fmt"

	"webchat/internal/models"
	"webchat/internal/repository"
)

type MessageService interface {
	SendMessage(req *models.SendMessageRequest, senderID string) (*models.Message, error)
	GetMessages(chatID *string, userID string, page, pageSize int) (*models.MessageResponse, error)
	EditMessage(messageID, content, userID string) error
	DeleteMessage(messageID, userID string, deleteForAll bool) error
	UploadFile(messageID, userID, filename, mimeType string, size int) (*models.File, error)
}

type messageService struct {
	messageRepo repository.MessageRepository
	chatRepo    repository.ChatRepository
	userRepo    repository.UserRepository
	uploadPath  string
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	uploadPath string,
) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		chatRepo:    chatRepo,
		userRepo:    userRepo,
		uploadPath:  uploadPath,
	}
}

func (s *messageService) SendMessage(req *models.SendMessageRequest, senderID string) (*models.Message, error) {
	if req.ChatID == nil && req.RecipientID == nil {
		return nil, errors.New("не указан чат или получатель")
	}

	if req.RecipientID != nil {
		blocked, err := s.messageRepo.CheckBlocked(senderID, *req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки блокировки: %w", err)
		}
		if blocked {
			return nil, errors.New("нельзя отправить сообщение заблокированному пользователю")
		}
	}

	if req.ChatID != nil {
		inChat, err := s.chatRepo.CheckUserInChat(*req.ChatID, senderID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступа: %w", err)
		}
		if !inChat {
			return nil, errors.New("вы не участник этого чата")
		}
	}

	message := &models.Message{
		ChatID:      req.ChatID,
		SenderID:    senderID,
		RecipientID: req.RecipientID,
		Content:     req.Content,
		IsEdited:    false,
		IsDeleted:   false,
	}

	err := s.messageRepo.CreateMessage(message)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания сообщения: %w", err)
	}

	if req.ChatID != nil {

	}

	return message, nil
}

func (s *messageService) GetMessages(chatID *string, userID string, page, pageSize int) (*models.MessageResponse, error) {
	if chatID != nil {
		inChat, err := s.chatRepo.CheckUserInChat(*chatID, userID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступа: %w", err)
		}
		if !inChat && !s.isPublicChat(*chatID) {
			return nil, errors.New("доступ запрещен")
		}
	}

	messages, total, err := s.messageRepo.GetMessages(chatID, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}

	if chatID != nil {
		s.messageRepo.UpdateLastSeen(*chatID, userID)
	}

	pages := (total + pageSize - 1) / pageSize

	return &models.MessageResponse{
		Messages: messages,
		Total:    total,
		Page:     page,
		Pages:    pages,
	}, nil
}

func (s *messageService) EditMessage(messageID, content, userID string) error {
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("ошибка получения сообщения: %w", err)
	}

	if message == nil {
		return errors.New("сообщение не найдено")
	}

	if message.SenderID != userID {
		return errors.New("можно редактировать только свои сообщения")
	}

	if message.IsDeleted {
		return errors.New("нельзя редактировать удаленное сообщение")
	}

	err = s.messageRepo.UpdateMessage(messageID, content)
	if err != nil {
		return fmt.Errorf("ошибка обновления сообщения: %w", err)
	}

	return nil
}

func (s *messageService) DeleteMessage(messageID, userID string, deleteForAll bool) error {
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("ошибка получения сообщения: %w", err)
	}

	if message == nil {
		return errors.New("сообщение не найдено")
	}

	if message.ChatID != nil {
		hasPermission, err := s.chatRepo.UserHasPermission(*message.ChatID, userID, "can_delete_messages")
		if err != nil {
			return fmt.Errorf("ошибка проверки прав: %w", err)
		}

		if !hasPermission && message.SenderID != userID {
			return errors.New("недостаточно прав для удаления сообщения")
		}
	} else {
		if message.SenderID != userID && message.RecipientID != nil && *message.RecipientID != userID {
			return errors.New("нельзя удалить чужое личное сообщение")
		}
	}

	if deleteForAll && message.ChatID != nil {
		err = s.messageRepo.DeleteMessageForAll(messageID)
	} else {
		err = s.messageRepo.DeleteMessageForUser(messageID, userID)
	}

	if err != nil {
		return fmt.Errorf("ошибка удаления сообщения: %w", err)
	}

	return nil
}

func (s *messageService) UploadFile(messageID, userID, filename, mimeType string, size int) (*models.File, error) {
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщения: %w", err)
	}

	if message == nil {
		return nil, errors.New("сообщение не найдено")
	}

	if message.SenderID != userID {
		return nil, errors.New("нельзя прикреплять файлы к чужим сообщениям")
	}

	fileURL := fmt.Sprintf("/uploads/%s", filename)

	file := &models.File{
		URL:      fileURL,
		Name:     filename,
		Size:     size,
		MimeType: mimeType,
	}

	err = s.messageRepo.AddFileToMessage(messageID, file)
	if err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	return file, nil
}

func (s *messageService) isPublicChat(chatID string) bool {
	chat, err := s.chatRepo.GetChatByID(chatID)
	if err != nil || chat == nil {
		return false
	}
	return chat.IsPublic
}
