package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Errorf("Error loading .env file: %w", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Home Security Assistant",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Serve static files
	app.Static("/", "./static")

	// API endpoints
	app.Post("/api/chat", handleChat)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app.Listen(":" + port)
}

func handleChat(c *fiber.Ctx) error {
	type Request struct {
		Message string `json:"message"`
	}

	req := new(Request)

	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	response, err := generateGeminiResponse(req.Message)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"response": response})
}

func generateGeminiResponse(userInput string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("Error creating AI client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(userInput))
	if err != nil {
		return "", fmt.Errorf("Error generating content: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			ret := string(text)
			ret = strings.ReplaceAll(ret, "***", "\n")
			ret = strings.ReplaceAll(ret, "**", "\n")
			ret = strings.ReplaceAll(ret, "*", "\n")
			return ret, nil
		}
	}
	return "No response generated.", fmt.Errorf("no valid candidates found in response")
}
