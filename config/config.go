package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// WhatsApp Settings
	DBPath      string
	BotPrefix   string
	PairingMode bool
	PhoneNumber string

	// ASCII Web API Integration
	AsciiAPIURL  string
	AsciiAPIKey  string
	AsciiWebURL  string
	AsciiDBPath  string

	// Webhook / HTTP Server
	ServerPort   int
	BotSecret    string
	EnableServer bool
}

func LoadConfig() *Config {
	// Try loading .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] Info: No .env file found or error loading, reading from environment variables")
	}

	cfg := &Config{
		DBPath:       getEnv("DB_PATH", "whatsapp.db"),
		BotPrefix:    getEnv("BOT_PREFIX", "!"),
		PairingMode:  getEnvBool("PAIRING_MODE", false),
		PhoneNumber:  getEnv("PHONE_NUMBER", ""),
		AsciiAPIURL:  getEnv("ASCII_API_URL", "http://localhost:3000/api"),
		AsciiAPIKey:  getEnv("ASCII_API_KEY", "ascii-secret-api-key-2026"),
		AsciiWebURL:  getEnv("ASCII_WEB_URL", "https://ascii.web.id"),
		AsciiDBPath:  getEnv("ASCII_DB_PATH", "../ascii-if/local.db"),
		ServerPort:   getEnvInt("SERVER_PORT", 8080),
		BotSecret:    getEnv("BOT_SECRET", "ascii-secret-bot-key"),
		EnableServer: getEnvBool("ENABLE_SERVER", true),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
