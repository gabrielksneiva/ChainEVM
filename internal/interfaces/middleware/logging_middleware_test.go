package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLoggingMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	require.NotNil(t, middleware)
}

func TestLogRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	ctx := context.Background()

	// Não deve panic
	middleware.LogRequest(ctx, "op-123", "ETHEREUM")
}

func TestLogResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	ctx := context.Background()

	// Não deve panic
	middleware.LogResponse(ctx, "op-123", "SUCCESS", nil)
}

func TestLogRequestMultipleChains(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	ctx := context.Background()

	chains := []string{"ETHEREUM", "POLYGON", "BSC", "ARBITRUM", "OPTIMISM"}

	for _, chain := range chains {
		t.Run("log_request_"+chain, func(t *testing.T) {
			middleware.LogRequest(ctx, "op-"+chain, chain)
		})
	}
}

func TestLogResponseWithErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	ctx := context.Background()

	t.Run("log response with error", func(t *testing.T) {
		err := errors.New("test error")
		middleware.LogResponse(ctx, "op-error", "FAILED", err)
	})

	t.Run("log response success", func(t *testing.T) {
		middleware.LogResponse(ctx, "op-success", "SUCCESS", nil)
	})

	t.Run("log response pending", func(t *testing.T) {
		middleware.LogResponse(ctx, "op-pending", "PENDING", nil)
	})
}

func TestLogRequestWithContextValues(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	middleware.LogRequest(ctx, "op-ctx", "ETHEREUM")
}

func TestLoggingMiddlewareSequence(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewLoggingMiddleware(logger)

	ctx := context.Background()

	// Log request
	middleware.LogRequest(ctx, "op-seq", "ETHEREUM")

	// Log response success
	middleware.LogResponse(ctx, "op-seq", "SUCCESS", nil)

	// Log another request
	middleware.LogRequest(ctx, "op-seq-2", "POLYGON")

	// Log response failure
	middleware.LogResponse(ctx, "op-seq-2", "FAILED", nil)
}
