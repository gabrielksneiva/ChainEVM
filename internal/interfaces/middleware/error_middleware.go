package middleware

import (
	"context"
	"fmt"

	pkgerrors "github.com/gabrielksneiva/ChainEVM/pkg/errors"
	"go.uber.org/zap"
)

// ErrorMiddleware middleware para tratamento de erros
type ErrorMiddleware struct {
	logger *zap.Logger
}

// NewErrorMiddleware cria um novo middleware de erro
func NewErrorMiddleware(logger *zap.Logger) *ErrorMiddleware {
	return &ErrorMiddleware{
		logger: logger,
	}
}

// HandleError processa um erro e retorna uma resposta apropriada
func (m *ErrorMiddleware) HandleError(ctx context.Context, err error) (string, int, string) {
	if err == nil {
		return "SUCCESS", 200, ""
	}

	appErr, ok := err.(*pkgerrors.AppError)
	if !ok {
		m.logger.Error("unknown error type",
			zap.Error(err))
		return "ERROR", 500, fmt.Sprintf("internal server error: %v", err)
	}

	switch appErr.Code {
	case pkgerrors.ErrValidationFailed.Code:
		return "VALIDATION_ERROR", 400, appErr.Message

	case pkgerrors.ErrChainNotSupported.Code:
		return "CHAIN_NOT_SUPPORTED", 400, appErr.Message

	case pkgerrors.ErrRPCFailed.Code:
		return "RPC_ERROR", 502, appErr.Message

	case pkgerrors.ErrDatabaseError.Code:
		return "DATABASE_ERROR", 500, appErr.Message

	default:
		return "ERROR", 500, appErr.Message
	}
}
