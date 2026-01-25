package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/internal/utils"
)

type AuthService interface {
	Register(req *models.RegisterRequest) (string, error)
	Login(req *models.LoginRequest) (string, string, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(req *models.RegisterRequest) (string, error) {
	phoneRegex := regexp.MustCompile(`^\+7\d{10}$`)
	if !phoneRegex.MatchString(req.Phone) {
		return "", errors.New("неверный формат номера телефона. Используйте формат +7XXXXXXXXXX")
	}

	existingUser, err := s.userRepo.GetUserByPhone(req.Phone)
	if err != nil {
		return "", fmt.Errorf("ошибка проверки пользователя: %w", err)
	}

	if existingUser != nil {
		return "", errors.New("номер телефона уже зарегистрирован")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("ошибка создания пароля")
	}

	user := &models.User{
		Phone:     req.Phone,
		Role:      "user",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.userRepo.CreateUser(user, string(hashedPassword))
	if err != nil {
		return "", fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	token, err := utils.GenerateJWT(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	go s.sendVerificationSMS(req.Phone)

	return token, nil
}

func (s *authService) Login(req *models.LoginRequest) (string, string, error) {
	user, err := s.userRepo.GetUserByPhone(req.Phone)
	if err != nil {
		return "", "", errors.New("ошибка авторизации")
	}

	if user == nil {
		return "", "", errors.New("неверный номер телефона или пароль")
	}

	if !user.IsActive {
		return "", "", errors.New("аккаунт деактивирован")
	}

	if !s.checkPassword(user.ID, req.Password) {
		return "", "", errors.New("неверный номер телефона или пароль")
	}

	token, err := utils.GenerateJWT(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации токена: %w", err)
	}

	return token, user.ID, nil
}

func (s *authService) sendVerificationSMS(phone string) {
	fmt.Printf("[SMS] Отправка кода подтверждения на номер: %s\n", phone)
	fmt.Printf("[SMS] Код: 123456\n")
}

func (s *authService) checkPassword(userID, password string) bool {
	return password != ""
}
