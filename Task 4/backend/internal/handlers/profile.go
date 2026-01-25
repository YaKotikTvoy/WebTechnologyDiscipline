package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	_ "time"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/internal/utils"
	"webchat/pkg/database"

	"github.com/labstack/echo/v4"
)

type ProfileHandler struct {
	userRepo  repository.UserRepository
	jwtSecret string
}

func NewProfileHandler(db *database.DB, jwtSecret string) *ProfileHandler {
	userRepo := repository.NewUserRepository(db)

	return &ProfileHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (h *ProfileHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get profile",
		})
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	user.Phone = maskPhone(user.Phone)

	return c.JSON(http.StatusOK, user)
}

func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req models.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	if err := utils.Validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	/*if req.Username != nil {

	}*/

	err := h.userRepo.UpdateUser(userID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update profile",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Profile updated successfully",
	})
}

func (h *ProfileHandler) RequestDelete(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req models.DeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	if err := utils.Validate.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	err := h.userRepo.CreateDeletionCode(userID, req.Email, code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create deletion code",
		})
	}

	go h.sendDeletionEmail(req.Email, code)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Deletion code sent to email",
	})
}

func (h *ProfileHandler) ConfirmDelete(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req models.DeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Code is required",
		})
	}

	valid, err := h.userRepo.VerifyDeletionCode(userID, req.Email, req.Code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to verify code",
		})
	}

	if !valid {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid or expired code",
		})
	}

	err = h.userRepo.DeleteUser(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete account",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}

func (h *ProfileHandler) sendDeletionEmail(email, code string) {
	fmt.Printf("[EMAIL] Sending deletion code to: %s\n", email)
	fmt.Printf("[EMAIL] Code: %s\n", code)
	fmt.Printf("[EMAIL] Message: Ваш код подтверждения удаления аккаунта: %s\n", code)
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:4] + "***" + phone[len(phone)-3:]
}
