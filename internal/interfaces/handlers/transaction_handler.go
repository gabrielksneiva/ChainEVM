package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gabrielksneiva/ChainEVM/internal/application/dtos"
	"github.com/gabrielksneiva/ChainEVM/internal/application/usecases"
	pkgconfig "github.com/gabrielksneiva/ChainEVM/pkg/config"
	"go.uber.org/zap"
)

// TransactionHandler gerencia requisições de transações
type TransactionHandler struct {
	executeUseCase *usecases.ExecuteEVMTransactionUseCase
	config         *pkgconfig.Config
	logger         *zap.Logger
}

// NewTransactionHandler cria um novo handler de transações
func NewTransactionHandler(
	executeUseCase *usecases.ExecuteEVMTransactionUseCase,
	config *pkgconfig.Config,
	logger *zap.Logger,
) *TransactionHandler {
	return &TransactionHandler{
		executeUseCase: executeUseCase,
		config:         config,
		logger:         logger,
	}
}

// ExecuteTransaction processa a execução de uma transação
func (h *TransactionHandler) ExecuteTransaction(
	ctx context.Context,
	req *dtos.ExecuteTransactionRequest,
) (*dtos.ExecuteTransactionResponse, int, error) {
	return h.executeTransactionInternal(ctx, req)
}

// executeTransactionInternal processa internamente a execução de uma transação
func (h *TransactionHandler) executeTransactionInternal(
	ctx context.Context,
	req *dtos.ExecuteTransactionRequest,
) (*dtos.ExecuteTransactionResponse, int, error) {
	if h.executeUseCase == nil {
		h.logger.Error("use case not initialized")
		return nil, http.StatusInternalServerError, errors.New("use case not initialized")
	}

	h.logger.Info("executing transaction via handler",
		zap.String("operation_id", req.OperationID),
		zap.String("chain_type", req.ChainType))

	// Chamar use case
	response, err := h.executeUseCase.Execute(ctx, req)
	if err != nil {
		h.logger.Error("failed to execute transaction",
			zap.String("operation_id", req.OperationID),
			zap.Error(err))
		return nil, http.StatusInternalServerError, err
	}

	h.logger.Info("transaction executed successfully",
		zap.String("operation_id", response.OperationID),
		zap.String("status", response.Status))

	return response, http.StatusOK, nil
}

// ValidateRequest valida uma requisição de transação
func (h *TransactionHandler) ValidateRequest(req *dtos.ExecuteTransactionRequest) error {
	if req.OperationID == "" {
		return errors.New("operation_id is required")
	}

	if req.ChainType == "" {
		return errors.New("chain_type is required")
	}

	if req.FromAddress == "" {
		return errors.New("from_address is required")
	}

	if req.ToAddress == "" {
		return errors.New("to_address is required")
	}

	return nil
}
