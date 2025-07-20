package service

import "go.uber.org/zap"

type Service struct {
	logger *zap.Logger
}

func NewService(llmClient interface{}, logger *zap.Logger) *Service {
	return &Service{logger: logger}
}
