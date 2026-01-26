package handlers

import (
	"net/http"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/internal/service"
	"webchat/internal/utils"
	"webchat/pkg/database"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(db *database.DB) *ChatHandler {
	chatRepo := repository.NewChatRepository(db)
	userRepo := repository.NewUserRepository(db)
	chatService := service.NewChatService(chatRepo, userRepo)

	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) GetUserChats(c echo.Context) error {
	userID := c.Get("user_id").(string)

	chats, err := h.chatService.GetUserChats(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get chats",
		})
	}

	return c.JSON(http.StatusOK, chats)
}

func (h *ChatHandler) GetPublicChats(c echo.Context) error {
	chats, err := h.chatService.GetPublicChats()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get public chats",
		})
	}

	return c.JSON(http.StatusOK, chats)
}

func (h *ChatHandler) GetPublicChat(c echo.Context) error {
	chatID := c.Param("id")
	userID := ""

	claims := c.Get("user")
	if claims != nil {
		if userClaims, ok := claims.(jwt.MapClaims); ok {
			if id, ok := userClaims["user_id"].(string); ok {
				userID = id
			}
		}
	}

	chat, err := h.chatService.GetChat(chatID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Chat not found or access denied",
		})
	}

	return c.JSON(http.StatusOK, chat)
}

func (h *ChatHandler) GetChat(c echo.Context) error {
	chatID := c.Param("id")
	userID := c.Get("user_id").(string)

	chat, err := h.chatService.GetChat(chatID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": "Access denied",
		})
	}

	return c.JSON(http.StatusOK, chat)
}

func (h *ChatHandler) CreateChat(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req models.CreateChatRequest
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

	chat, err := h.chatService.CreateChat(&req, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create chat",
		})
	}

	return c.JSON(http.StatusCreated, chat)
}

func (h *ChatHandler) UpdateChat(c echo.Context) error {
	chatID := c.Param("id")
	userID := c.Get("user_id").(string)

	var req models.UpdateChatRequest
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

	err := h.chatService.UpdateChat(chatID, &req, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Chat updated successfully",
	})
}

func (h *ChatHandler) CreateInvite(c echo.Context) error {
	chatID := c.Param("id")
	userID := c.Get("user_id").(string)

	var req models.CreateInviteRequest
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

	invite, err := h.chatService.CreateInvite(chatID, userID, &req)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, invite)
}

func (h *ChatHandler) JoinChat(c echo.Context) error {
	inviteCode := c.Param("code")
	userID := c.Get("user_id").(string)

	err := h.chatService.JoinChat(inviteCode, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully joined chat",
	})
}

func (h *ChatHandler) AssignRole(c echo.Context) error {
	chatID := c.Param("id")
	assignerID := c.Get("user_id").(string)

	var req models.AssignRoleRequest
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

	err := h.chatService.AssignRole(chatID, assignerID, &req)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Role assigned successfully",
	})
}

func (h *ChatHandler) GetChatMembers(c echo.Context) error {
	chatID := c.Param("id")
	userID := c.Get("user_id").(string)

	members, err := h.chatService.GetChatMembers(chatID, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, members)
}

func (h *ChatHandler) RemoveMember(c echo.Context) error {
	chatID := c.Param("id")
	memberID := c.Param("userId")
	removerID := c.Get("user_id").(string)

	err := h.chatService.RemoveMember(chatID, removerID, memberID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Member removed successfully",
	})
}
