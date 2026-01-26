package service

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/internal/utils"
	"webchat/pkg/notification"
)

type AuthService struct {
	userRepo  repository.UserRepository
	notifier  *notification.ConsoleService
	jwtSecret string
}

func NewAuthService(userRepo repository.UserRepository, notifier *notification.ConsoleService, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		notifier:  notifier,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) SendRegistrationCode(phone string) (string, error) {
	if !isValidPhone(phone) {
		return "", fmt.Errorf("неверный формат телефона")
	}

	existing, _ := s.userRepo.GetUserByPhone(phone)
	if existing != nil {
		return "", fmt.Errorf("номер уже зарегистрирован")
	}

	code := s.generateCode()
	err := s.notifier.SendSMS(phone, code, "Регистрация аккаунта")
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *AuthService) VerifyRegistrationCode(phone, code string) error {
	err := s.notifier.VerifySMS(phone, code, "Регистрация аккаунта")
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) Register(req *models.RegisterRequest, code string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("ошибка создания пароля")
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

	return token, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (string, string, error) {
	user, err := s.userRepo.GetUserByPhone(req.Phone)
	if err != nil {
		return "", "", fmt.Errorf("ошибка авторизации")
	}

	if user == nil {
		return "", "", fmt.Errorf("пользователь не найден")
	}

	if !user.IsActive {
		return "", "", fmt.Errorf("аккаунт деактивирован")
	}

	var storedHash string
	err = s.userRepo.GetDB().QueryRow("SELECT password_hash FROM users WHERE phone = $1 AND deleted_at IS NULL", req.Phone).Scan(&storedHash)
	if err != nil {
		return "", "", fmt.Errorf("ошибка проверки пароля")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password))
	if err != nil {
		return "", "", fmt.Errorf("неверный пароль")
	}

	token, err := utils.GenerateJWT(user.ID, user.Role, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("ошибка генерации токена")
	}

	return token, user.ID, nil
}

func (s *AuthService) generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func isValidPhone(phone string) bool {
	regex := regexp.MustCompile(`^\+7\d{10}$`)
	return regex.MatchString(phone)
}
