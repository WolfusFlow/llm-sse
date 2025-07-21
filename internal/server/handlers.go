package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"llmsse/internal/llm"
	"llmsse/internal/service"

	"go.uber.org/zap"
)

type Handler struct {
	svc    *service.Service
	logger *zap.Logger
}

type processRequest struct {
	Message        string `json:"message"`
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id,omitempty"`
}

func (h *Handler) ProcessMessage(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	var req processRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.Message == "" ||
		req.MessageID == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	eventChan := make(chan service.StatusEvent)
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
	defer cancel()

	go func() {
		defer close(eventChan)

		tasks := createPromptTasks(req.Message)

		if err := h.svc.ProcessMessage(
			r.Context(),
			req.MessageID,
			req.ConversationID,
			tasks,
			eventChan,
		); err != nil {
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
			// time.Sleep(300 * time.Millisecond) // For small throttling
		}
	}
}

func createPromptTasks(message string) []service.PromptTask {
	return []service.PromptTask{
		{
			ID: "llm-1",
			Prompt: []llm.ChatMessage{
				{Role: "system", Content: "You are LLM 1"},
				{Role: "user", Content: message},
			},
		},
		{
			ID: "llm-2",
			Prompt: []llm.ChatMessage{
				{Role: "system", Content: "You are LLM 2"},
				{Role: "user", Content: message},
			},
		},
	}
}
