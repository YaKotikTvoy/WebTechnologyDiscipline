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
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	notifier    *notification.ConsoleService
	jwtSecret   string
	sessionTTL  time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	notifier *notification.ConsoleService,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		notifier:    notifier,
		jwtSecret:   jwtSecret,
		sessionTTL:  24 * time.Hour,
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

func (s *AuthService) Register(req *models.RegisterRequest, code string) (*models.User, string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка создания пароля")
	}

	user := &models.User{
		Phone:        req.Phone,
		Role:         "user",
		IsActive:     true,
		StorageLimit: 100 * 1024 * 1024, // 100MB
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.CreateUser(user, string(hashedPassword))
	if err != nil {
		return nil, "", fmt.Errorf("ошибка создания пользователя: %w", err)
	}

	token, session, err := s.createSession(user.ID, "")
	if err != nil {
		return nil, "", fmt.Errorf("ошибка создания сессии: %w", err)
	}

	err = s.userRepo.UpdateUserStatus(user.ID, true)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Login(req *models.LoginRequest, ip string) (*models.User, string, error) {
	user, err := s.userRepo.GetUserByPhone(req.Phone)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка авторизации")
	}

	if user == nil {
		return nil, "", fmt.Errorf("пользователь не найден")
	}

	if !user.IsActive {
		return nil, "", fmt.Errorf("аккаунт деактивирован")
	}

	var storedHash string
	err = s.userRepo.GetDB().QueryRow(`
		SELECT password_hash FROM users 
		WHERE phone = $1 AND deleted_at IS NULL
	`, req.Phone).Scan(&storedHash)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка проверки пароля")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password))
	if err != nil {
		return nil, "", fmt.Errorf("неверный пароль")
	}

	token, session, err := s.createSession(user.ID, req.Device)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка создания сессии: %w", err)
	}

	err = s.userRepo.UpdateUserStatus(user.ID, true)
	if err != nil {
		return nil, "", fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Logout(token string) error {
	tokenHash := utils.HashToken(token)
	return s.sessionRepo.DeleteSession(tokenHash)
}

func (s *AuthService) ValidateSession(token string) (*models.User, error) {
	tokenHash := utils.HashToken(token)

	session, err := s.sessionRepo.GetSession(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена")
	}

	if session.ExpiresAt.Before(time.Now()) {
		s.sessionRepo.DeleteSession(tokenHash)
		return nil, fmt.Errorf("сессия истекла")
	}

	user, err := s.userRepo.GetUserByID(session.UserID)
	if err != nil || user == nil || !user.IsActive {
		s.sessionRepo.DeleteSession(tokenHash)
		return nil, fmt.Errorf("пользователь не активен")
	}

	s.sessionRepo.UpdateSessionActivity(tokenHash)
	s.userRepo.UpdateLastSeen(user.ID)

	return user, nil
}

func (s *AuthService) GetUserSessions(userID string) ([]models.UserSession, error) {
	return s.sessionRepo.GetUserSessions(userID)
}

func (s *AuthService) LogoutAllSessions(userID string) error {
	return s.sessionRepo.DeleteAllUserSessions(userID)
}

func (s *AuthService) LogoutOtherSessions(userID, currentToken string) error {
	tokenHash := utils.HashToken(currentToken)
	return s.sessionRepo.DeleteOtherUserSessions(userID, tokenHash)
}

func (s *AuthService) createSession(userID, device string) (string, *models.UserSession, error) {
	token, err := utils.GenerateJWT(userID, s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("ошибка генерации токена: %w", err)
	}

	tokenHash := utils.HashToken(token)
	session := &models.UserSession{
		UserID:         userID,
		TokenHash:      tokenHash,
		DeviceInfo:     &device,
		LastActivityAt: time.Now(),
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(s.sessionTTL),
	}

	err = s.sessionRepo.CreateSession(session)
	if err != nil {
		return "", nil, fmt.Errorf("ошибка сохранения сессии: %w", err)
	}

	return token, session, nil
}

func (s *AuthService) generateCode() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func isValidPhone(phone string) bool {
	regex := regexp.MustCompile(`^\+7\d{10}$`)
	return regex.MatchString(phone)
}
