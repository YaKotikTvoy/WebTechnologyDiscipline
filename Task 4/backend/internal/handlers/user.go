package handlers

import (
	"net/http"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/pkg/database"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(db *database.DB) *UserHandler {
	userRepo := repository.NewUserRepository(db)

	return &UserHandler{
		userRepo: userRepo,
	}
}

func (h *UserHandler) GetContacts(c echo.Context) error {
	userID := c.Get("user_id").(string)

	contacts, err := h.userRepo.GetContacts(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get contacts",
		})
	}

	return c.JSON(http.StatusOK, contacts)
}

func (h *UserHandler) AddContact(c echo.Context) error {
	userID := c.Get("user_id").(string)

	type AddContactRequest struct {
		ContactID string  `json:"contact_id" validate:"required"`
		Alias     *string `json:"alias,omitempty"`
	}

	var req AddContactRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	if userID == req.ContactID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Cannot add yourself as contact",
		})
	}

	contact, err := h.userRepo.GetUserByID(req.ContactID)
	if err != nil || contact == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	err = h.userRepo.AddContact(userID, req.ContactID, req.Alias)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add contact",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Contact added successfully",
	})
}

func (h *UserHandler) RemoveContact(c echo.Context) error {
	userID := c.Get("user_id").(string)
	contactID := c.Param("id")

	err := h.userRepo.RemoveContact(userID, contactID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to remove contact",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Contact removed successfully",
	})
}

func (h *UserHandler) AddToBlacklist(c echo.Context) error {
	userID := c.Get("user_id").(string)

	type BlacklistRequest struct {
		BlockedID string `json:"blocked_id" validate:"required"`
	}

	var req BlacklistRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request",
		})
	}

	if userID == req.BlockedID {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Cannot block yourself",
		})
	}

	blockedUser, err := h.userRepo.GetUserByID(req.BlockedID)
	if err != nil || blockedUser == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	err = h.userRepo.AddToBlacklist(userID, req.BlockedID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add to blacklist",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "User added to blacklist",
	})
}

func (h *UserHandler) RemoveFromBlacklist(c echo.Context) error {
	userID := c.Get("user_id").(string)
	blockedID := c.Param("id")

	err := h.userRepo.RemoveFromBlacklist(userID, blockedID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to remove from blacklist",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "User removed from blacklist",
	})
}

func (h *UserHandler) GetBlacklist(c echo.Context) error {
	userID := c.Get("user_id").(string)

	blacklist, err := h.userRepo.GetBlacklist(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get blacklist",
		})
	}

	return c.JSON(http.StatusOK, blacklist)
}

func (h *UserHandler) SearchUser(c echo.Context) error {
	phone := c.QueryParam("phone")
	if phone == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Phone parameter is required",
		})
	}

	user, err := h.userRepo.GetUserByPhone(phone)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to search user",
		})
	}

	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, []models.User{*user})
}
