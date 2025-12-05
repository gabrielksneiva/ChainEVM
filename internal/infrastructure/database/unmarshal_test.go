package database

import (
	"testing"
	"time"

	"github.com/gabrielksneiva/ChainEVM/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUnmarshalTransactionItem_StatusProcessing(t *testing.T) {
	t.Parallel()

	item := TransactionItem{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		OperationType:  "TRANSFER",
		FromAddress:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:      "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Status:         string(entities.TransactionStatusProcessing),
		CreatedAt:      time.Now().Format(time.RFC3339),
		IdempotencyKey: "idem123",
	}

	logger, _ := zap.NewDevelopment()
	tx, err := unmarshalTransactionItem(item, logger)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, entities.TransactionStatusProcessing, tx.Status())
}

func TestUnmarshalTransactionItem_StatusConfirmed(t *testing.T) {
	t.Parallel()

	item := TransactionItem{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		OperationType:  "TRANSFER",
		FromAddress:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:      "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Status:         string(entities.TransactionStatusConfirmed),
		CreatedAt:      time.Now().Format(time.RFC3339),
		IdempotencyKey: "idem123",
	}

	logger, _ := zap.NewDevelopment()
	tx, err := unmarshalTransactionItem(item, logger)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, entities.TransactionStatusConfirmed, tx.Status())
}

func TestUnmarshalTransactionItem_StatusFailed(t *testing.T) {
	t.Parallel()

	errorMsg := "transaction failed"
	item := TransactionItem{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		OperationType:  "TRANSFER",
		FromAddress:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:      "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Status:         string(entities.TransactionStatusFailed),
		ErrorMessage:   errorMsg,
		CreatedAt:      time.Now().Format(time.RFC3339),
		IdempotencyKey: "idem123",
	}

	logger, _ := zap.NewDevelopment()
	tx, err := unmarshalTransactionItem(item, logger)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, entities.TransactionStatusFailed, tx.Status())
	assert.Equal(t, errorMsg, tx.ErrorMessage())
}

func TestUnmarshalTransactionItem_StatusSuccessWithHash(t *testing.T) {
	t.Parallel()

	blockNum := int64(12345)
	gasUsed := int64(21000)
	item := TransactionItem{
		OperationID:     "550e8400-e29b-41d4-a716-446655440000",
		ChainType:       "ETHEREUM",
		OperationType:   "TRANSFER",
		FromAddress:     "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:       "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Status:          string(entities.TransactionStatusSuccess),
		TransactionHash: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		BlockNumber:     &blockNum,
		GasUsed:         &gasUsed,
		CreatedAt:       time.Now().Format(time.RFC3339),
		IdempotencyKey:  "idem123",
	}

	logger, _ := zap.NewDevelopment()
	tx, err := unmarshalTransactionItem(item, logger)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, entities.TransactionStatusSuccess, tx.Status())
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", tx.TxHash().String())
	assert.Equal(t, &blockNum, tx.BlockNumber())
	assert.Equal(t, &gasUsed, tx.GasUsed())
}

func TestUnmarshalTransactionItem_StatusSuccessWithoutHash(t *testing.T) {
	t.Parallel()

	blockNum := int64(12345)
	gasUsed := int64(21000)
	item := TransactionItem{
		OperationID:     "550e8400-e29b-41d4-a716-446655440000",
		ChainType:       "ETHEREUM",
		OperationType:   "TRANSFER",
		FromAddress:     "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:       "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Status:          string(entities.TransactionStatusSuccess),
		TransactionHash: "",
		BlockNumber:     &blockNum,
		GasUsed:         &gasUsed,
		CreatedAt:       time.Now().Format(time.RFC3339),
		IdempotencyKey:  "idem123",
	}

	logger, _ := zap.NewDevelopment()
	tx, err := unmarshalTransactionItem(item, logger)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	// Quando não há TransactionHash, o MarkAsSuccess não é chamado, então permanece PENDING
	assert.Equal(t, entities.TransactionStatusPending, tx.Status())
}
