package service

import (
	"errors"
	"fmt"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/pkg/websocket"
)

type MessageService interface {
	SendMessage(req *models.SendMessageRequest, senderID string) (*models.Message, error)
	GetMessages(chatID *string, userID string, page, pageSize int) (*models.MessageResponse, error)
	EditMessage(messageID, content, userID string) error
	DeleteMessage(messageID, userID string, deleteForAll bool) error
	UploadFile(messageID, userID, filename, mimeType string, size int) (*models.File, error)
	GetDirectMessages(userID, contactID string, page, pageSize int) (*models.MessageResponse, error)
	MarkAsRead(messageID, userID string) error
}

type messageService struct {
	messageRepo   repository.MessageRepository
	chatRepo      repository.ChatRepository
	userRepo      repository.UserRepository
	contactRepo   repository.ContactRepository
	blacklistRepo repository.BlacklistRepository
	fileService   *FileService
	wsHub         *websocket.Hub
}

func NewMessageService(
	messageRepo repository.MessageRepository,
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	contactRepo repository.ContactRepository,
	blacklistRepo repository.BlacklistRepository,
	fileService *FileService,
	wsHub *websocket.Hub,
) MessageService {
	return &messageService{
		messageRepo:   messageRepo,
		chatRepo:      chatRepo,
		userRepo:      userRepo,
		contactRepo:   contactRepo,
		blacklistRepo: blacklistRepo,
		fileService:   fileService,
		wsHub:         wsHub,
	}
}

func (s *messageService) SendMessage(req *models.SendMessageRequest, senderID string) (*models.Message, error) {
	if req.ChatID == nil && req.RecipientID == nil {
		return nil, errors.New("не указан чат или получатель")
	}

	if req.RecipientID != nil {
		blocked, err := s.blacklistRepo.CheckBlocked(senderID, *req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки блокировки: %w", err)
		}
		if blocked {
			return nil, errors.New("нельзя отправить сообщение заблокированному пользователю")
		}

		isContact, err := s.contactRepo.IsContact(senderID, *req.RecipientID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки контактов: %w", err)
		}
		if !isContact {
			return nil, errors.New("можно отправлять сообщения только контактам")
		}
	}

	if req.ChatID != nil {
		inChat, err := s.chatRepo.CheckUserInChat(*req.ChatID, senderID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступа: %w", err)
		}
		if !inChat {
			isPublic, err := s.chatRepo.CheckPublicChat(*req.ChatID)
			if err != nil {
				return nil, fmt.Errorf("ошибка проверки публичного чата: %w", err)
			}
			if !isPublic {
				return nil, errors.New("вы не участник этого чата")
			}
		}
	}

	message := &models.Message{
		ChatID:          req.ChatID,
		SenderID:        senderID,
		RecipientID:     req.RecipientID,
		Content:         req.Content,
		IsEdited:        false,
		IsDeleted:       false,
		ReadByRecipient: false,
	}

	err := s.messageRepo.CreateMessage(message)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания сообщения: %w", err)
	}

	if s.wsHub != nil {
		wsEvent := models.WSEvent{
			Type: "message",
			Payload: models.WSMessage{
				ChatID:      req.ChatID,
				RecipientID: req.RecipientID,
				Message:     *message,
			},
		}

		if req.ChatID != nil {
			s.wsHub.SendToChat(*req.ChatID, senderID, wsEvent)
		} else if req.RecipientID != nil {
			s.wsHub.SendToUser(*req.RecipientID, wsEvent)
		}
	}

	return message, nil
}

func (s *messageService) GetMessages(chatID *string, userID string, page, pageSize int) (*models.MessageResponse, error) {
	if chatID != nil {
		inChat, err := s.chatRepo.CheckUserInChat(*chatID, userID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки доступа: %w", err)
		}

		if !inChat && userID != "" {
			isPublic, err := s.chatRepo.CheckPublicChat(*chatID)
			if err != nil {
				return nil, fmt.Errorf("ошибка проверки публичного чата: %w", err)
			}
			if !isPublic {
				return nil, errors.New("доступ запрещен")
			}
		}
	}

	messages, total, err := s.messageRepo.GetMessages(chatID, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}

	if chatID != nil && userID != "" {
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
		if message.ChatID != nil {
			hasPermission, err := s.chatRepo.UserHasPermission(*message.ChatID, userID, "can_delete_messages")
			if err != nil {
				return fmt.Errorf("ошибка проверки прав: %w", err)
			}
			if !hasPermission {
				return errors.New("недостаточно прав для редактирования сообщения")
			}
		} else {
			return errors.New("можно редактировать только свои сообщения")
		}
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

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	if user.StorageUsed+int64(size) > user.StorageLimit {
		return nil, errors.New("превышен лимит хранилища")
	}

	file, err := s.fileService.UploadFile(userID, filename, mimeType, size)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	err = s.messageRepo.AddFileToMessage(messageID, file)
	if err != nil {
		s.fileService.DeleteFile(file.ID)
		return nil, fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	return file, nil
}

func (s *messageService) GetDirectMessages(userID, contactID string, page, pageSize int) (*models.MessageResponse, error) {
	isContact, err := s.contactRepo.IsContact(userID, contactID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки контактов: %w", err)
	}
	if !isContact {
		return nil, errors.New("пользователь не в ваших контактах")
	}

	blocked, err := s.blacklistRepo.CheckBlocked(userID, contactID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки блокировки: %w", err)
	}
	if blocked {
		return nil, errors.New("доступ запрещен (пользователь заблокирован)")
	}

	messages, total, err := s.messageRepo.GetDirectMessages(userID, contactID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сообщений: %w", err)
	}

	err = s.messageRepo.MarkMessagesAsRead(userID, contactID)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса прочтения: %w", err)
	}

	pages := (total + pageSize - 1) / pageSize

	return &models.MessageResponse{
		Messages: messages,
		Total:    total,
		Page:     page,
		Pages:    pages,
	}, nil
}

func (s *messageService) MarkAsRead(messageID, userID string) error {
	message, err := s.messageRepo.GetMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("ошибка получения сообщения: %w", err)
	}

	if message == nil {
		return errors.New("сообщение не найдено")
	}

	if message.RecipientID == nil || *message.RecipientID != userID {
		return errors.New("нельзя пометить как прочитанное чужое сообщение")
	}

	err = s.messageRepo.MarkMessageAsRead(messageID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	return nil
}
