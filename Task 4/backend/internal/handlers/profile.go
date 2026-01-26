package handlers

import (
	"net/http"

	"webchat/internal/models"
	"webchat/internal/service"
	"webchat/internal/utils"

	"github.com/labstack/echo/v4"
)

type ProfileHandler struct {
	profileService *service.ProfileService
}

func NewProfileHandler(profileService *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(string)

	user, err := h.profileService.GetUserByID(userID)
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

	user.Phone = h.profileService.MaskPhone(user.Phone)

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

	err := h.profileService.UpdateProfile(userID, &req)
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

	code, err := h.profileService.RequestDeletion(userID, req.Email)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Deletion code sent to email",
		"code":    code,
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

	err := h.profileService.VerifyDeletionCode(userID, req.Email, req.Code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Account deleted successfully",
	})
}
