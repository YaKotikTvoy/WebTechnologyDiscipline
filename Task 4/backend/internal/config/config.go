package config

import (
	"os"
)

type Config struct {
	DBURL     string
	JWTSecret string
	Port      string
}

func Load() *Config {
	return &Config{
		DBURL:     getEnv("DB_URL", "user=webchat password=web1234 dbname=webchatdb sslmode=disable"),
		JWTSecret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Port:      getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
