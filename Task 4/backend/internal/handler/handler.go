package handler

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
	"webchat/internal/models"
	"webchat/internal/service"
	"webchat/internal/ws"

	"github.com/gorilla/websocket"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Handler struct {
	service *service.Service
	hub     *ws.Hub
}

func NewHandler(service *service.Service, hub *ws.Hub) *Handler {
	return &Handler{
		service: service,
		hub:     hub,
	}
}

func (h *Handler) RegisterEndpoints(e *echo.Echo) {
	api := e.Group("/api")

	api.POST("/auth/register", h.Register)
	api.POST("/auth/login", h.Login)

	api.POST("/auth/verify", h.VerifyCode)
	api.POST("/auth/resend-code", h.ResendCode)

	auth := api.Group("")
	auth.Use(h.AuthMiddleware)

	auth.PUT("/auth/profile", h.UpdateProfile)
	auth.GET("/auth/me", h.GetMe)
	auth.POST("/auth/logout", h.Logout)
	auth.POST("/auth/logout-all", h.LogoutAll)

	auth.GET("/auth/me", h.GetMe)
	auth.POST("/auth/logout", h.Logout)
	auth.POST("/contacts", h.AddContact)

	auth.DELETE("/chats/:chat_id/messages/:message_id", h.DeleteMessage)
	auth.PUT("/chats/:chat_id/messages/:message_id", h.EditMessage)

	auth.GET("/users/search", h.SearchUser)

	auth.GET("/chats", h.GetChats)
	auth.POST("/chats", h.CreateChat)
	auth.GET("/chats/:id", h.GetChat)
	auth.POST("/chats/:id/members", h.AddChatMember)
	auth.DELETE("/chats/:id/members/:user_id", h.RemoveChatMember)

	auth.GET("/chats/:id/messages", h.GetMessages)
	auth.POST("/chats/:id/messages", h.SendMessage)
	auth.DELETE("/messages/:id", h.DeleteMessage)

	auth.POST("/chats/:id/decline", h.DeclineChatInvite)

	auth.POST("/chats/:id/join", h.JoinChat)
	auth.POST("/chats/:id/decline", h.DeclineChatInvite)

	auth.GET("/chats/invites", h.GetChatInvites)

	auth.PUT("/chats/invites/:id", h.RespondToChatInvite)

	auth.GET("/chats/:id/unread", h.GetUnreadCount)

	auth.POST("/chats/:id/read", h.MarkChatAsRead)

	auth.GET("/chats/search", h.SearchChats)
	auth.POST("/chats/:id/join-request", h.SendChatJoinRequest)
	auth.GET("/chats/:id/join-requests", h.GetChatJoinRequests)
	auth.PUT("/chat-join-requests/:id", h.RespondToChatJoinRequest)
	auth.DELETE("/chat-join-requests/:id", h.DeleteChatJoinRequest)
	auth.GET("/user/chat-join-requests", h.GetUserChatJoinRequests)
	auth.PUT("/chats/:id/visibility", h.UpdateChatVisibility)
	auth.POST("/chats/:chat_id/messages/:message_id/read", h.MarkMessageAsRead)

	auth.DELETE("/chats/:id/leave", h.LeaveChat)
	auth.DELETE("/chats/:id", h.DeleteChat)
	auth.DELETE("/chats/:id/members/:user_id", h.RemoveChatMember)
	e.GET("/ws", h.WebSocket)
	e.Static("/uploads", "uploads")

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))
}

func (h *Handler) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
		}

		userID, err := h.service.ValidateToken(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		c.Set("userID", userID)
		return next(c)
	}
}

func (h *Handler) getUserID(c echo.Context) uint {
	return c.Get("userID").(uint)
}

