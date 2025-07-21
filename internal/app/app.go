package app

import (
	"context"
	"llmsse/internal/config"
	"llmsse/internal/llm"
	"llmsse/internal/server"
	"llmsse/internal/service"

	"go.uber.org/zap"
)

type App struct {
	Server *server.Server
	Logger *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) *App {
	var llmClient llm.Interface
	if cfg.UseMockLLM {
		logger.Info("Using mock LLM client")
		llmClient = llm.NewMockClient()
	} else {
		logger.Info("Using real OpenAI LLM client")
		llmClient = llm.NewClient(cfg.LLMKey, "gpt-4o", "https://api.openai.com", logger)
	}

	svc := service.NewService(llmClient, logger)
	router := server.NewRouter(svc, logger)
	srv := server.NewServer(cfg.HTTPAddr, router, logger)

	return &App{
		Server: srv,
		Logger: logger,
	}
}

func (a *App) Run() error {
	a.Logger.Info("Starting server...")
	return a.Server.Run()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Info("Shutting down server...")
	return a.Server.Shutdown(ctx)
}
