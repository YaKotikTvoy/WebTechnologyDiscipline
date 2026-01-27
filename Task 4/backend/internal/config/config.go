package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        string
	DBHost      string
	DBPort      string
	DBName      string
	DBUser      string
	DBPassword  string
	JWTSecret   string
	JWTExpire   string
	UploadDir   string
	MaxFileSize int64
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBName:      getEnv("DB_NAME", "webchatdb"),
		DBUser:      getEnv("DB_USER", "webchat"),
		DBPassword:  getEnv("DB_PASSWORD", "web1234"),
		JWTSecret:   getEnv("JWT_SECRET", "secret-key-change-in-production"),
		JWTExpire:   getEnv("JWT_EXPIRE", "24h"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", 10485760),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}
