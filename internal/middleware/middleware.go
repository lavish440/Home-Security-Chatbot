package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"github.com/lavish440/Home-Security-Chatbot/internal/models"
)

func Register(app *fiber.App, cfg *models.Config) {
	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{cfg.Origin},
		AllowMethods: []string{fiber.MethodGet, fiber.MethodPost},
	}))

	// Compression
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	// Panic recovery
	app.Use(recover.New())

	// Logger
	app.Use(logger.New(logger.Config{
		TimeFormat: "02-Jan-2006 03:04:05 PM",
	}))

	if cfg.EnableMonitoring {
		app.Get("/metrics", monitor.New())
	}

	// Rate limiter
	app.Use(limiter.New(limiter.Config{
		Max:        1000,
		Expiration: 30 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests, please try again later.",
			})
		},
	}))

	// HTTPS enforcement (behind proxy)
	if cfg.EnforceHTTPS {
		app.Use(func(c fiber.Ctx) error {
			if c.Get(fiber.HeaderXForwardedProto) == "http" {
				return c.Redirect().Status(fiber.StatusMovedPermanently).To(
					fmt.Sprintf("https://%s%s", c.Hostname(), c.OriginalURL()),
				)
			}
			return c.Next()
		})
	}
}
