package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
)

func BasicAuth(user, pass string) func(fiber.Ctx) error {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			user: pass,
		},
	})
}
