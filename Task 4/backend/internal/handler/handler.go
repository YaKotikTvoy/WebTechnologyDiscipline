package handler

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
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

	//auth.POST("/auth/logout-all", h.LogoutAll) // Нафиг я вообще возился с этой логикой
	/*
		auth.GET("/friends/requests", h.GetFriendRequests)
		auth.POST("/friends/requests", h.SendFriendRequest)
		auth.PUT("/friends/requests/:id", h.RespondToFriendRequest)
		auth.GET("/friends", h.GetFriends)
		auth.DELETE("/friends/:id", h.RemoveFriend)
	*/
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
		return c.JSON(http.StatusOK, map[string]string{
			"message": "code_sent",
			"status":  "Код подтверждения отправлен в консоль",
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
		chat, err = h.service.CreatePrivateChat(userID, member.ID)
		if err != nil {
			if strings.Contains(err.Error(), "запрос в друзья отправлен") {
				return echo.NewHTTPError(http.StatusOK, map[string]interface{}{
					"message":         "Запрос в друзья отправлен",
					"need_friendship": true,
				})
			}
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	} else {
		chat, err = h.service.CreateGroupChat(userID, req.Name, req.MemberPhones, req.IsSearchable)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, chat)
}

func (h *Handler) GetChat(c echo.Context) error {
	userID := h.getUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	chat, err := h.service.GetChat(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
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

func (h *Handler) RemoveChatMember(c echo.Context) error {
	userID := h.getUserID(c)
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid chat id")
	}

	memberID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	if err := h.service.RemoveChatMember(uint(chatID), userID, uint(memberID)); err != nil {
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

	sender, err := h.service.Repo.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get sender info")
	}

	chat, err := h.service.GetChat(uint(chatID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get chat info")
	}

	for _, member := range chat.Members {
		if member.ID != userID {
			unreadCount, _ := h.service.GetUnreadCount(uint(chatID), member.ID)

			h.hub.SendToUser(member.ID, models.WSMessage{
				Type: "new_message",
				Data: map[string]interface{}{
					"chat_id":   chatID,
					"chatName":  chat.Name,
					"chat_type": chat.Type,
					"message":   message,
					"sender": map[string]interface{}{
						"id":       sender.ID,
						"phone":    sender.Phone,
						"username": sender.Username,
					},
					"unread_count": unreadCount,
					"timestamp":    time.Now().Unix(),
				},
			})
		}
	}

	return c.JSON(http.StatusOK, message)
}

func (h *Handler) DeleteMessage(c echo.Context) error {
	userID := h.getUserID(c)
	messageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.service.DeleteMessage(uint(messageID), userID); err != nil {
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

/*
func (h *Handler) GetFriendRequests(c echo.Context) error {
	userID := h.getUserID(c)
	requests, err := h.service.GetFriendRequests(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, requests)
}

func (h *Handler) RespondToFriendRequest(c echo.Context) error {
	userID := h.getUserID(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Status != "accepted" && req.Status != "rejected" {
		return echo.NewHTTPError(http.StatusBadRequest, "status must be 'accepted' or 'rejected'")
	}

	request, err := h.service.Repo.GetFriendRequestByID(uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "friend request not found")
	}

	if err := h.service.RespondToFriendRequest(uint(id), userID, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	recipient, _ := h.service.Repo.GetUserByID(userID)

	h.hub.SendToUser(request.SenderID, models.WSMessage{
		Type: "friend_request_responded",
		Data: map[string]interface{}{
			"request_id": request.ID,
			"status":     req.Status,
			"recipient": map[string]interface{}{
				"id":       recipient.ID,
				"phone":    recipient.Phone,
				"username": recipient.Username,
			},
		},
	})

	return c.NoContent(http.StatusOK)
}
*/
//func (h *Handler) GetFriends(c echo.Context) error {
//	userID := h.getUserID(c)
//	friends, err := h.service.GetFriends(userID)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
//	}
//	return c.JSON(http.StatusOK, friends)
//}

/*func (h *Handler) SendFriendRequest(c echo.Context) error {
	userID := h.getUserID(c)
	var req models.FriendRequestInput
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "неверный формат данных")
	}

	request, err := h.service.SendFriendRequest(userID, req.RecipientPhone)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sender, _ := h.service.Repo.GetUserByID(userID)

	h.hub.SendToUser(request.RecipientID, models.WSMessage{
		Type: "friend_request",
		Data: map[string]interface{}{
			"request_id": request.ID,
			"sender": map[string]interface{}{
				"id":       sender.ID,
				"phone":    sender.Phone,
				"username": sender.Username,
			},
			"message": fmt.Sprintf("%s (%s) хочет добавить вас в друзья",
				sender.Username, sender.Phone),
		},
	})

	return c.JSON(http.StatusOK, request)


}

func (h *Handler) RemoveFriend(c echo.Context) error {
	userID := h.getUserID(c)
	friendID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.service.RemoveFriend(userID, uint(friendID)); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
*/