func (h *Handler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	token, err := h.service.Register(req.Phone, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if token == "" {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "code_sent",
			"status":  "Код подтверждения отправлен",
			"code":    "Проверьте консоль сервера для получения кода",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.service.Login(req.Phone, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) Logout(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if err := h.service.Logout(token); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) LogoutAll(c echo.Context) error {
	userID := h.getUserID(c)
	if err := h.service.LogoutAll(userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetMe(c echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	user, err := h.service.GetCurrentUser(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) AddContact(c echo.Context) error {
	userID := h.getUserID(c)

	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	chat, err := h.service.AddContact(userID, req.Phone)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"chat":    chat,
		"message": "Контакт добавлен",
	})
}

func (h *Handler) SearchUser(c echo.Context) error {
	phone := c.QueryParam("phone")
	if phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "phone parameter is required")
	}

	user, err := h.service.SearchUserByPhone(phone)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	return c.JSON(http.StatusOK, user)
}

func (h *Handler) GetChats(c echo.Context) error {
	userID := h.getUserID(c)
	chats, err := h.service.GetChats(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, chats)
}

func (h *Handler) CreateChat(c echo.Context) error {
	userID := h.getUserID(c)
	var req struct {
		Name         string   `json:"name"`
		Type         string   `json:"type" validate:"required,oneof=private group"`
		MemberPhones []string `json:"member_phones"`
		IsSearchable bool     `json:"is_searchable"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	var chat *models.Chat
	var err error

	if req.Type == "private" {
		if len(req.MemberPhones) != 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "приватный чат требует ровно одного участника")
		}
		member, err := h.service.SearchUserByPhone(req.MemberPhones[0])
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "пользователь не найден")
		}

		if member.ID == userID {
			return echo.NewHTTPError(http.StatusBadRequest, "нельзя создать чат с самим собой")
		}

		chat, err = h.service.CreatePrivateChat(userID, member.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		chat, err = h.service.CreateGroupChat(userID, req.Name, req.MemberPhones, req.IsSearchable)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	fullChat, err := h.service.Repo.GetChatByID(chat.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "не удалось загрузить информацию о чате")
	}

	return c.JSON(http.StatusOK, fullChat)
}

func (h *Handler) GetChat(c echo.Context) error {
	userID := h.getUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	chat, err := h.service.GetChat(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "chat not found")
	}

	isMember, err := h.service.IsChatMember(uint(id), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "not a chat member")
	}

	return c.JSON(http.StatusOK, chat)
}

func (h *Handler) AddChatMember(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	if err := h.service.AddChatMember(uint(chatID), userID, req.Phone); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetMessages(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	limit := 50
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	messages, err := h.service.GetChatMessages(uint(chatID), userID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, messages)
}

func (h *Handler) SendMessage(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	content := c.FormValue("content")

	form, err := c.MultipartForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse form")
	}

	var files []*multipart.FileHeader
	if form.File != nil {
		for _, fileHeaders := range form.File {
			for _, fileHeader := range fileHeaders {
				files = append(files, fileHeader)
			}
		}
	}

	message, err := h.service.SendMessage(uint(chatID), userID, content, files)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, message)
}

func (h *Handler) DeleteMessage(c echo.Context) error {
	userID := h.getUserID(c)

	chatID, err := strconv.Atoi(c.Param("chat_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный ID чата")
	}

	messageID, err := strconv.Atoi(c.Param("message_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный ID сообщения")
	}

	isAdmin, err := h.service.IsChatAdmin(uint(chatID), userID)
	if err != nil {
		isAdmin = false
	}

	if err := h.service.DeleteMessage(uint(chatID), uint(messageID), userID, isAdmin); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) WebSocket(c echo.Context) error {
	token := c.QueryParam("token")
	userID, err := h.service.ValidateToken(token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Токен не действителен")
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	client := &ws.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan models.WSMessage, 256),
	}

	h.hub.Register <- client

	return nil
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	userID := h.getUserID(c)

	var req models.UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.service.UpdateProfile(userID, req.Username); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) JoinChat(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	chat, err := h.service.GetChat(uint(chatID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "chat not found")
	}

	if chat.Type != "group" {
		return echo.NewHTTPError(http.StatusBadRequest, "only group chats are joinable")
	}

	isMember, err := h.service.IsChatMember(uint(chatID), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if isMember {
		return echo.NewHTTPError(http.StatusBadRequest, "already a member")
	}

	if err := h.service.Repo.AddChatMember(uint(chatID), userID, false); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetUnreadCount(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	count, err := h.service.GetUnreadCount(uint(chatID), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

func (h *Handler) MarkChatAsRead(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	isMember, err := h.service.IsChatMember(uint(chatID), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if !isMember {
		return echo.NewHTTPError(http.StatusForbidden, "не являетесь участником чата")
	}

	if err := h.service.MarkChatMessagesAsRead(uint(chatID), userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) VerifyCode(c echo.Context) error {
	var req models.VerifyCodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	token, err := h.service.VerifyCode(req.Phone, req.Code, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func (h *Handler) ResendCode(c echo.Context) error {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	err := h.service.ResendCode(req.Phone)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Код отправлен"})
}

func (h *Handler) DeclineChatInvite(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	invite, err := h.service.Repo.GetChatInvite(uint(chatID), userID)
	if err != nil || invite == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "приглашение не найдено")
	}

	if err := h.service.Repo.UpdateChatInviteStatus(invite.ID, "rejected"); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "не удалось отклонить приглашение")
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetChatInvites(c echo.Context) error {
	userID := h.getUserID(c)
	invites, err := h.service.Repo.GetChatInvitesByUserID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, invites)
}

func (h *Handler) RespondToChatInvite(c echo.Context) error {
	userID := h.getUserID(c)
	inviteID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID приглашения")
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.service.RespondToChatInvite(uint(inviteID), userID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) SearchChats(c echo.Context) error {
	search := c.QueryParam("search")
	chats, err := h.service.SearchPublicChats(search)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, chats)
}

func (h *Handler) SendChatJoinRequest(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	if err := h.service.SendChatJoinRequest(userID, uint(chatID)); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetChatJoinRequests(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	requests, err := h.service.GetChatJoinRequests(uint(chatID), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, requests)
}

func (h *Handler) RespondToChatJoinRequest(c echo.Context) error {
	userID := h.getUserID(c)
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID заявки")
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	if err := h.service.RespondToChatJoinRequest(uint(requestID), userID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetUserChatJoinRequests(c echo.Context) error {
	userID := h.getUserID(c)
	requests, err := h.service.GetUserChatJoinRequests(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, requests)
}

func (h *Handler) UpdateChatVisibility(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	var req struct {
		IsSearchable bool `json:"is_searchable"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	if err := h.service.UpdateChatVisibility(uint(chatID), userID, req.IsSearchable); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteChatJoinRequest(c echo.Context) error {
	userID := h.getUserID(c)
	requestID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID заявки")
	}

	request, err := h.service.Repo.GetChatJoinRequest(uint(requestID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "заявка не найдена")
	}

	if request.UserID != userID {
		return echo.NewHTTPError(http.StatusForbidden, "недостаточно прав")
	}

	if request.Status != "pending" {
		return echo.NewHTTPError(http.StatusBadRequest, "нельзя удалить обработанную заявку")
	}

	if err := h.service.Repo.DeleteChatJoinRequest(uint(requestID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateMessage(c echo.Context) error {
	userID := h.getUserID(c)
	messageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req struct {
		Content string `json:"content" validate:"required,max=5000"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	if err := h.service.UpdateMessage(uint(messageID), userID, req.Content); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) MarkMessageAsRead(c echo.Context) error {
	userID := h.getUserID(c)

	chatID, err := strconv.Atoi(c.Param("chat_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный ID чата")
	}

	messageID, err := strconv.Atoi(c.Param("message_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный ID сообщения")
	}

	if err := h.service.MarkMessageAsRead(uint(chatID), uint(messageID), userID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) EditMessage(c echo.Context) error {
	userID := h.getUserID(c)

	chatID, err := strconv.Atoi(c.Param("chat_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный ID чата")
	}

	messageID, err := strconv.Atoi(c.Param("message_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный ID сообщения")
	}

	var req struct {
		Content string `json:"content" validate:"required,max=5000"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Неверный формат данных")
	}

	updatedMessage, err := h.service.EditMessage(uint(chatID), uint(messageID), userID, req.Content)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, updatedMessage)
}

func (h *Handler) LeaveChat(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	chat, err := h.service.Repo.GetChatByID(uint(chatID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "чат не найден")
	}

	if chat.Type == "private" {
		if err := h.service.Repo.DeleteChat(uint(chatID)); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "ошибка удаления чата")
		}

		members, _ := h.service.Repo.GetChatMembers(uint(chatID))
		for _, member := range members {
			if member.UserID != userID {
				h.hub.SendToUser(member.UserID, models.WSMessage{
					Type: "chat_deleted",
					Data: map[string]interface{}{
						"chat_id":    chatID,
						"deleted_by": userID,
						"timestamp":  time.Now().Unix(),
					},
				})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Приватный чат удален",
		})
	} else {
		if err := h.service.Repo.RemoveChatMember(uint(chatID), userID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "ошибка выхода из чата")
		}

		user, _ := h.service.Repo.GetUserByID(userID)
		systemMessage := &models.Message{
			ChatID:   uint(chatID),
			SenderID: userID,
			Content:  fmt.Sprintf("%s покинул(а) группу", user.Username),
			Type:     "system_user_removed",
		}
		h.service.Repo.CreateMessage(systemMessage)

		for _, member := range chat.Members {
			if member.ID != userID {
				h.hub.SendToUser(member.ID, models.WSMessage{
					Type: "user_left_chat",
					Data: map[string]interface{}{
						"chat_id": chatID,
						"user_id": userID,
						"user": map[string]interface{}{
							"id":       user.ID,
							"phone":    user.Phone,
							"username": user.Username,
						},
						"timestamp": time.Now().Unix(),
					},
				})
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Вы вышли из группы",
		})
	}
}

func (h *Handler) DeleteChat(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	forAll := c.QueryParam("forAll") == "true"

	if forAll {
		if err := h.service.DeletePrivateChat(uint(chatID), userID); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Чат удален",
		})
	}

	return echo.NewHTTPError(http.StatusBadRequest, "Неверный запрос")
}

