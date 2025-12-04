package errors

import "fmt"

// AppError erro customizado da aplicação
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError cria um novo erro da aplicação
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Erros comuns
var (
	ErrInvalidInput        = &AppError{Code: "INVALID_INPUT", Message: "invalid input"}
	ErrValidationFailed    = &AppError{Code: "VALIDATION_FAILED", Message: "validation failed"}
	ErrRPCFailed           = &AppError{Code: "RPC_FAILED", Message: "RPC call failed"}
	ErrTransactionFailed   = &AppError{Code: "TRANSACTION_FAILED", Message: "transaction execution failed"}
	ErrChainNotSupported   = &AppError{Code: "CHAIN_NOT_SUPPORTED", Message: "chain type not supported"}
	ErrOperationNotFound   = &AppError{Code: "OPERATION_NOT_FOUND", Message: "operation not found"}
	ErrNotImplemented      = &AppError{Code: "NOT_IMPLEMENTED", Message: "feature not implemented"}
	ErrDatabaseError       = &AppError{Code: "DATABASE_ERROR", Message: "database error"}
	ErrSQSError            = &AppError{Code: "SQS_ERROR", Message: "SQS error"}
	ErrGasEstimationFailed = &AppError{Code: "GAS_ESTIMATION_FAILED", Message: "gas estimation failed"}
	ErrInsufficientFunds   = &AppError{Code: "INSUFFICIENT_FUNDS", Message: "insufficient funds for transaction"}
)
