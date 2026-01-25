package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

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

	// Параметры пагинации
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// Параметры чата
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

	// Получаем файл
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "No file uploaded",
		})
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	// Создаем уникальное имя файла
	filename := filepath.Join(h.uploadPath, file.Filename)

	// Создаем файл на диске
	dst, err := os.Create(filename)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create file",
		})
	}
	defer dst.Close()

	// Копируем содержимое
	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save file",
		})
	}

	// Сохраняем информацию о файле в БД
	uploadedFile, err := h.messageService.UploadFile(
		messageID,
		userID,
		file.Filename,
		file.Header.Get("Content-Type"),
		int(file.Size),
	)

	if err != nil {
		// Удаляем файл если не удалось сохранить в БД
		os.Remove(filename)
		return c.JSON(http.StatusForbidden, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, uploadedFile)
}