func (h *Handler) RemoveChatMember(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID чата")
	}

	memberID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный ID пользователя")
	}

	isAdmin, err := h.service.Repo.IsChatAdmin(uint(chatID), userID)
	if err != nil || !isAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "недостаточно прав")
	}

	if uint(memberID) == userID {
		return echo.NewHTTPError(http.StatusBadRequest, "нельзя удалить самого себя")
	}

	chat, err := h.service.Repo.GetChatByID(uint(chatID))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "чат не найден")
	}

	if err := h.service.Repo.RemoveChatMember(uint(chatID), uint(memberID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ошибка удаления участника")
	}

	user, _ := h.service.Repo.GetUserByID(userID)
	removedUser, _ := h.service.Repo.GetUserByID(uint(memberID))

	systemMessage := &models.Message{
		ChatID:   uint(chatID),
		SenderID: userID,
		Content:  fmt.Sprintf("%s удалил(а) %s из группы", user.Username, removedUser.Username),
		Type:     "system_user_removed",
	}
	h.service.Repo.CreateMessage(systemMessage)

	h.hub.SendToUser(uint(memberID), models.WSMessage{
		Type: "removed_from_chat",
		Data: map[string]interface{}{
			"chat_id":    chatID,
			"chat_name":  chat.Name,
			"removed_by": userID,
			"timestamp":  time.Now().Unix(),
		},
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Участник удален",
	})
}
