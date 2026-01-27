package main

import (
	"fmt"
	"log"
	"os"
	"webchat/internal/config"
	"webchat/internal/handler"
	"webchat/internal/models"
	"webchat/internal/repository"
	"webchat/internal/service"
	"webchat/internal/ws"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.FriendRequest{},
		&models.Friend{},
		&models.Chat{},
		&models.ChatMember{},
		&models.Message{},
		&models.MessageFile{},
		&models.UserSession{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	repo := repository.NewRepository(db)
	svc := service.NewService(repo, &config.Config{
		UploadDir:   cfg.UploadDir,
		MaxFileSize: cfg.MaxFileSize,
	}, cfg.JWTSecret)

	hub := ws.NewHub()
	go hub.Run()

	h := handler.NewHandler(svc, hub)

	e := echo.New()
	h.RegisterEndpoints(e)

	fmt.Printf("Server starting on port %s\n", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
