package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

type ChatSession struct {
	Session  *genai.ChatSession
	LastUsed time.Time
}

var (
	chatSessions = sync.Map{}
	client       *genai.Client
	clientErr    error
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:                 "Home Security Assistant",
		JSONEncoder:             json.Marshal,
		JSONDecoder:             json.Unmarshal,
		EnableTrustedProxyCheck: true,
		ProxyHeader:             fiber.HeaderXForwardedFor,
		TrustedProxies:          []string{os.Getenv("REVERSE_PROXY_IP")},
	})

	allowedOrigins, ok := os.LookupEnv("ORIGIN")
	if !ok {
		log.Println("Warning: ORIGIN environment variable not set. CORS might be too permissive.")
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: "GET,POST",
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestCompression,
	}))

	app.Use(recover.New())

	app.Use(logger.New(logger.Config{
		TimeFormat: "02-Jan-2006 03:04:05 PM",
	}))

	if os.Getenv("ENFORCE_HTTPS") == "true" {
		app.Use(func(c *fiber.Ctx) error {
			if c.Get(fiber.HeaderXForwardedProto) == "http" {
				return c.Redirect(fmt.Sprintf("https://%s%s", c.Hostname(), c.OriginalURL()), fiber.StatusMovedPermanently)
			}
			return c.Next()
		})
	}

	app.Static("/", "./static")

	app.Post("/api/chat", handleChat)

	go cleanupSessions()

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}

	log.Printf("Starting server on port %s", port)
	app.Listen(":" + port)
}

func handleChat(c *fiber.Ctx) error {
	type Request struct {
		Message string `json:"message"`
	}

	req := new(Request)

	if err := c.BodyParser(req); err != nil {
		log.Printf("Error parsing request body from %s: %v", c.IP(), err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ip := c.IP()

	response, err := generateGeminiResponse(ip, req.Message)
	if err != nil {
		log.Printf("Error generating Gemini response for %s: %v", ip, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"response": response})
}

func generateGeminiResponse(ip, userInput string) (string, error) {
	apiKey, ok := os.LookupEnv("GEMINI_API_KEY")
	if !ok {
		return "", fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()

	if client == nil && clientErr == nil {
		client, clientErr = genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if clientErr != nil {
			return "", fmt.Errorf("Error creating AI client: %w", clientErr)
		}
	}

	if clientErr != nil {
		return "", fmt.Errorf("Error creating AI client: %w", clientErr)
	}

	model := client.GenerativeModel("gemini-2.0-flash")

	model.SetTemperature(1)
	model.SetTopK(40)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "text/plain"
	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text("You are a specialized AI assistant for home security systems. Answer the following question about home security. If the question is not related to home security, politely decline to answer and explain that you only answer questions about home security systems, cameras, alarms, sensors, etc. Keep responses concise, informative, and helpful for home owners. If the user asks you to control a home security device, behave as if you have done it.")}}

	session, _ := chatSessions.LoadOrStore(ip, &ChatSession{Session: model.StartChat(), LastUsed: time.Now()})
	cs := session.(*ChatSession)
	cs.LastUsed = time.Now()

	resp, err := cs.Session.SendMessage(ctx, genai.Text(userInput))
	if err != nil {
		log.Printf("Error sending message to Gemini: %v", err)
		return "", fmt.Errorf("Error sending message: %w", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(text), nil
		}
	}

	return "No response generated.", fmt.Errorf("no valid candidates found in response")
}

func cleanupSessions() {
	for {
		time.Sleep(10 * time.Minute)
		now := time.Now()

		chatSessions.Range(func(key, value any) bool {
			cs, ok := value.(*ChatSession)
			if !ok {
				log.Printf("Unexpected value type in chatSessions for key %v", key)
				return true
			}

			if now.Sub(cs.LastUsed) > 30*time.Minute {
				chatSessions.Delete(key)
				log.Printf("Deleted inactive session for IP: %v", key)
			}
			return true
		})
	}
}
