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
	MessageID      string `json:"message_id"`
	ConversationID string `json:"conversation_id,omitempty"`
	Status         Status `json:"status"`
	Source         string `json:"source,omitempty"`
	Message        string `json:"message,omitempty"`
	Final          bool   `json:"final,omitempty"`
}

type Status string

type PromptTask struct {
	ID     string
	Prompt []llm.ChatMessage
}

type LLMResult struct {
	ID      string
	Message string
	Err     error
}

type Service struct {
	llm    llm.Interface
	logger *zap.Logger
}

func NewService(llmClient llm.Interface, logger *zap.Logger) *Service {
	return &Service{llm: llmClient, logger: logger}
}

func (s *Service) ProcessMessage(
	ctx context.Context,
	messageID, conversationID string,
	tasks []PromptTask,
	stream chan<- StatusEvent,
) error {
	llmResults := make(chan LLMResult, len(tasks))
	wg := sync.WaitGroup{}

	for _, task := range tasks {
		wg.Add(1)
		go func(task PromptTask) {
			defer wg.Done()

			if ctx.Err() != nil {
				s.logger.Warn("Skipping task due to cancelled context",
					zap.String("task", task.ID),
					zap.String("message_id", messageID),
					zap.String("conversation_id", conversationID),
				)
				return
			}

			s.logger.Debug("Calling "+task.ID, zap.String("current time", time.Now().Format("3:04:05")))
			stream <- StatusEvent{
				MessageID:      messageID,
				ConversationID: conversationID,
				Status:         Status("Sending to " + task.ID),
				Source:         task.ID,
			}

			res, err := s.llm.Call(ctx, task.Prompt)
			if ctx.Err() != nil {
				s.logger.Warn("Context cancelled during llm.Call",
					zap.String("task", task.ID),
					zap.String("message_id", messageID),
					zap.String("conversation_id", conversationID),
				)
				return
			}
			s.logger.Info("response llm", zap.String("task", task.ID), zap.String("resp", res), zap.Error(err))
			if err != nil {
				llmResults <- LLMResult{
					ID:  task.ID,
					Err: fmt.Errorf("%s failed: %w", task.ID, err),
				}
				return
			}

			llmResults <- LLMResult{
				ID:      task.ID,
				Message: res,
			}
			s.logger.Debug("Called "+task.ID, zap.String("current time", time.Now().Format("3:04:05")))
		}(task)
	}

	go func() {
		wg.Wait()
		close(llmResults)
	}()

	// Build final combined input
	builder := strings.Builder{}
	for res := range llmResults {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if res.Err != nil {
			return fmt.Errorf("failed task %s: %w", res.ID, res.Err)
		}
		builder.WriteString(res.Message)
		builder.WriteString("\n---\n")
	}

	combinedInput := strings.TrimSuffix(builder.String(), "\n---\n")

	s.logger.Debug("Combining LLM 3", zap.String("current time", time.Now().Format("3:04:05")), zap.String("message for llm combined", combinedInput))
	stream <- StatusEvent{
		MessageID:      messageID,
		ConversationID: conversationID,
		Status:         "Combining via LLM 3",
		Source:         "llm-combine",
	}

	resultStream := s.llm.Stream(ctx, []llm.ChatMessage{
		{Role: "system", Content: "You are LLM 3. Combine and summarize the following responses:"},
		{Role: "user", Content: combinedInput},
	})

	for res := range resultStream {
		if ctx.Err() != nil {
			s.logger.Warn("Context cancelled during llm.Stream",
				zap.String("message_id", messageID),
				zap.String("conversation_id", conversationID),
			)
			return ctx.Err()
		}
		if res.Err != nil {
			return res.Err
		}
		stream <- StatusEvent{
			MessageID:      messageID,
			ConversationID: conversationID,
			Status:         "Streaming",
			Source:         "llm-combine",
			Message:        res.Content,
		}
	}

	stream <- StatusEvent{
		MessageID:      messageID,
		ConversationID: conversationID,
		Status:         "Completed",
		Source:         "llm-combine",
		Final:          true,
	}

	s.logger.Debug("Completed via LLM 3", zap.String("current time", time.Now().Format("3:04:05")))
	return nil
}
