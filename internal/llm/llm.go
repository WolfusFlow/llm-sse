package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Interface interface {
	Call(ctx context.Context, messages []ChatMessage) (string, error)
	Stream(ctx context.Context, messages []ChatMessage) <-chan StreamResult
}

type StreamResult struct {
	Content string
	Err     error
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// Call sends a prompt to the LLM and returns the result
func (c *Client) Call(ctx context.Context, messages []ChatMessage) (string, error) {
	reqBody := ChatCompletionRequest{
		Model:    c.model,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm error [%d]: %s", resp.StatusCode, b)
	}

	var parsed ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty LLM response")
	}

	return parsed.Choices[0].Message.Content, nil
}

func (c *Client) Stream(ctx context.Context, messages []ChatMessage) <-chan StreamResult {
	resultChan := make(chan StreamResult)

	go func() {
		defer close(resultChan)

		reqBody := ChatCompletionRequest{
			Model:    c.model,
			Messages: messages,
			Stream:   true,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			resultChan <- StreamResult{Err: fmt.Errorf("marshal request: %w", err)}
			return
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
		if err != nil {
			resultChan <- StreamResult{Err: fmt.Errorf("create request: %w", err)}
			return
		}

		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			resultChan <- StreamResult{Err: fmt.Errorf("http request: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			resultChan <- StreamResult{Err: fmt.Errorf("llm stream error [%d]: %s", resp.StatusCode, b)}
			return
		}

		// Read line by line from the stream
		decoder := NewStreamingDecoder(resp.Body)
		for {
			select {
			case <-ctx.Done():
				resultChan <- StreamResult{Err: ctx.Err()}
				return
			default:
				content, err := decoder.NextChunk()
				if err == io.EOF {
					return
				}
				if err != nil {
					resultChan <- StreamResult{Err: err}
					return
				}
				if content != "" {
					resultChan <- StreamResult{Content: content}
				}
			}
		}
	}()

	return resultChan
}
