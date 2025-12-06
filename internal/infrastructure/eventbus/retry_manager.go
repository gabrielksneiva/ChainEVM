package eventbus

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ProcessorFunc tipo para função que processa mensagens
type ProcessorFunc func(ctx context.Context) error

// RetryConfig configuração de retry
type RetryConfig struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig retorna configuração padrão de retry
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryManager gerencia retries e Dead Letter Queue
type RetryManager struct {
	dlqHandler DLQSender
	config     RetryConfig
	logger     *zap.Logger
}

// NewRetryManager cria um novo gerenciador de retry
func NewRetryManager(dlqHandler DLQSender, maxRetries int, logger *zap.Logger) *RetryManager {
	config := DefaultRetryConfig()
	config.MaxRetries = maxRetries
	return &RetryManager{
		dlqHandler: dlqHandler,
		config:     config,
		logger:     logger,
	}
}

// ProcessWithRetry executa processamento com retry automático
func (rm *RetryManager) ProcessWithRetry(
	ctx context.Context,
	message *Message,
	processor ProcessorFunc,
) error {
	backoff := rm.config.InitialBackoff

	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		// Verificar se contexto foi cancelado
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		// Tentar processar
		err := processor(ctx)
		if err == nil {
			// Sucesso
			rm.logger.Info("message processed successfully",
				zap.String("operation_id", message.OperationID),
				zap.Int("attempt", attempt+1))
			return nil
		}

		// Se for última tentativa, enviar para DLQ
		if attempt == rm.config.MaxRetries {
			failureReason := fmt.Sprintf("max retries exceeded: %v", err)
			rm.logger.Error("max retries exceeded, sending to DLQ",
				zap.String("operation_id", message.OperationID),
				zap.Int("total_attempts", attempt+1),
				zap.Error(err))

			dlqErr := rm.dlqHandler.SendMessage(ctx, message, failureReason)
			if dlqErr != nil {
				return fmt.Errorf("failed to send to DLQ: %w", dlqErr)
			}

			return fmt.Errorf("max retries exceeded: %v", err)
		}

		// Log de retry
		rm.logger.Warn("processing failed, will retry",
			zap.String("operation_id", message.OperationID),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", rm.config.MaxRetries),
			zap.Duration("next_backoff", backoff),
			zap.Error(err))

		// Aguardar com backoff exponencial
		select {
		case <-time.After(backoff):
			// Continuar
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		}

		// Aumentar backoff exponencialmente, mas não exceder máximo
		backoff = time.Duration(float64(backoff) * rm.config.BackoffMultiplier)
		if backoff > rm.config.MaxBackoff {
			backoff = rm.config.MaxBackoff
		}
	}

	return fmt.Errorf("unexpected error: retry loop exited without result")
}

// GetRetryConfig retorna configuração atual de retry
func (rm *RetryManager) GetRetryConfig() RetryConfig {
	return rm.config
}
