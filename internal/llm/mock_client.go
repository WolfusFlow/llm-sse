package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MockClient struct{}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (m *MockClient) Call(ctx context.Context, messages []ChatMessage) (string, error) {
	// Simulate latency
	time.Sleep(1 * time.Second)

	// Simulate different logic based on system prompt
	system := ""
	for _, msg := range messages {
		if msg.Role == "system" {
			system = msg.Content
			break
		}
	}

	switch {
	case system == "You are LLM 1":
		time.Sleep(3 * time.Second)
		return fmt.Sprintf("LLM1 processed: %s", messages[len(messages)-1].Content), nil
	case system == "You are LLM 2":
		time.Sleep(3 * time.Second)
		return fmt.Sprintf("LLM2 processed: %s", messages[len(messages)-1].Content), nil
	case system == "You are LLM 3. Combine and summarize the following responses:":
		return "Combined summary of LLM1 and LLM2", nil
	default:
		return "Mock LLM response", nil
	}
}

func (m *MockClient) Stream(ctx context.Context, messages []ChatMessage) <-chan StreamResult {
	resultChan := make(chan StreamResult)

	go func() {
		defer close(resultChan)

		// Simulate a streaming response by splitting into "tokens"
		response := "Combined summary of LLM1 and LLM2"
		tokens := strings.Split(response, " ")

		for _, token := range tokens {
			select {
			case <-ctx.Done():
				resultChan <- StreamResult{Err: ctx.Err()}
				return
			default:
				resultChan <- StreamResult{Content: token + " "}
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	return resultChan
}
