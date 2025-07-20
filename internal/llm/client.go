package llm

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

func NewClient(apiKey, model, baseURL string, logger *zap.Logger) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		logger:     logger,
	}
}
