package rpc

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockEthClient mock do cliente Ethereum
type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockEthClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockEthClient) ChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) Close() {
	m.Called()
}

func TestEVMRPCClient_GetBalance_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	expectedBalance := big.NewInt(1000000000000000000) // 1 ETH

	mockClient.On("BalanceAt", mock.Anything, addr, mock.Anything).
		Return(expectedBalance, nil)

	balance, err := rpcClient.GetBalance(context.Background(), "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	assert.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetBalance_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	mockClient.On("BalanceAt", mock.Anything, addr, mock.Anything).
		Return(nil, errors.New("rpc error"))

	balance, err := rpcClient.GetBalance(context.Background(), "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.Contains(t, err.Error(), "failed to get balance")
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetNonce_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	mockClient.On("PendingNonceAt", mock.Anything, addr).
		Return(uint64(42), nil)

	nonce, err := rpcClient.GetNonce(context.Background(), "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	assert.NoError(t, err)
	assert.Equal(t, uint64(42), nonce)
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetNonce_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	addr := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	mockClient.On("PendingNonceAt", mock.Anything, addr).
		Return(uint64(0), errors.New("rpc error"))

	nonce, err := rpcClient.GetNonce(context.Background(), "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	assert.Error(t, err)
	assert.Equal(t, uint64(0), nonce)
	assert.Contains(t, err.Error(), "failed to get nonce")
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_SendTransaction_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	tx := types.NewTransaction(0, common.HexToAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72"), big.NewInt(1000), 21000, big.NewInt(1000000000), nil)

	mockClient.On("SendTransaction", mock.Anything, tx).
		Return(nil)

	err := rpcClient.SendTransaction(context.Background(), tx)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_SendTransaction_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	tx := types.NewTransaction(0, common.HexToAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72"), big.NewInt(1000), 21000, big.NewInt(1000000000), nil)

	mockClient.On("SendTransaction", mock.Anything, tx).
		Return(errors.New("rpc error"))

	err := rpcClient.SendTransaction(context.Background(), tx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send transaction")
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetChainID_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	expectedChainID := big.NewInt(1) // Mainnet

	mockClient.On("ChainID", mock.Anything).
		Return(expectedChainID, nil)

	chainID, err := rpcClient.GetChainID(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedChainID, chainID)
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetChainID_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	mockClient.On("ChainID", mock.Anything).
		Return(nil, errors.New("rpc error"))

	chainID, err := rpcClient.GetChainID(context.Background())

	assert.Error(t, err)
	assert.Nil(t, chainID)
	assert.Contains(t, err.Error(), "failed to get chain ID")
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetGasPrice_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	expectedGasPrice := big.NewInt(20000000000) // 20 Gwei

	mockClient.On("SuggestGasPrice", mock.Anything).
		Return(expectedGasPrice, nil)

	gasPrice, err := rpcClient.GetGasPrice(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, expectedGasPrice, gasPrice)
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetGasPrice_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	mockClient.On("SuggestGasPrice", mock.Anything).
		Return(nil, errors.New("rpc error"))

	gasPrice, err := rpcClient.GetGasPrice(context.Background())

	assert.Error(t, err)
	assert.Nil(t, gasPrice)
	assert.Contains(t, err.Error(), "failed to get gas price")
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetTransactionReceipt_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	expectedReceipt := &types.Receipt{
		Status: 1,
	}

	mockClient.On("TransactionReceipt", mock.Anything, txHash).
		Return(expectedReceipt, nil)

	receipt, err := rpcClient.GetTransactionReceipt(context.Background(), txHash.Hex())

	assert.NoError(t, err)
	assert.Equal(t, expectedReceipt, receipt)
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_GetTransactionReceipt_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	txHash := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	mockClient.On("TransactionReceipt", mock.Anything, txHash).
		Return(nil, errors.New("rpc error"))

	receipt, err := rpcClient.GetTransactionReceipt(context.Background(), txHash.Hex())

	assert.Error(t, err)
	assert.Nil(t, receipt)
	assert.Contains(t, err.Error(), "failed to get transaction receipt")
	mockClient.AssertExpectations(t)
}

func TestEVMRPCClient_Close(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	logger, _ := zap.NewDevelopment()

	rpcClient := &EVMRPCClient{
		client: mockClient,
		logger: logger,
	}

	mockClient.On("Close").Return()

	err := rpcClient.Close()

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}
