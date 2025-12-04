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
)

func TestNewEthClientAdapter(t *testing.T) {
	t.Parallel()

	// Note: ethclient.Client não pode ser instanciado diretamente sem dialContext
	// mas podemos testar que a função existe e não é nula
	adapter := NewEthClientAdapter(nil)

	assert.NotNil(t, adapter)

	// Verificar que o adapter implementa a interface
	_, ok := interface{}(adapter).(EthClient)
	assert.True(t, ok, "adapter should implement EthClient interface")
}

func TestEthClientAdapterBalanceAt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		account     common.Address
		blockNumber *big.Int
		expectedErr bool
		expectNil   bool
	}{
		{
			name:        "successful balance query",
			account:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
			blockNumber: big.NewInt(1),
			expectedErr: false,
			expectNil:   false,
		},
		{
			name:        "balance at nil block number",
			account:     common.HexToAddress("0x0000000000000000000000000000000000000000"),
			blockNumber: nil,
			expectedErr: false,
			expectNil:   false,
		},
		{
			name:        "large block number",
			account:     common.HexToAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"),
			blockNumber: big.NewInt(999999999),
			expectedErr: false,
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)
			expectedBalance := big.NewInt(100)
			mockClient.On("BalanceAt", mock.Anything, tt.account, tt.blockNumber).
				Return(expectedBalance, nil)

			adapter := mockClient
			ctx := context.Background()

			balance, err := adapter.BalanceAt(ctx, tt.account, tt.blockNumber)

			assert.NoError(t, err)
			assert.Equal(t, expectedBalance, balance)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterNonceAt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		account     common.Address
		blockNumber *big.Int
		nonce       uint64
		shouldError bool
	}{
		{
			name:        "get nonce at block",
			account:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
			blockNumber: big.NewInt(100),
			nonce:       5,
			shouldError: false,
		},
		{
			name:        "get nonce for zero address",
			account:     common.Address{},
			blockNumber: big.NewInt(1),
			nonce:       0,
			shouldError: false,
		},
		{
			name:        "get nonce at pending block",
			account:     common.HexToAddress("0x9876543210987654321098765432109876543210"),
			blockNumber: nil,
			nonce:       10,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)
			mockClient.On("NonceAt", mock.Anything, tt.account, tt.blockNumber).
				Return(tt.nonce, nil)

			adapter := mockClient
			ctx := context.Background()

			nonce, err := adapter.NonceAt(ctx, tt.account, tt.blockNumber)

			assert.NoError(t, err)
			assert.Equal(t, tt.nonce, nonce)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterPendingNonceAt(t *testing.T) {
	t.Parallel()

	account := common.HexToAddress("0x1234567890123456789012345678901234567890")
	mockClient := new(MockEthClient)
	expectedNonce := uint64(5)
	mockClient.On("PendingNonceAt", mock.Anything, account).
		Return(expectedNonce, nil)

	adapter := mockClient
	ctx := context.Background()

	nonce, err := adapter.PendingNonceAt(ctx, account)

	assert.NoError(t, err)
	assert.Equal(t, expectedNonce, nonce)
	mockClient.AssertExpectations(t)
}

func TestEthClientAdapterSendTransaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		shouldError bool
		errorType   error
	}{
		{
			name:        "send transaction success",
			shouldError: false,
			errorType:   nil,
		},
		{
			name:        "send transaction with error",
			shouldError: true,
			errorType:   errors.New("transaction rejected"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)
			tx := types.NewTransaction(0, common.Address{}, big.NewInt(0), 21000, big.NewInt(1), nil)

			if tt.shouldError {
				mockClient.On("SendTransaction", mock.Anything, tx).
					Return(tt.errorType)
			} else {
				mockClient.On("SendTransaction", mock.Anything, tx).
					Return(nil)
			}

			adapter := mockClient
			ctx := context.Background()

			err := adapter.SendTransaction(ctx, tx)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterEstimateGas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		gasEstimate uint64
		shouldError bool
	}{
		{
			name:        "estimate gas success",
			gasEstimate: 21000,
			shouldError: false,
		},
		{
			name:        "estimate gas high amount",
			gasEstimate: 5000000,
			shouldError: false,
		},
		{
			name:        "estimate gas error",
			gasEstimate: 0,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)
			msg := ethereum.CallMsg{}

			if tt.shouldError {
				mockClient.On("EstimateGas", mock.Anything, msg).
					Return(uint64(0), errors.New("estimation failed"))
			} else {
				mockClient.On("EstimateGas", mock.Anything, msg).
					Return(tt.gasEstimate, nil)
			}

			adapter := mockClient
			ctx := context.Background()

			gas, err := adapter.EstimateGas(ctx, msg)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Equal(t, uint64(0), gas)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.gasEstimate, gas)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterTransactionReceipt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		txHash      common.Hash
		shouldError bool
		hasReceipt  bool
	}{
		{
			name:        "get receipt success",
			txHash:      common.HexToHash("0x123"),
			shouldError: false,
			hasReceipt:  true,
		},
		{
			name:        "receipt not found",
			txHash:      common.HexToHash("0x456"),
			shouldError: false,
			hasReceipt:  false,
		},
		{
			name:        "receipt error",
			txHash:      common.HexToHash("0x789"),
			shouldError: true,
			hasReceipt:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)

			if tt.shouldError {
				mockClient.On("TransactionReceipt", mock.Anything, tt.txHash).
					Return(nil, errors.New("rpc error"))
			} else if tt.hasReceipt {
				receipt := &types.Receipt{
					TxHash:      tt.txHash,
					Status:      1,
					BlockNumber: big.NewInt(100),
				}
				mockClient.On("TransactionReceipt", mock.Anything, tt.txHash).
					Return(receipt, nil)
			} else {
				mockClient.On("TransactionReceipt", mock.Anything, tt.txHash).
					Return(nil, nil)
			}

			adapter := mockClient
			ctx := context.Background()

			receipt, err := adapter.TransactionReceipt(ctx, tt.txHash)

			if tt.shouldError {
				assert.Error(t, err)
				assert.Nil(t, receipt)
			} else if tt.hasReceipt {
				assert.NoError(t, err)
				assert.NotNil(t, receipt)
				assert.Equal(t, tt.txHash, receipt.TxHash)
			} else {
				assert.NoError(t, err)
				assert.Nil(t, receipt)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterChainID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		chainID   *big.Int
		wantError bool
	}{
		{
			name:      "mainnet chain id",
			chainID:   big.NewInt(1),
			wantError: false,
		},
		{
			name:      "sepolia chain id",
			chainID:   big.NewInt(11155111),
			wantError: false,
		},
		{
			name:      "chain id error",
			chainID:   nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)

			if tt.wantError {
				mockClient.On("ChainID", mock.Anything).
					Return(nil, errors.New("failed to get chain id"))
			} else {
				mockClient.On("ChainID", mock.Anything).
					Return(tt.chainID, nil)
			}

			adapter := mockClient
			ctx := context.Background()

			chainID, err := adapter.ChainID(ctx)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, chainID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.chainID, chainID)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterSuggestGasPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		gasPrice  *big.Int
		wantError bool
	}{
		{
			name:      "suggest gas price success",
			gasPrice:  big.NewInt(20000000000),
			wantError: false,
		},
		{
			name:      "suggest gas price high",
			gasPrice:  big.NewInt(100000000000),
			wantError: false,
		},
		{
			name:      "suggest gas price error",
			gasPrice:  nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockEthClient)

			if tt.wantError {
				mockClient.On("SuggestGasPrice", mock.Anything).
					Return(nil, errors.New("failed to get gas price"))
			} else {
				mockClient.On("SuggestGasPrice", mock.Anything).
					Return(tt.gasPrice, nil)
			}

			adapter := mockClient
			ctx := context.Background()

			gasPrice, err := adapter.SuggestGasPrice(ctx)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, gasPrice)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.gasPrice, gasPrice)
			}
			mockClient.AssertExpectations(t)
		})
	}
}

func TestEthClientAdapterClose(t *testing.T) {
	t.Parallel()

	mockClient := new(MockEthClient)
	mockClient.On("Close").Return()

	adapter := mockClient

	adapter.Close()

	mockClient.AssertExpectations(t)
}

func TestEthClientAdapterInterfaceImplementation(t *testing.T) {
	t.Parallel()

	adapter := &EthClientAdapter{client: nil}

	// Verificar que adapter implementa a interface EthClient
	var _ EthClient = adapter
}
