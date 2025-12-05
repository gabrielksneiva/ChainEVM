package middleware

import (
	"context"

	"go.uber.org/zap"
)

// LoggingMiddleware middleware para logging de requisições
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware cria um novo middleware de logging
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// LogRequest registra detalhes de uma requisição
func (m *LoggingMiddleware) LogRequest(ctx context.Context, operationID string, chainType string) {
	m.logger.Info("request received",
		zap.String("operation_id", operationID),
		zap.String("chain_type", chainType))
}

// LogResponse registra detalhes de uma resposta
func (m *LoggingMiddleware) LogResponse(ctx context.Context, operationID string, status string, err error) {
	if err != nil {
		m.logger.Error("request failed",
			zap.String("operation_id", operationID),
			zap.String("status", status),
			zap.Error(err))
	} else {
		m.logger.Info("request completed",
			zap.String("operation_id", operationID),
			zap.String("status", status))
	}
}
