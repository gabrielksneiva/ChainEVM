package entities

import (
	"testing"

	"github.com/gabrielksneiva/ChainEVM/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEVMTransaction(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	tx := NewEVMTransaction(
		operationID,
		chainType,
		operationType,
		fromAddr,
		toAddr,
		map[string]interface{}{"amount": "1.5"},
		"idempotency-key-123",
	)

	require.NotNil(t, tx)
	assert.Equal(t, operationID, tx.OperationID())
	assert.Equal(t, chainType, tx.ChainType())
	assert.Equal(t, operationType, tx.OperationType())
	assert.Equal(t, fromAddr, tx.FromAddress())
	assert.Equal(t, toAddr, tx.ToAddress())
	assert.Equal(t, TransactionStatusPending, tx.Status())
	assert.Equal(t, "idempotency-key-123", tx.IdempotencyKey())
}

func TestMarkAsProcessing(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	tx := NewEVMTransaction(operationID, chainType, operationType, fromAddr, toAddr, map[string]interface{}{}, "key")
	tx.MarkAsProcessing()

	assert.Equal(t, TransactionStatusProcessing, tx.Status())
}

func TestGetters(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	payload := map[string]interface{}{"amount": "1.5"}
	tx := NewEVMTransaction(operationID, chainType, operationType, fromAddr, toAddr, payload, "key")

	assert.Equal(t, payload, tx.Payload())
	assert.NotNil(t, tx.CreatedAt())
}

func TestMarkAsConfirmed(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	tx := NewEVMTransaction(operationID, chainType, operationType, fromAddr, toAddr, map[string]interface{}{}, "key")
	tx.MarkAsConfirmed()

	assert.Equal(t, TransactionStatusConfirmed, tx.Status())
}

func TestMarkAsSuccess(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	tx := NewEVMTransaction(operationID, chainType, operationType, fromAddr, toAddr, map[string]interface{}{}, "key")

	txHash, _ := valueobjects.NewTransactionHash("0x1234567890123456789012345678901234567890123456789012345678901234")
	tx.MarkAsSuccess(txHash, 12345, 21000)

	assert.Equal(t, TransactionStatusSuccess, tx.Status())
	assert.Equal(t, txHash, tx.TxHash())
	assert.Equal(t, int64(12345), *tx.BlockNumber())
	assert.Equal(t, int64(21000), *tx.GasUsed())
	assert.NotNil(t, tx.ExecutedAt())
}

func TestMarkAsFailed(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	tx := NewEVMTransaction(operationID, chainType, operationType, fromAddr, toAddr, map[string]interface{}{}, "key")
	tx.MarkAsFailed("insufficient funds")

	assert.Equal(t, TransactionStatusFailed, tx.Status())
	assert.Equal(t, "insufficient funds", tx.ErrorMessage())
	assert.NotNil(t, tx.ExecutedAt())
}

func TestSetTxMetadata(t *testing.T) {
	operationID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	operationType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
	toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

	tx := NewEVMTransaction(operationID, chainType, operationType, fromAddr, toAddr, map[string]interface{}{}, "key")
	tx.SetTxMetadata("20000000000", 5)

	assert.NotNil(t, tx.GasPrice())
	assert.Equal(t, "20000000000", *tx.GasPrice())
	assert.NotNil(t, tx.Nonce())
	assert.Equal(t, int64(5), *tx.Nonce())
}
