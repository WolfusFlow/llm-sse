package server

import (
	"encoding/json"
	"net/http"
	"time"

	"llmsse/internal/service"

	"go.uber.org/zap"
)

type Handler struct {
	svc    *service.Service
	logger *zap.Logger
}

type processRequest struct {
	Message string `json:"message"`
}

func (h *Handler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Message == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventChan := make(chan service.StatusEvent)
	go func() {
		defer close(eventChan)
		if err := h.svc.ProcessMessage(ctx, req.Message, eventChan); err != nil {
			h.logger.Error("processing message", zap.Error(err))
			eventChan <- service.StatusEvent{
				Status:  "error",
				Message: err.Error(),
			}
		}
	}()

	enc := json.NewEncoder(w)
	for {
		select {
		case <-ctx.Done():
			h.logger.Warn("client disconnected")
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			w.Write([]byte("data: "))
			if err := enc.Encode(event); err != nil {
				h.logger.Error("encoding SSE", zap.Error(err))
				return
			}
			w.Write([]byte("\n"))
			flusher.Flush()
			time.Sleep(300 * time.Millisecond) // For small throttling
		}
	}
}
