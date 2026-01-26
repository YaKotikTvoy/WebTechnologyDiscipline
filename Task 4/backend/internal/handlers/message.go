package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/internal/service"
	"webchat/internal/utils"
	"webchat/pkg/database"

	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	messageService service.MessageService
	uploadPath     string
}

func NewMessageHandler(db *database.DB) *MessageHandler {
	messageRepo := repository.NewMessageRepository(db)
	chatRepo := repository.NewChatRepository(db)
	userRepo := repository.NewUserRepository(db)

	uploadPath := "uploads"
	os.MkdirAll(uploadPath, 0755)

	messageService := service.NewMessageService(messageRepo, chatRepo, userRepo, uploadPath)

	return &MessageHandler{
		messageService: messageService,
		uploadPath:     uploadPath,
	}
}

func (h *MessageHandler) GetMessages(c echo.Context) error {
	userID := c.Get("user_id").(string)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	chatID := c.QueryParam("chat_id")
	var chatIDPtr *string
	if chatID != "" {
		chatIDPtr = &chatID
	}

	response, err := h.messageService.GetMessages(chatIDPtr, userID, page, pageSize)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (h *MessageHandler) SendMessage(c echo.Context) error {
	userID := c.Get("user_id").(string)

	var req models.SendMessageRequest
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

	message, err := h.messageService.SendMessage(&req, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, message)
}

func (h *MessageHandler) EditMessage(c echo.Context) error {
	messageID := c.Param("id")
	userID := c.Get("user_id").(string)

	var req models.EditMessageRequest
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

	err := h.messageService.EditMessage(messageID, req.Content, userID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Message updated successfully",
	})
}

func (h *MessageHandler) DeleteMessage(c echo.Context) error {
	messageID := c.Param("id")
	userID := c.Get("user_id").(string)

	deleteForAll := c.QueryParam("for_all") == "true"

	err := h.messageService.DeleteMessage(messageID, userID, deleteForAll)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Message deleted successfully",
	})
}

func (h *MessageHandler) UploadFile(c echo.Context) error {
	messageID := c.FormValue("message_id")
	userID := c.Get("user_id").(string)

	form, err := c.MultipartForm()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No files uploaded",
		})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No files uploaded",
		})
	}

	var uploadedFiles []models.File

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to open file",
			})
		}
		defer src.Close()

		uniqueFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		filename := filepath.Join(h.uploadPath, uniqueFilename)

		if err := os.MkdirAll(h.uploadPath, 0755); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to create upload directory",
			})
		}

		uploadedFile, err := h.messageService.UploadFile(
			messageID,
			userID,
			uniqueFilename,
			file.Header.Get("Content-Type"),
			int(file.Size),
		)

		if err != nil {
			os.Remove(filename)
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": err.Error(),
			})
		}

		uploadedFiles = append(uploadedFiles, *uploadedFile)
	}

	return c.JSON(http.StatusCreated, uploadedFiles)
}

func (h *MessageHandler) GetDirectMessages(c echo.Context) error {
	userID := c.Get("user_id").(string)

	contactID := c.QueryParam("contact_id")
	if contactID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "contact_id parameter is required",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	response, err := h.messageService.GetDirectMessages(userID, contactID, page, pageSize)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
