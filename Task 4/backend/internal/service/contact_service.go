package service

import (
	"errors"
	"fmt"
	"time"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/pkg/websocket"
)

type ContactService struct {
	contactRepo   repository.ContactRepository
	userRepo      repository.UserRepository
	blacklistRepo repository.BlacklistRepository
	wsHub         *websocket.Hub
}

func NewContactService(
	contactRepo repository.ContactRepository,
	userRepo repository.UserRepository,
	blacklistRepo repository.BlacklistRepository,
	wsHub *websocket.Hub,
) *ContactService {
	return &ContactService{
		contactRepo:   contactRepo,
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		wsHub:         wsHub,
	}
}

func (s *ContactService) SearchUser(userID, phone string) (*models.SearchUserResponse, error) {
	if userID == "" {
		return nil, errors.New("требуется авторизация")
	}

	user, err := s.userRepo.GetUserByPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("ошибка поиска пользователя: %w", err)
	}

	if user == nil {
		return nil, errors.New("пользователь не найден")
	}

	if user.ID == userID {
		return nil, errors.New("нельзя добавить самого себя")
	}

	blocked, err := s.blacklistRepo.CheckBlocked(userID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки блокировки: %w", err)
	}
	if blocked {
		return nil, errors.New("пользователь заблокирован")
	}

	response := &models.SearchUserResponse{
		User: user,
	}

	isContact, err := s.contactRepo.IsContact(userID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки контакта: %w", err)
	}
	response.IsContact = isContact

	if !isContact {
		hasPending, requestID, err := s.contactRepo.HasPendingRequest(userID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки заявки: %w", err)
		}

		response.HasPendingReq = hasPending
		response.RequestID = requestID
		response.CanSendRequest = !hasPending
	}

	return response, nil
}

func (s *ContactService) SendRequest(userID, recipientID string, message *string) (*models.ContactRequest, error) {
	if userID == recipientID {
		return nil, errors.New("нельзя отправить заявку самому себе")
	}

	recipient, err := s.userRepo.GetUserByID(recipientID)
	if err != nil || recipient == nil {
		return nil, errors.New("пользователь не найден")
	}

	blocked, err := s.blacklistRepo.CheckBlocked(userID, recipientID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки блокировки: %w", err)
	}
	if blocked {
		return nil, errors.New("нельзя отправить заявку заблокированному пользователю")
	}

	isContact, err := s.contactRepo.IsContact(userID, recipientID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки контакта: %w", err)
	}
	if isContact {
		return nil, errors.New("пользователь уже в контактах")
	}

	hasPending, _, err := s.contactRepo.HasPendingRequest(userID, recipientID)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки заявки: %w", err)
	}
	if hasPending {
		return nil, errors.New("заявка уже отправлена")
	}

	request := &models.ContactRequest{
		RequesterID: userID,
		RecipientID: recipientID,
		Status:      "pending",
		Message:     message,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.contactRepo.CreateRequest(request)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания заявки: %w", err)
	}

	if s.wsHub != nil {
		s.wsHub.SendToUser(recipientID, models.WSEvent{
			Type: "contact_request",
			Payload: models.WSContactRequest{
				Request: *request,
				Action:  "new",
			},
		})
	}

	return request, nil
}

func (s *ContactService) GetRequests(userID string) ([]models.ContactRequest, error) {
	return s.contactRepo.GetUserRequests(userID)
}

func (s *ContactService) UpdateRequest(requestID, userID string, status string) (*models.ContactRequest, error) {
	request, err := s.contactRepo.GetRequestByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения заявки: %w", err)
	}

	if request == nil {
		return nil, errors.New("заявка не найдена")
	}

	if request.RecipientID != userID {
		return nil, errors.New("нет прав для изменения заявки")
	}

	if request.Status != "pending" {
		return nil, errors.New("заявка уже обработана")
	}

	if status != "accepted" && status != "rejected" {
		return nil, errors.New("неверный статус")
	}

	err = s.contactRepo.UpdateRequestStatus(requestID, status)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления заявки: %w", err)
	}

	request.Status = status
	request.UpdatedAt = time.Now()

	if s.wsHub != nil {
		s.wsHub.SendToUser(request.RequesterID, models.WSEvent{
			Type: "contact_request",
			Payload: models.WSContactRequest{
				Request: *request,
				Action:  "update",
			},
		})
	}

	return request, nil
}

func (s *ContactService) CancelRequest(requestID, userID string) error {
	request, err := s.contactRepo.GetRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("ошибка получения заявки: %w", err)
	}

	if request == nil {
		return errors.New("заявка не найдена")
	}

	if request.RequesterID != userID {
		return errors.New("нет прав для отмены заявки")
	}

	if request.Status != "pending" {
		return errors.New("заявка уже обработана")
	}

	return s.contactRepo.DeleteRequest(requestID)
}

func (s *ContactService) GetContacts(userID string) ([]models.Contact, error) {
	return s.contactRepo.GetUserContacts(userID)
}

func (s *ContactService) RemoveContact(userID, contactID string) error {
	if userID == contactID {
		return errors.New("нельзя удалить самого себя")
	}

	isContact, err := s.contactRepo.IsContact(userID, contactID)
	if err != nil {
		return fmt.Errorf("ошибка проверки контакта: %w", err)
	}
	if !isContact {
		return errors.New("пользователь не в контактах")
	}

	err = s.contactRepo.DeleteContact(userID, contactID)
	if err != nil {
		return fmt.Errorf("ошибка удаления контакта: %w", err)
	}

	err = s.contactRepo.DeleteContact(contactID, userID)
	if err != nil {
		return fmt.Errorf("ошибка удаления контакта: %w", err)
	}

	return nil
}
