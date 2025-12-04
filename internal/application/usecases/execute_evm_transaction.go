package usecases

import (
	"context"
	"time"

	"github.com/gabrielksneiva/ChainEVM/internal/application/dtos"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/entities"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/database"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/rpc"
	pkgerrors "github.com/gabrielksneiva/ChainEVM/pkg/errors"
	"go.uber.org/zap"
)

// ExecuteEVMTransactionUseCase caso de uso para executar transações EVM
type ExecuteEVMTransactionUseCase struct {
	rpcClients      map[string]rpc.RPCClient
	transactionRepo database.TransactionRepository
	logger          *zap.Logger
}

// NewExecuteEVMTransactionUseCase cria uma nova instância do caso de uso
func NewExecuteEVMTransactionUseCase(
	rpcClients map[string]rpc.RPCClient,
	transactionRepo database.TransactionRepository,
	logger *zap.Logger,
) *ExecuteEVMTransactionUseCase {
	return &ExecuteEVMTransactionUseCase{
		rpcClients:      rpcClients,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

// Execute executa uma transação EVM
func (uc *ExecuteEVMTransactionUseCase) Execute(
	ctx context.Context,
	req *dtos.ExecuteTransactionRequest,
) (*dtos.ExecuteTransactionResponse, error) {
	uc.logger.Info("executing EVM transaction",
		zap.String("operation_id", req.OperationID),
		zap.String("chain_type", req.ChainType),
		zap.String("operation_type", req.OperationType),
	)

	// Validação básica
	chainType, err := valueobjects.NewChainType(req.ChainType)
	if err != nil {
		uc.logger.Error("invalid chain type", zap.Error(err))
		return nil, pkgerrors.NewAppError(pkgerrors.ErrChainNotSupported.Code, err.Error(), err)
	}

	operationType, err := valueobjects.NewOperationType(req.OperationType)
	if err != nil {
		uc.logger.Error("invalid operation type", zap.Error(err))
		return nil, pkgerrors.NewAppError(pkgerrors.ErrValidationFailed.Code, err.Error(), err)
	}

	operationID, err := valueobjects.NewOperationID(req.OperationID)
	if err != nil {
		uc.logger.Error("invalid operation ID", zap.Error(err))
		return nil, pkgerrors.NewAppError(pkgerrors.ErrValidationFailed.Code, err.Error(), err)
	}

	fromAddr, err := valueobjects.NewEVMAddress(req.FromAddress)
	if err != nil {
		uc.logger.Error("invalid from address", zap.Error(err))
		return nil, pkgerrors.NewAppError(pkgerrors.ErrValidationFailed.Code, err.Error(), err)
	}

	toAddr, err := valueobjects.NewEVMAddress(req.ToAddress)
	if err != nil {
		uc.logger.Error("invalid to address", zap.Error(err))
		return nil, pkgerrors.NewAppError(pkgerrors.ErrValidationFailed.Code, err.Error(), err)
	}

	// Verificar idempotência - check se a transação já foi processada
	existingTx, err := uc.transactionRepo.GetByIdempotencyKey(ctx, req.IdempotencyKey)
	if err == nil && existingTx != nil {
		uc.logger.Info("transaction already processed (idempotent)",
			zap.String("idempotency_key", req.IdempotencyKey))
		return buildResponse(existingTx), nil
	}

	// Criar entidade de domínio
	transaction := entities.NewEVMTransaction(
		operationID,
		chainType,
		operationType,
		fromAddr,
		toAddr,
		req.Payload,
		req.IdempotencyKey,
	)

	// Marcar como processando
	transaction.MarkAsProcessing()

	// Salvar transação no banco
	if err := uc.transactionRepo.Save(ctx, transaction); err != nil {
		uc.logger.Error("failed to save transaction", zap.Error(err))
		transaction.MarkAsFailed("database error")
		uc.transactionRepo.Save(ctx, transaction)
		return nil, pkgerrors.NewAppError(pkgerrors.ErrDatabaseError.Code, "failed to save transaction", err)
	}

	// Executar operação
	rpcClient, ok := uc.rpcClients[chainType.String()]
	if !ok {
		uc.logger.Error("RPC client not found for chain", zap.String("chain", chainType.String()))
		transaction.MarkAsFailed("RPC client not found")
		uc.transactionRepo.Save(ctx, transaction)
		return nil, pkgerrors.NewAppError(pkgerrors.ErrChainNotSupported.Code, "chain not supported", nil)
	}

	// Executar baseado no tipo de operação
	var txHash valueobjects.TransactionHash
	var blockNumber int64
	var gasUsed int64

	if operationType.IsWriteOperation() {
		// Executar transação de escrita
		uc.logger.Info("executing write operation", zap.String("operation_type", operationType.String()))

		// Get nonce
		nonce, err := rpcClient.GetNonce(ctx, fromAddr.String())
		if err != nil {
			uc.logger.Error("failed to get nonce", zap.Error(err))
			transaction.MarkAsFailed("failed to get nonce")
			uc.transactionRepo.Save(ctx, transaction)
			return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get nonce", err)
		}

		// Get gas price
		gasPrice, err := rpcClient.GetGasPrice(ctx)
		if err != nil {
			uc.logger.Error("failed to get gas price", zap.Error(err))
			transaction.MarkAsFailed("failed to get gas price")
			uc.transactionRepo.Save(ctx, transaction)
			return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get gas price", err)
		}

		transaction.SetTxMetadata(gasPrice.String(), int64(nonce))

		// Demo: registrar que a operação foi processada
		// Em produção, assinaria e enviaria a transação real
		txHash, _ = valueobjects.NewTransactionHash("0x0000000000000000000000000000000000000000000000000000000000000000")
		blockNumber = 0
		gasUsed = 21000

	} else {
		// Executar query (read-only)
		uc.logger.Info("executing read operation", zap.String("operation_type", operationType.String()))

		switch operationType {
		case valueobjects.OperationTypeGetBalance:
			balance, err := rpcClient.GetBalance(ctx, toAddr.String())
			if err != nil {
				uc.logger.Error("failed to get balance", zap.Error(err))
				transaction.MarkAsFailed("failed to get balance")
				uc.transactionRepo.Save(ctx, transaction)
				return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get balance", err)
			}
			_ = balance.String()

		case valueobjects.OperationTypeGetNonce:
			nonce, err := rpcClient.GetNonce(ctx, fromAddr.String())
			if err != nil {
				uc.logger.Error("failed to get nonce", zap.Error(err))
				transaction.MarkAsFailed("failed to get nonce")
				uc.transactionRepo.Save(ctx, transaction)
				return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get nonce", err)
			}
			_ = nonce
		}

		transaction.MarkAsSuccess(txHash, blockNumber, gasUsed)
	}

	// Salvar transação com resultado
	if err := uc.transactionRepo.Save(ctx, transaction); err != nil {
		uc.logger.Error("failed to update transaction", zap.Error(err))
		return nil, pkgerrors.NewAppError(pkgerrors.ErrDatabaseError.Code, "failed to update transaction", err)
	}

	uc.logger.Info("transaction executed successfully",
		zap.String("operation_id", operationID.String()),
		zap.String("status", string(transaction.Status())),
	)

	response := buildResponse(transaction)
	return response, nil
}

func buildResponse(tx *entities.EVMTransaction) *dtos.ExecuteTransactionResponse {
	executedAt := ""
	if tx.ExecutedAt() != nil {
		executedAt = tx.ExecutedAt().Format(time.RFC3339)
	}

	return &dtos.ExecuteTransactionResponse{
		OperationID:     tx.OperationID().String(),
		ChainType:       tx.ChainType().String(),
		TransactionHash: tx.TxHash().String(),
		Status:          string(tx.Status()),
		BlockNumber:     tx.BlockNumber(),
		GasUsed:         tx.GasUsed(),
		GasPrice:        tx.GasPrice(),
		ErrorMessage:    tx.ErrorMessage(),
		CreatedAt:       tx.CreatedAt().Format(time.RFC3339),
		ExecutedAt:      &executedAt,
	}
}
