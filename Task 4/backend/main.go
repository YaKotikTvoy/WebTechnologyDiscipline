package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"webchat/internal/config"
	"webchat/internal/handlers"
	"webchat/internal/middleware/jwt"
	"webchat/pkg/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db_sql, err := database.NewPostgresConnection(cfg.DBURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	var db *database.DB = &database.DB{DB: db_sql}
	defer db.Close()

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8080"},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization, echo.HeaderXRequestedWith},
		AllowCredentials: true,
	}))

	e.Static("/uploads", "uploads")

	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	userHandler := handlers.NewUserHandler(db)
	chatHandler := handlers.NewChatHandler(db)
	messageHandler := handlers.NewMessageHandler(db)
	profileHandler := handlers.NewProfileHandler(db, cfg.JWTSecret)

	api := e.Group("/api")

	api.POST("/register", authHandler.Register)
	api.POST("/login", authHandler.Login)

	api.GET("/public-chats", chatHandler.GetPublicChats)
	api.GET("/public-chats/:id", chatHandler.GetPublicChat)

	authorized := api.Group("")
	authorized.Use(jwt.JWTMiddleware(cfg.JWTSecret))

	authorized.GET("/profile", profileHandler.GetProfile)
	authorized.PUT("/profile", profileHandler.UpdateProfile)
	authorized.POST("/profile/delete-request", profileHandler.RequestDelete)
	authorized.POST("/profile/confirm-delete", profileHandler.ConfirmDelete)

	authorized.GET("/chats", chatHandler.GetUserChats)
	authorized.POST("/chats", chatHandler.CreateChat)
	authorized.GET("/chats/:id", chatHandler.GetChat)
	authorized.PUT("/chats/:id", chatHandler.UpdateChat)
	authorized.POST("/chats/:id/invite", chatHandler.CreateInvite)
	authorized.POST("/chats/join/:code", chatHandler.JoinChat)
	authorized.POST("/chats/:id/members/:userId/role", chatHandler.AssignRole)
	authorized.GET("/chats/:id/members", chatHandler.GetChatMembers)
	authorized.DELETE("/chats/:id/members/:userId", chatHandler.RemoveMember)

	authorized.GET("/messages", messageHandler.GetMessages)
	authorized.POST("/messages", messageHandler.SendMessage)
	authorized.PUT("/messages/:id", messageHandler.EditMessage)
	authorized.DELETE("/messages/:id", messageHandler.DeleteMessage)
	authorized.POST("/messages/upload", messageHandler.UploadFile)

	authorized.GET("/contacts", userHandler.GetContacts)
	authorized.POST("/contacts", userHandler.AddContact)
	authorized.DELETE("/contacts/:id", userHandler.RemoveContact)

	authorized.POST("/blacklist", userHandler.AddToBlacklist)
	authorized.DELETE("/blacklist/:id", userHandler.RemoveFromBlacklist)
	authorized.GET("/blacklist", userHandler.GetBlacklist)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}
