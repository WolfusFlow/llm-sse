package main

import (
	"context"
	"llmsse/internal/app"
	"llmsse/internal/config"
	"llmsse/internal/logger"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logr, err := logger.New(cfg.LogLevel, cfg.Production)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logr.Sync()

	application := app.New(cfg, logr)

	go func() {
		if err := application.Run(); err != nil {
			logr.Fatal("Server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := application.Shutdown(ctx); err != nil {
		logr.Fatal("Shutdown error", zap.Error(err))
	}
}
