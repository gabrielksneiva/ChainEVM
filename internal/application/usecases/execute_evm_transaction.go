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
	signer          rpc.SignedTransactionClient
	logger          *zap.Logger
}

// NewExecuteEVMTransactionUseCase cria uma nova instância do caso de uso
func NewExecuteEVMTransactionUseCase(
	rpcClients map[string]rpc.RPCClient,
	transactionRepo database.TransactionRepository,
	signer rpc.SignedTransactionClient,
	logger *zap.Logger,
) *ExecuteEVMTransactionUseCase {
	return &ExecuteEVMTransactionUseCase{
		rpcClients:      rpcClients,
		transactionRepo: transactionRepo,
		signer:          signer,
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
		if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
			uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
		}
		return nil, pkgerrors.NewAppError(pkgerrors.ErrDatabaseError.Code, "failed to save transaction", err)
	}

	// Executar operação
	rpcClient, ok := uc.rpcClients[chainType.String()]
	if !ok {
		uc.logger.Error("RPC client not found for chain", zap.String("chain", chainType.String()))
		transaction.MarkAsFailed("RPC client not found")
		if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
			uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
		}
		return nil, pkgerrors.NewAppError(pkgerrors.ErrChainNotSupported.Code, "chain not supported", nil)
	}

	// Executar baseado no tipo de operação
	txHash := valueobjects.TransactionHash("")
	blockNumber := int64(0)
	gasUsed := int64(0)

	if operationType.IsWriteOperation() {
		// Executar transação de escrita
		uc.logger.Info("executing write operation", zap.String("operation_type", operationType.String()))

		if uc.signer == nil {
			uc.logger.Error("transaction signer not configured")
			transaction.MarkAsFailed("transaction signer not configured")
			if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
				uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
			}
			return nil, pkgerrors.NewAppError(pkgerrors.ErrValidationFailed.Code, "signer not configured", nil)
		}

		// Get nonce
		nonce, err := rpcClient.GetNonce(ctx, fromAddr.String())
		if err != nil {
			uc.logger.Error("failed to get nonce", zap.Error(err))
			transaction.MarkAsFailed("failed to get nonce")
			if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
				uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
			}
			return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get nonce", err)
		}

		// Get gas price
		gasPrice, err := rpcClient.GetGasPrice(ctx)
		if err != nil {
			uc.logger.Error("failed to get gas price", zap.Error(err))
			transaction.MarkAsFailed("failed to get gas price")
			if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
				uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
			}
			return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get gas price", err)
		}

		transaction.SetTxMetadata(gasPrice.String(), int64(nonce))

		// Sign and send transaction - extract private key from payload or environment
		// NOTE: In production, private keys should come from AWS Secrets Manager or similar
		txHashStr, err := uc.signer.SignAndSendTransaction(ctx, nil, "")
		if err != nil {
			uc.logger.Error("failed to sign and send transaction", zap.Error(err))
			transaction.MarkAsFailed(err.Error())
			if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
				uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
			}
			return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to sign and send transaction", err)
		}

		// Wait for confirmations
		receipt, err := uc.signer.WaitForConfirmations(ctx, txHashStr, 12) // 12 confirmations
		if err != nil {
			uc.logger.Error("transaction confirmation timeout", zap.Error(err), zap.String("tx_hash", txHashStr))
			transaction.MarkAsFailed("confirmation timeout")
			if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
				uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
			}
			return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "transaction not confirmed", err)
		}

		// Create transaction hash from result
		txHash, hashErr := valueobjects.NewTransactionHash(txHashStr)
		if hashErr != nil {
			uc.logger.Error("failed to create transaction hash", zap.Error(hashErr))
		}

		// Mark as success with confirmation data
		transaction.MarkAsSuccess(txHash, int64(receipt.BlockNumber.Uint64()), int64(receipt.GasUsed))
		blockNumber = int64(receipt.BlockNumber.Uint64())
		gasUsed = int64(receipt.GasUsed)

	} else {
		// Executar query (read-only)
		uc.logger.Info("executing read operation", zap.String("operation_type", operationType.String()))

		switch operationType {
		case valueobjects.OperationTypeGetBalance:
			balance, err := rpcClient.GetBalance(ctx, toAddr.String())
			if err != nil {
				uc.logger.Error("failed to get balance", zap.Error(err))
				transaction.MarkAsFailed("failed to get balance")
				if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
					uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
				}
				return nil, pkgerrors.NewAppError(pkgerrors.ErrRPCFailed.Code, "failed to get balance", err)
			}
			_ = balance.String()

		case valueobjects.OperationTypeGetNonce:
			nonce, err := rpcClient.GetNonce(ctx, fromAddr.String())
			if err != nil {
				uc.logger.Error("failed to get nonce", zap.Error(err))
				transaction.MarkAsFailed("failed to get nonce")
				if saveErr := uc.transactionRepo.Save(ctx, transaction); saveErr != nil {
					uc.logger.Error("failed to save failed transaction", zap.Error(saveErr))
				}
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
