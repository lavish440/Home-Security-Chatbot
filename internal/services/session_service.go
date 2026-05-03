package services

import (
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
)

type ChatSession struct {
	Session  *genai.ChatSession
	LastUsed time.Time
}

type SessionService struct {
	store sync.Map
}

func NewSessionService() *SessionService {
	return &SessionService{}
}

func (s *SessionService) GetOrCreate(ip string, factory func() *genai.ChatSession) *ChatSession {
	val, _ := s.store.LoadOrStore(ip, &ChatSession{
		Session:  factory(),
		LastUsed: time.Now(),
	})

	cs := val.(*ChatSession)
	cs.LastUsed = time.Now()

	return cs
}

func (s *SessionService) Cleanup(timeout time.Duration) {
	now := time.Now()

	s.store.Range(func(key, value any) bool {
		cs := value.(*ChatSession)

		if now.Sub(cs.LastUsed) > timeout {
			s.store.Delete(key)
		}

		return true
	})
}
