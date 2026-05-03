package config

import (
	"cmp"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"

	"github.com/lavish440/Home-Security-Chatbot/internal/models"
)

var (
	envOnce sync.Once
	Config  *models.Config
)

func init() {
	envOnce.Do(func() {
		if err := godotenv.Load(); err != nil {
			log.Print("No .env file found")
		}

		Config = &models.Config{
			Port:           getEnv("PORT", "3000"),
			GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
			Origin:         os.Getenv("ORIGIN"),
			ReverseProxyIP: os.Getenv("REVERSE_PROXY_IP"),
			BasicAuthUser:  os.Getenv("BASIC_AUTH_USER"),
			BasicAuthPass:  os.Getenv("BASIC_AUTH_PASS"),
		}

		Config.EnforceHTTPS = os.Getenv("ENFORCE_HTTPS") == "true"
		Config.EnableMonitoring = os.Getenv("ENABLE_MONITORING") == "true"
		Config.EnableDebug = os.Getenv("ENABLE_DEBUG_ENDPOINTS") == "true"

		if Config.GeminiAPIKey == "" {
			log.Fatal("GEMINI_API_KEY is required")
		}
	})
}

func getEnv(key, fallback string) string {
	return cmp.Or(os.Getenv(key), fallback)
}
