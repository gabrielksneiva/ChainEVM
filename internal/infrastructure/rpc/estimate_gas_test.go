package rpc

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockCallMsg implementa a interface necessária para EstimateGas
type mockCallMsg struct {
	from  common.Address
	to    *common.Address
	data  []byte
	value *big.Int
}

func (m mockCallMsg) GetFrom() common.Address {
	return m.from
}

func (m mockCallMsg) GetTo() *common.Address {
	return m.to
}

func (m mockCallMsg) GetData() []byte {
	return m.data
}

func (m mockCallMsg) GetValue() *big.Int {
	return m.value
}

func TestNewEVMRPCClient_InvalidURL(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()

	// URL inválida deve retornar erro
	client, err := NewEVMRPCClient("invalid://url", 30*time.Second, logger)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to connect to EVM RPC")
}

func TestNewEVMRPCClient_Success(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()

	// Valid http URL (will fail to connect but will test the success path of validation)
	// We expect this to fail since we don't have a real RPC endpoint, but it tests the URL parsing
	client, err := NewEVMRPCClient("http://localhost:8545", 30*time.Second, logger)

	// This might error but should not error on URL parsing
	if err != nil {
		// Error is expected when connecting to non-existent RPC
		assert.Contains(t, err.Error(), "failed to connect")
	} else {
		assert.NotNil(t, client)
	}
}

func TestEstimateGas_NilData(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	mockClient := new(MockEthClient)

	rpcClient := &EVMRPCClient{
		client:  mockClient,
		rpcURL:  "http://localhost:8545",
		timeout: 30 * time.Second,
		logger:  logger,
	}

	msg := mockCallMsg{
		from:  common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"),
		to:    nil,
		data:  nil,
		value: big.NewInt(0),
	}

	gas, err := rpcClient.EstimateGas(context.Background(), msg)

	assert.NoError(t, err)
	assert.Equal(t, uint64(21000), gas, "Gas for nil data should be base gas (21000)")
}

func TestEstimateGas_EmptyData(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	mockClient := new(MockEthClient)

	rpcClient := &EVMRPCClient{
		client:  mockClient,
		rpcURL:  "http://localhost:8545",
		timeout: 30 * time.Second,
		logger:  logger,
	}

	msg := mockCallMsg{
		from:  common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"),
		to:    nil,
		data:  []byte{},
		value: big.NewInt(0),
	}

	gas, err := rpcClient.EstimateGas(context.Background(), msg)

	assert.NoError(t, err)
	assert.Equal(t, uint64(21000), gas, "Gas for empty data should be base gas (21000)")
}

func TestEstimateGas_WithData(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	mockClient := new(MockEthClient)

	rpcClient := &EVMRPCClient{
		client:  mockClient,
		rpcURL:  "http://localhost:8545",
		timeout: 30 * time.Second,
		logger:  logger,
	}

	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	msg := mockCallMsg{
		from:  common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"),
		to:    nil,
		data:  data,
		value: big.NewInt(0),
	}

	gas, err := rpcClient.EstimateGas(context.Background(), msg)

	assert.NoError(t, err)
	expectedGas := uint64(21000) + uint64(len(data)*4)
	assert.Equal(t, expectedGas, gas, "Gas should be base + 4 per byte")
}

func TestEstimateGas_LargeData(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	mockClient := new(MockEthClient)

	rpcClient := &EVMRPCClient{
		client:  mockClient,
		rpcURL:  "http://localhost:8545",
		timeout: 30 * time.Second,
		logger:  logger,
	}

	data := make([]byte, 1000)
	msg := mockCallMsg{
		from:  common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"),
		to:    nil,
		data:  data,
		value: big.NewInt(0),
	}

	gas, err := rpcClient.EstimateGas(context.Background(), msg)

	assert.NoError(t, err)
	expectedGas := uint64(21000) + uint64(1000*4)
	assert.Equal(t, expectedGas, gas, "Gas should scale with data size")
}

func TestEstimateGas_InvalidMessageType(t *testing.T) {
	t.Parallel()

	logger, _ := zap.NewDevelopment()
	mockClient := new(MockEthClient)

	rpcClient := &EVMRPCClient{
		client:  mockClient,
		rpcURL:  "http://localhost:8545",
		timeout: 30 * time.Second,
		logger:  logger,
	}

	// Passar um tipo inválido
	invalidMsg := "invalid message"

	gas, err := rpcClient.EstimateGas(context.Background(), invalidMsg)

	assert.Error(t, err)
	assert.Equal(t, uint64(0), gas)
	assert.Contains(t, err.Error(), "invalid message type")
}
