package service

import (
	"context"
	"llmsse/internal/llm"

	"go.uber.org/zap"
)

type StatusEvent struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type Service struct {
	llm    llm.Interface
	logger *zap.Logger
}

func NewService(llmClient llm.Interface, logger *zap.Logger) *Service {
	return &Service{llm: llmClient, logger: logger}
}

func (s *Service) ProcessMessage(ctx context.Context, msg string, stream chan<- StatusEvent) error {
	stream <- StatusEvent{Status: "Sending to LLM 1"}
	res1, err := s.llm.Call(ctx, []llm.ChatMessage{
		{Role: "system", Content: "You are LLM 1"},
		{Role: "user", Content: msg},
	})
	if err != nil {
		return err
	}

	stream <- StatusEvent{Status: "Sending to LLM 2"}
	res2, err := s.llm.Call(ctx, []llm.ChatMessage{
		{Role: "system", Content: "You are LLM 2"},
		{Role: "user", Content: msg},
	})
	if err != nil {
		return err
	}

	stream <- StatusEvent{Status: "Combining via LLM 3"}
	combined, err := s.llm.Call(ctx, []llm.ChatMessage{
		{Role: "system", Content: "You are LLM 3. Combine and summarize the following responses:"},
		{Role: "user", Content: res1 + "\n---\n" + res2},
	})
	if err != nil {
		return err
	}

	stream <- StatusEvent{Status: "Completed", Message: combined}
	return nil
}
