package main

import (
	"llmsse/internal/llm"
	"llmsse/internal/server"
	"llmsse/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// llmClient := llm.NewClient(
	// 	os.Getenv("LLM_KEY"),
	// 	"gpt-4o",
	// 	"https://api.openai.com",
	// 	logger,
	// )

	mockClient := llm.NewMockClient()

	svc := service.NewService(mockClient, logger)
	router := server.NewRouter(svc, logger)
	srv := server.NewServer(":8080", router, logger)

	go func() {
		if err := srv.Run(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	if err := srv.Shutdown(nil); err != nil {
		log.Fatalf("Server Shutdown: %v", err)
	}
}
