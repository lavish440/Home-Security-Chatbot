package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/lavish440/Home-Security-Chatbot/internal/models"
	"github.com/lavish440/Home-Security-Chatbot/internal/services"
)

type ChatHandler struct {
	AI       *services.AIService
	Sessions *services.SessionService
}

func NewChatHandler(ai *services.AIService, sessions *services.SessionService) *ChatHandler {
	return &ChatHandler{AI: ai, Sessions: sessions}
}

func (h *ChatHandler) Handle(c *fiber.Ctx) error {
	var req models.ChatMessageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	ip := c.IP()
	if ip == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid IP"})
	}

	cs := h.Sessions.GetOrCreate(ip, h.AI.StartChat)

	resp, err := h.AI.Send(context.Background(), cs.Session, req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"response": resp})
}
