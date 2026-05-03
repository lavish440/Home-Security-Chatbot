package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/lavish440/Home-Security-Chatbot/internal/services"
)

type DebugHandler struct {
	Sessions *services.SessionService
}

func NewDebugHandler(s *services.SessionService) *DebugHandler {
	return &DebugHandler{Sessions: s}
}

func (h *DebugHandler) SessionsDump(c *fiber.Ctx) error {
	return c.JSON(h.Sessions.Dump())
}
