package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lavish440/Home-Security-Chatbot/internal/app"
	"github.com/lavish440/Home-Security-Chatbot/internal/config"
)

func main() {
	cfg := config.Config

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	appInstance, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		sig := <-sigChan
		log.Printf("Received signal: %v. Shutting down...", sig)

		if err := appInstance.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	log.Printf("Server starting on port %s", cfg.Port)

	if err := appInstance.Listen("localhost:" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
