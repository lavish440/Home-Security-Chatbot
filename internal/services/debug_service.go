package services

import "github.com/google/generative-ai-go/genai"

func (s *SessionService) Dump() map[string][]map[string]string {
	result := make(map[string][]map[string]string)

	s.store.Range(func(key, value any) bool {
		cs := value.(*ChatSession)

		history := []map[string]string{}

		for _, msg := range cs.Session.History {
			for _, part := range msg.Parts {
				if text, ok := part.(genai.Text); ok {
					history = append(history, map[string]string{
						"role": msg.Role,
						"text": string(text),
					})
				}
			}
		}

		result[key.(string)] = history
		return true
	})

	return result
}
