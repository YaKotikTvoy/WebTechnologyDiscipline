package handlers

import (
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"webchat/internal/models"
	"webchat/internal/utils"
	"webchat/pkg/database"
)

type AuthHandler struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthHandler(db *database.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := utils.Validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	phoneRegex := regexp.MustCompile(`^\+7\d{10}$`)
	if !phoneRegex.MatchString(req.Phone) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid phone format"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}

	var userID string
	err = h.db.QueryRow(`
		INSERT INTO users (phone, password_hash, role)
		VALUES ($1, $2, 'user')
		RETURNING id
	`, req.Phone, string(hashedPassword)).Scan(&userID)

	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_phone_key\"" {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Phone already registered"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	token, err := utils.GenerateJWT(userID, "user", h.jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"token":   token,
		"user_id": userID,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := utils.Validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	var user struct {
		ID           string
		PasswordHash string
		Role         string
		IsActive     bool
	}

	err := h.db.QueryRow(`
		SELECT id, password_hash, role, is_active
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
	`, req.Phone).Scan(&user.ID, &user.PasswordHash, &user.Role, &user.IsActive)

	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	if !user.IsActive {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Account is deactivated"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials"})
	}

	token, err := utils.GenerateJWT(user.ID, user.Role, h.jwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"token":   token,
		"user_id": user.ID,
		"role":    user.Role,
	})
}
