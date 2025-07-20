package server

import (
	"net/http"

	"llmsse/internal/service"

	"go.uber.org/zap"
)

type Handler struct {
	svc    *service.Service
	logger *zap.Logger
}

func (h *Handler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from ProcessMessage handler"))
}
