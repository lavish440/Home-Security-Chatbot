package app

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"github.com/lavish440/Home-Security-Chatbot/internal/handlers"
	"github.com/lavish440/Home-Security-Chatbot/internal/middleware"
	"github.com/lavish440/Home-Security-Chatbot/internal/models"
	"github.com/lavish440/Home-Security-Chatbot/internal/services"
)

func New(ctx context.Context, cfg *models.Config) (*fiber.App, error) {
	app := fiber.New()

	aiService, err := services.NewAIService(ctx, cfg.GeminiAPIKey)
	if err != nil {
		return nil, err
	}

	sessionService := services.NewSessionService()

	chatHandler := handlers.NewChatHandler(aiService, sessionService)
	debugHandler := handlers.NewDebugHandler(sessionService)

	app.Use("/", static.New("./static"))
	app.Post("/api/chat", chatHandler.Handle)

	middleware.Register(app, cfg)

	if cfg.EnableDebug {
		app.Get(
			"/api/debug/sessions",
			middleware.BasicAuth(cfg.BasicAuthUser, cfg.BasicAuthPass),
			debugHandler.SessionsDump,
		)
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for range ticker.C {
			sessionService.Cleanup(30 * time.Minute)
		}
	}()

	return app, nil
}
