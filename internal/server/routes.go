package server

import (
	"llmsse/internal/service"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func NewRouter(svc *service.Service, logger *zap.Logger) *chi.Mux {
	router := chi.NewRouter()

	h := &Handler{svc: svc, logger: logger}

	router.Post("/api/process", h.ProcessMessage)

	return router
}
