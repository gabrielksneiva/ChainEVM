package middleware

import (
	"context"
	"testing"

	pkgerrors "github.com/gabrielksneiva/ChainEVM/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewErrorMiddleware(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	require.NotNil(t, middleware)
}

func TestHandleErrorSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	status, code, msg := middleware.HandleError(context.Background(), nil)

	assert.Equal(t, "SUCCESS", status)
	assert.Equal(t, 200, code)
	assert.Equal(t, "", msg)
}

func TestHandleErrorValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		pkgerrors.ErrValidationFailed.Code,
		"invalid input",
		nil,
	)

	status, code, _ := middleware.HandleError(context.Background(), err)

	assert.Equal(t, "VALIDATION_ERROR", status)
	assert.Equal(t, 400, code)
}

func TestHandleErrorRPC(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		pkgerrors.ErrRPCFailed.Code,
		"RPC timeout",
		nil,
	)

	status, code, _ := middleware.HandleError(context.Background(), err)

	assert.Equal(t, "RPC_ERROR", status)
	assert.Equal(t, 502, code)
}

func TestHandleErrorDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		pkgerrors.ErrDatabaseError.Code,
		"database connection error",
		nil,
	)

	status, code, _ := middleware.HandleError(context.Background(), err)

	assert.Equal(t, "DATABASE_ERROR", status)
	assert.Equal(t, 500, code)
}

func TestHandleErrorUnknown(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		"UNKNOWN_ERROR",
		"something went wrong",
		nil,
	)

	status, code, _ := middleware.HandleError(context.Background(), err)

	assert.Equal(t, "ERROR", status)
	assert.Equal(t, 500, code)
}

func TestHandleErrorOperationNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		pkgerrors.ErrOperationNotFound.Code,
		"resource not found",
		nil,
	)

	status, code, _ := middleware.HandleError(context.Background(), err)

	// ErrOperationNotFound not explicitly handled, defaults to ERROR/500
	assert.Equal(t, "ERROR", status)
	assert.Equal(t, 500, code)
}

func TestHandleErrorInvalidInput(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		pkgerrors.ErrInvalidInput.Code,
		"invalid request",
		nil,
	)

	status, code, _ := middleware.HandleError(context.Background(), err)

	// ErrInvalidInput not explicitly handled, defaults to ERROR/500
	assert.Equal(t, "ERROR", status)
	assert.Equal(t, 500, code)
}

func TestHandleErrorWithContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	ctx := context.WithValue(context.Background(), "operation_id", "op-123")
	err := pkgerrors.NewAppError(
		pkgerrors.ErrValidationFailed.Code,
		"validation failed",
		nil,
	)

	status, code, msg := middleware.HandleError(ctx, err)

	assert.Equal(t, "VALIDATION_ERROR", status)
	assert.Equal(t, 400, code)
	assert.NotEmpty(t, msg)
}

func TestHandleErrorMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	errorMsg := "custom error message"
	err := pkgerrors.NewAppError(
		pkgerrors.ErrRPCFailed.Code,
		errorMsg,
		nil,
	)

	_, _, msg := middleware.HandleError(context.Background(), err)

	assert.Equal(t, errorMsg, msg)
}

func TestHandleErrorChainNotSupported(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		pkgerrors.ErrChainNotSupported.Code,
		"chain not supported",
		nil,
	)

	status, code, msg := middleware.HandleError(context.Background(), err)

	assert.Equal(t, "CHAIN_NOT_SUPPORTED", status)
	assert.Equal(t, 400, code)
	assert.Equal(t, "chain not supported", msg)
}

func TestHandleErrorDefaultCase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	err := pkgerrors.NewAppError(
		"UNKNOWN_ERROR_CODE",
		"some unknown error",
		nil,
	)

	status, code, msg := middleware.HandleError(context.Background(), err)

	assert.Equal(t, "ERROR", status)
	assert.Equal(t, 500, code)
	assert.Equal(t, "some unknown error", msg)
}

func TestHandleErrorContextPropagation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	ctx := context.WithValue(context.Background(), "request_id", "12345")

	err := pkgerrors.NewAppError(
		pkgerrors.ErrDatabaseError.Code,
		"db connection failed",
		nil,
	)

	status, code, msg := middleware.HandleError(ctx, err)

	assert.Equal(t, "DATABASE_ERROR", status)
	assert.Equal(t, 500, code)
	assert.Equal(t, "db connection failed", msg)
}

func TestHandleErrorAllErrorCodes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	middleware := NewErrorMiddleware(logger)

	tests := []struct {
		name           string
		errorCode      string
		expectedStatus string
		expectedCode   int
	}{
		{
			name:           "validation failed",
			errorCode:      pkgerrors.ErrValidationFailed.Code,
			expectedStatus: "VALIDATION_ERROR",
			expectedCode:   400,
		},
		{
			name:           "rpc failed",
			errorCode:      pkgerrors.ErrRPCFailed.Code,
			expectedStatus: "RPC_ERROR",
			expectedCode:   502,
		},
		{
			name:           "database error",
			errorCode:      pkgerrors.ErrDatabaseError.Code,
			expectedStatus: "DATABASE_ERROR",
			expectedCode:   500,
		},
		{
			name:           "chain not supported",
			errorCode:      pkgerrors.ErrChainNotSupported.Code,
			expectedStatus: "CHAIN_NOT_SUPPORTED",
			expectedCode:   400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pkgerrors.NewAppError(
				tt.errorCode,
				"error message",
				nil,
			)

			status, code, _ := middleware.HandleError(context.Background(), err)

			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedCode, code)
		})
	}
}
