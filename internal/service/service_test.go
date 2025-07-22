package service_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"llmsse/internal/llm"
	"llmsse/internal/service"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestProcessMessage_Success(t *testing.T) {
	mockLLM := llm.NewMockClient()
	logger := zap.NewNop()

	svc := service.NewService(mockLLM, logger)

	eventChan := make(chan service.StatusEvent, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tasks := []service.PromptTask{
		{ID: "llm-1", Prompt: []llm.ChatMessage{{Role: "system", Content: "You are LLM 1"}, {Role: "user", Content: "test input"}}},
		{ID: "llm-2", Prompt: []llm.ChatMessage{{Role: "system", Content: "You are LLM 2"}, {Role: "user", Content: "test input"}}},
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for e := range eventChan {
			t.Logf("Stream Event: %+v", e)
		}
	}()

	err := svc.ProcessMessage(ctx, "msg-123", "conv-abc", tasks, eventChan)
	close(eventChan)
	require.NoError(t, err)

	wg.Wait()
}

func newTestLogger() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	cfg.OutputPaths = []string{"stdout"}
	logger, _ := cfg.Build()
	return logger
}
