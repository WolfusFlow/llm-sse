package service

import (
	"context"
	"fmt"
	"llmsse/internal/llm"
	"strings"
	"sync"
	"time"

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

type LLMResult struct {
	message string
	err     error
}

func NewService(llmClient llm.Interface, logger *zap.Logger) *Service {
	return &Service{llm: llmClient, logger: logger}
}

func (s *Service) ProcessMessage(ctx context.Context, msg string, stream chan<- StatusEvent) error {
	llmResults := make(chan LLMResult, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		s.logger.Debug("Calling LLM 1", zap.String("current time", time.Now().Format("3:04:05")))
		stream <- StatusEvent{Status: "Sending to LLM 1"}
		res1, err := s.llm.Call(ctx, []llm.ChatMessage{
			{Role: "system", Content: "You are LLM 1"},
			{Role: "user", Content: msg},
		})
		if err != nil {
			llmResults <- LLMResult{
				err: fmt.Errorf("LLM 1 failed: %w", err),
			}
			return
		}
		llmResults <- LLMResult{
			message: res1,
			err:     nil,
		}
		s.logger.Debug("Called LLM 1", zap.String("current time", time.Now().Format("3:04:05")))
	}()

	go func() {
		defer wg.Done()
		s.logger.Debug("Calling LLM 2", zap.String("current time", time.Now().Format("3:04:05")))
		stream <- StatusEvent{Status: "Sending to LLM 2"}
		res2, err := s.llm.Call(ctx, []llm.ChatMessage{
			{Role: "system", Content: "You are LLM 2"},
			{Role: "user", Content: msg},
		})
		if err != nil {
			llmResults <- LLMResult{
				err: fmt.Errorf("LLM 2 failed: %w", err),
			}
			return
		}
		llmResults <- LLMResult{
			message: res2,
			err:     nil,
		}
		s.logger.Debug("Called LLM 2", zap.String("current time", time.Now().Format("3:04:05")))
	}()

	go func() {
		wg.Wait()
		close(llmResults)
	}()

	builder := strings.Builder{}
	for res := range llmResults {
		if res.err != nil {
			return res.err
		}
		builder.WriteString(res.message)
		builder.WriteString("\n---\n")
	}

	combinedInput := builder.String()
	combinedInput = strings.TrimSuffix(combinedInput, "\n---\n")

	s.logger.Debug("Combining LLM 3", zap.String("current time", time.Now().Format("3:04:05")))

	stream <- StatusEvent{Status: "Combining via LLM 3"}
	combined, err := s.llm.Call(ctx, []llm.ChatMessage{
		{Role: "system", Content: "You are LLM 3. Combine and summarize the following responses:"},
		{Role: "user", Content: combinedInput}, //res1 + "\n---\n" + res2
	})
	if err != nil {
		return err
	}
	s.logger.Debug("Combined LLM 3", zap.String("current time", time.Now().Format("3:04:05")))

	stream <- StatusEvent{Status: "Completed", Message: combined}
	return nil
}
