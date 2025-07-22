package server

import (
	"llmsse/internal/server/middleware"
	"llmsse/internal/service"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func NewRouter(svc *service.Service, logger *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	rateLimiter := middleware.NewRateLimiter(rate.Every(1*time.Second), 3, 5*time.Minute, logger)

	router.Use(rateLimiter.Middleware)

	h := &Handler{svc: svc, logger: logger}

	router.Post("/api/process", h.ProcessMessage)

	return router
}
