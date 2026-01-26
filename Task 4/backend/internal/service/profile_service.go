package service

import (
	"fmt"
	"math/rand"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/pkg/notification"
)

type ProfileService struct {
	userRepo repository.UserRepository
	notifier *notification.ConsoleService
}

func NewProfileService(userRepo repository.UserRepository, notifier *notification.ConsoleService) *ProfileService {
	return &ProfileService{
		userRepo: userRepo,
		notifier: notifier,
	}
}

func (s *ProfileService) GetUserByID(userID string) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}

func (s *ProfileService) MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:4] + "***" + phone[len(phone)-3:]
}

func (s *ProfileService) RequestDeletion(userID, email string) (string, error) {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return "", fmt.Errorf("пользователь не найден")
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	err = s.notifier.SendEmail(email, code, "Удаление аккаунта")
	if err != nil {
		return "", err
	}

	err = s.userRepo.CreateDeletionCode(userID, email, code)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *ProfileService) VerifyDeletionCode(userID, email, code string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil || user == nil {
		return fmt.Errorf("пользователь не найден")
	}

	err = s.notifier.VerifyEmail(email, code, "Удаление аккаунта")
	if err != nil {
		return err
	}

	return s.userRepo.DeleteUser(userID)
}

func (s *ProfileService) UpdateProfile(userID string, req *models.UpdateProfileRequest) error {
	return s.userRepo.UpdateUser(userID, req)
}
