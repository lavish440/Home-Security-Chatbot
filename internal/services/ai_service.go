package services

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

func NewAIService(ctx context.Context, apiKey string) (*AIService, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	model := client.GenerativeModel("gemini-flash-latest")

	model.SetTemperature(0.7)
	model.SetTopK(40)
	model.SetTopP(0.9)
	model.SetMaxOutputTokens(2048)

	model.ResponseMIMEType = "text/plain"
	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text("You are a specialized AI assistant for home security systems. Answer the following question about home security. If the question is not related to home security, politely decline to answer and explain that you only answer questions about home security systems, cameras, alarms, sensors, etc. Keep responses concise, informative, and helpful for home owners. If the user asks you to control a home security device, behave as if you have done it.")}}

	return &AIService{
		client: client,
		model:  model,
	}, nil
}

func (s *AIService) StartChat() *genai.ChatSession {
	return s.model.StartChat()
}

func (s *AIService) Send(ctx context.Context, session *genai.ChatSession, msg string) (string, error) {
	resp, err := session.SendMessage(ctx, genai.Text(msg))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 {
		return "", nil
	}

	part := resp.Candidates[0].Content.Parts[0]
	if text, ok := part.(genai.Text); ok {
		return string(text), nil
	}

	return "", nil
}
