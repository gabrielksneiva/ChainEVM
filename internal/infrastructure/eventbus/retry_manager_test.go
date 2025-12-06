package eventbus

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// mockDLQHandler para testes de retry
type mockDLQHandler struct {
	mock.Mock
}

func (m *mockDLQHandler) SendMessage(ctx context.Context, message *Message, reason string) error {
	args := m.Called(ctx, message, reason)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return args.Error(0)
}

// TestRetryManager_ProcessWithRetry_Success testa processamento bem-sucedido na primeira tentativa
func TestRetryManager_ProcessWithRetry_Success(t *testing.T) {
	// Arrange
	mockDLQ := new(mockDLQHandler)
	logger := zap.NewNop()
	retryManager := NewRetryManager(mockDLQ, 3, logger)

	processedCalls := 0
	processorFunc := func(ctx context.Context) error {
		processedCalls++
		return nil
	}

	message := &Message{
		OperationID:   "op-123",
		ChainType:     "sepolia",
		OperationType: "TRANSFER",
	}

	// Act
	err := retryManager.ProcessWithRetry(context.Background(), message, processorFunc)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, processedCalls)
}

// TestRetryManager_ProcessWithRetry_RetryAndSucceed testa falha + sucesso na retry
func TestRetryManager_ProcessWithRetry_RetryAndSucceed(t *testing.T) {
	// Arrange
	mockDLQ := new(mockDLQHandler)
	logger := zap.NewNop()
	retryManager := NewRetryManager(mockDLQ, 3, logger)

	processedCalls := 0
	processorFunc := func(ctx context.Context) error {
		processedCalls++
		if processedCalls < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	message := &Message{
		OperationID:   "op-123",
		ChainType:     "sepolia",
		OperationType: "TRANSFER",
	}

	// Act
	err := retryManager.ProcessWithRetry(context.Background(), message, processorFunc)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, processedCalls)
}

// TestRetryManager_ProcessWithRetry_ExhaustedRetries testa esgotamento de retries
func TestRetryManager_ProcessWithRetry_ExhaustedRetries(t *testing.T) {
	// Arrange
	mockDLQ := new(mockDLQHandler)
	logger := zap.NewNop()
	retryManager := NewRetryManager(mockDLQ, 2, logger)

	processorFunc := func(ctx context.Context) error {
		return errors.New("persistent error")
	}

	message := &Message{
		OperationID:   "op-123",
		ChainType:     "sepolia",
		OperationType: "TRANSFER",
	}

	mockDLQ.On("SendMessage", mock.Anything, message, mock.MatchedBy(func(reason string) bool {
		return reason == "max retries exceeded: persistent error"
	})).Return(nil)

	// Act
	err := retryManager.ProcessWithRetry(context.Background(), message, processorFunc)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "max retries exceeded: persistent error", err.Error())
	mockDLQ.AssertExpectations(t)
}

// TestRetryManager_ProcessWithRetry_DLQError testa erro ao enviar para DLQ
func TestRetryManager_ProcessWithRetry_DLQError(t *testing.T) {
	// Arrange
	mockDLQ := new(mockDLQHandler)
	logger := zap.NewNop()
	retryManager := NewRetryManager(mockDLQ, 1, logger)

	processorFunc := func(ctx context.Context) error {
		return errors.New("processing error")
	}

	message := &Message{
		OperationID:   "op-123",
		ChainType:     "sepolia",
		OperationType: "TRANSFER",
	}

	mockDLQ.On("SendMessage", mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("failed to send to DLQ"))

	// Act
	err := retryManager.ProcessWithRetry(context.Background(), message, processorFunc)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send to DLQ")
}

// TestRetryManager_ProcessWithRetry_ContextCancelled testa contexto cancelado
func TestRetryManager_ProcessWithRetry_ContextCancelled(t *testing.T) {
	// Arrange
	mockDLQ := new(mockDLQHandler)
	logger := zap.NewNop()
	retryManager := NewRetryManager(mockDLQ, 3, logger)

	processorFunc := func(ctx context.Context) error {
		return ctx.Err()
	}

	message := &Message{
		OperationID:   "op-123",
		ChainType:     "sepolia",
		OperationType: "TRANSFER",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	err := retryManager.ProcessWithRetry(ctx, message, processorFunc)

	// Assert
	assert.Error(t, err)
}

// TestRetryManager_GetRetryConfig testa obtenção de configuração de retry
func TestRetryManager_GetRetryConfig(t *testing.T) {
	// Arrange
	mockDLQ := new(mockDLQHandler)
	logger := zap.NewNop()
	retryManager := NewRetryManager(mockDLQ, 5, logger)

	// Act
	config := retryManager.GetRetryConfig()

	// Assert
	assert.Equal(t, 5, config.MaxRetries)
}
