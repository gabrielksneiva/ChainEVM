package rpc

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestNewTransactionSigner(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	timeout := 30 * time.Second

	signer := NewTransactionSigner(mockClient, chainID, logger, timeout)

	assert.NotNil(t, signer)
	assert.Equal(t, chainID, signer.chainID)
	assert.Equal(t, timeout, signer.timeout)
	assert.Equal(t, 3*time.Second, signer.pollInterval)
}

func TestWaitForConfirmations_Success(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	signer := NewTransactionSigner(mockClient, chainID, logger, 10*time.Second)

	txHash := common.HexToHash("0x123abc")
	blockNumber := uint64(1000)

	// Transaction already mined and confirmed immediately
	mockClient.On("TransactionReceipt", mock.Anything, txHash).
		Return(&types.Receipt{
			BlockNumber: big.NewInt(int64(blockNumber)),
			GasUsed:     21000,
		}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(1015), nil) // 15 confirmations

	receipt, err := signer.WaitForConfirmations(context.Background(), "0x123abc", 12)

	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, blockNumber, receipt.BlockNumber.Uint64())
}

func TestWaitForConfirmations_Timeout(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	signer := NewTransactionSigner(mockClient, chainID, logger, 100*time.Millisecond)

	txHash := common.HexToHash("0x123abc")

	// Always return not mined
	mockClient.On("TransactionReceipt", mock.Anything, txHash).
		Return(nil, errors.New("not mined"))

	receipt, err := signer.WaitForConfirmations(context.Background(), "0x123abc", 12)

	assert.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "timeout")
}

func TestWaitForConfirmations_ErrorOnBlockNumber(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	signer := NewTransactionSigner(mockClient, chainID, logger, 5*time.Second)

	txHash := common.HexToHash("0x123abc")

	mockClient.On("TransactionReceipt", mock.Anything, txHash).
		Return(&types.Receipt{
			BlockNumber: big.NewInt(1000),
			GasUsed:     21000,
		}, nil)
	mockClient.On("BlockNumber", mock.Anything).Return(uint64(0), errors.New("network error"))

	receipt, err := signer.WaitForConfirmations(context.Background(), "0x123abc", 12)

	// Deve continuar tentando quando BlockNumber falha
	assert.Error(t, err)
	assert.Nil(t, receipt)
}

func TestParsePrivateKey_Valid(t *testing.T) {
	// Valid test private key (secp256k1)
	validKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb476c6b8d6c1f02b28d62f0a77fc"

	pk, err := parsePrivateKey(validKey)

	assert.NoError(t, err)
	assert.NotNil(t, pk)
}

func TestParsePrivateKey_WithPrefix(t *testing.T) {
	// Valid test private key with 0x prefix
	validKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb476c6b8d6c1f02b28d62f0a77fc"

	pk, err := parsePrivateKey(validKey)

	assert.NoError(t, err)
	assert.NotNil(t, pk)
}

func TestParsePrivateKey_Invalid(t *testing.T) {
	invalidKey := "invalid"

	pk, err := parsePrivateKey(invalidKey)

	assert.Error(t, err)
	assert.Nil(t, pk)
}

func TestSignAndSendTransaction_InvalidPrivateKey(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	signer := NewTransactionSigner(mockClient, chainID, logger, 5*time.Second)

	ctx := context.Background()
	tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 0, big.NewInt(0), nil)

	txHash, err := signer.SignAndSendTransaction(ctx, tx, "invalid_key")

	assert.Error(t, err)
	assert.Empty(t, txHash)
}

func TestSignAndSendTransaction_Success(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	signer := NewTransactionSigner(mockClient, chainID, logger, 5*time.Second)

	ctx := context.Background()
	to := common.HexToAddress("0x0987654321098765432109876543210987654321")
	tx := types.NewTransaction(0, to, big.NewInt(1000), 21000, big.NewInt(20000000000), nil)

	// Valid Ethereum private key (32 bytes hex)
	validPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	mockClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).Return(nil)

	txHash, err := signer.SignAndSendTransaction(ctx, tx, validPrivateKey)

	assert.NoError(t, err)
	assert.NotEmpty(t, txHash)
	assert.True(t, strings.HasPrefix(txHash, "0x"))
	mockClient.AssertExpectations(t)
}

func TestSignAndSendTransaction_SendError(t *testing.T) {
	mockClient := new(MockEthClient)
	logger := zap.NewNop()
	chainID := big.NewInt(11155111)
	signer := NewTransactionSigner(mockClient, chainID, logger, 5*time.Second)

	ctx := context.Background()
	to := common.HexToAddress("0x0987654321098765432109876543210987654321")
	tx := types.NewTransaction(0, to, big.NewInt(1000), 21000, big.NewInt(20000000000), nil)

	validPrivateKey := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	mockClient.On("SendTransaction", mock.Anything, mock.AnythingOfType("*types.Transaction")).Return(errors.New("send failed"))

	txHash, err := signer.SignAndSendTransaction(ctx, tx, validPrivateKey)

	assert.Error(t, err)
	assert.Empty(t, txHash)
	assert.Contains(t, err.Error(), "send failed")
	mockClient.AssertExpectations(t)
}
