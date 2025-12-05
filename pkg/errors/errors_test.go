package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppError(t *testing.T) {
	t.Parallel()

	t.Run("create error without underlying error", func(t *testing.T) {
		t.Parallel()

		err := NewAppError("TEST_CODE", "test message", nil)

		require.NotNil(t, err)
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, "test message", err.Message)
		assert.Nil(t, err.Err)
		assert.Equal(t, "TEST_CODE: test message", err.Error())
	})

	t.Run("create error with underlying error", func(t *testing.T) {
		t.Parallel()

		underlyingErr := errors.New("underlying error")
		err := NewAppError("TEST_CODE", "test message", underlyingErr)

		require.NotNil(t, err)
		assert.Equal(t, "TEST_CODE", err.Code)
		assert.Equal(t, "test message", err.Message)
		assert.Equal(t, underlyingErr, err.Err)
		assert.Contains(t, err.Error(), "TEST_CODE")
		assert.Contains(t, err.Error(), "test message")
		assert.Contains(t, err.Error(), "underlying error")
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *AppError
		wantCode string
	}{
		{"ErrInvalidInput", ErrInvalidInput, "INVALID_INPUT"},
		{"ErrValidationFailed", ErrValidationFailed, "VALIDATION_FAILED"},
		{"ErrRPCFailed", ErrRPCFailed, "RPC_FAILED"},
		{"ErrTransactionFailed", ErrTransactionFailed, "TRANSACTION_FAILED"},
		{"ErrChainNotSupported", ErrChainNotSupported, "CHAIN_NOT_SUPPORTED"},
		{"ErrOperationNotFound", ErrOperationNotFound, "OPERATION_NOT_FOUND"},
		{"ErrNotImplemented", ErrNotImplemented, "NOT_IMPLEMENTED"},
		{"ErrDatabaseError", ErrDatabaseError, "DATABASE_ERROR"},
		{"ErrSQSError", ErrSQSError, "SQS_ERROR"},
		{"ErrGasEstimationFailed", ErrGasEstimationFailed, "GAS_ESTIMATION_FAILED"},
		{"ErrInsufficientFunds", ErrInsufficientFunds, "INSUFFICIENT_FUNDS"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.NotEmpty(t, tt.err.Message)
			assert.Contains(t, tt.err.Error(), tt.wantCode)
		})
	}
}

func TestAppErrorError(t *testing.T) {
	t.Parallel()

	t.Run("error message format without underlying error", func(t *testing.T) {
		t.Parallel()

		err := &AppError{
			Code:    "TEST",
			Message: "test message",
			Err:     nil,
		}

		assert.Equal(t, "TEST: test message", err.Error())
	})

	t.Run("error message format with underlying error", func(t *testing.T) {
		t.Parallel()

		err := &AppError{
			Code:    "TEST",
			Message: "test message",
			Err:     errors.New("root cause"),
		}

		errMsg := err.Error()
		assert.Contains(t, errMsg, "TEST")
		assert.Contains(t, errMsg, "test message")
		assert.Contains(t, errMsg, "root cause")
	})
}
