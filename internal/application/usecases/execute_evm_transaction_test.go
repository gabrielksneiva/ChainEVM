package usecases

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gabrielksneiva/ChainEVM/internal/application/dtos"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/entities"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/valueobjects"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockRPCClient implementa rpc.RPCClient interface
type MockRPCClient struct {
	mock.Mock
}

func (m *MockRPCClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

func (m *MockRPCClient) EstimateGas(ctx context.Context, msg interface{}) (uint64, error) {
	args := m.Called(ctx, msg)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockRPCClient) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockRPCClient) GetChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockRPCClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockTransactionRepository implementa database.TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Save(ctx context.Context, transaction *entities.EVMTransaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByOperationID(ctx context.Context, operationID string) (*entities.EVMTransaction, error) {
	args := m.Called(ctx, operationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EVMTransaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entities.EVMTransaction, error) {
	args := m.Called(ctx, idempotencyKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.EVMTransaction), args.Error(1)
}

func (m *MockTransactionRepository) UpdateStatus(ctx context.Context, operationID string, status entities.TransactionStatus) error {
	args := m.Called(ctx, operationID, status)
	return args.Error(0)
}

func TestExecuteEVMTransactionUseCase_Execute(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("execute get_balance successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{
			"ETHEREUM": mockRPC,
		}

		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440001",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetBalance", mock.Anything, mock.AnythingOfType("string")).Return(big.NewInt(5000000000000000000), nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, req.OperationID, resp.OperationID)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("execute get_nonce successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{
			"ETHEREUM": mockRPC,
		}

		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440003",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_NONCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440004",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetNonce", mock.Anything, mock.AnythingOfType("string")).Return(uint64(42), nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("execute write operation successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{
			"ETHEREUM": mockRPC,
		}

		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440005",
			ChainType:      "ETHEREUM",
			OperationType:  "TRANSFER",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{"amount": "1000000000000000000"},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440006",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetNonce", mock.Anything, mock.AnythingOfType("string")).Return(uint64(10), nil)
		mockRPC.On("GetGasPrice", mock.Anything).Return(big.NewInt(20000000000), nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("return existing transaction with idempotency key", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{
			"ETHEREUM": mockRPC,
		}

		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		chainType, _ := valueobjects.NewChainType("ETHEREUM")
		opType, _ := valueobjects.NewOperationType("GET_BALANCE")
		opID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440007")
		fromAddr, _ := valueobjects.NewEVMAddress("0x1234567890123456789012345678901234567890")
		toAddr, _ := valueobjects.NewEVMAddress("0x0987654321098765432109876543210987654321")

		existingTx := entities.NewEVMTransaction(opID, chainType, opType, fromAddr, toAddr, map[string]interface{}{}, "550e8400-e29b-41d4-a716-446655440008")

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440007",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440008",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(existingTx, nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertNotCalled(t, "GetBalance")
	})

	t.Run("fail with invalid chain type", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440009",
			ChainType:      "INVALID",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440010",
		}

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("fail with invalid operation ID", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "invalid",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440012",
		}

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("fail when save fails", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440018",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440019",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(errors.New("db error"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail when RPC client not found", func(t *testing.T) {
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440020",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440021",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		mockRepo.AssertExpectations(t)
	})

	t.Run("fail when GetBalance fails", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440022",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x1234567890123456789012345678901234567890",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440023",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetBalance", mock.Anything, mock.AnythingOfType("string")).Return(nil, errors.New("rpc error"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("fail when GetNonce fails in write operation", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440026",
			ChainType:      "ETHEREUM",
			OperationType:  "TRANSFER",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{"amount": "1000000000000000000"},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440027",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetNonce", mock.Anything, mock.AnythingOfType("string")).Return(uint64(0), errors.New("rpc error"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("fail when GetGasPrice fails", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440028",
			ChainType:      "ETHEREUM",
			OperationType:  "TRANSFER",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{"amount": "1000000000000000000"},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440029",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetNonce", mock.Anything, mock.AnythingOfType("string")).Return(uint64(10), nil)
		mockRPC.On("GetGasPrice", mock.Anything).Return(nil, errors.New("rpc error"))

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("execute CALL operation successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440032",
			ChainType:      "ETHEREUM",
			OperationType:  "CALL",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{"data": "0x"},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440033",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetNonce", mock.Anything, mock.AnythingOfType("string")).Return(uint64(15), nil)
		mockRPC.On("GetGasPrice", mock.Anything).Return(big.NewInt(30000000000), nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("execute QUERY operation successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440034",
			ChainType:      "ETHEREUM",
			OperationType:  "QUERY",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440035",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		mockRepo.AssertExpectations(t)
	})

	t.Run("execute APPROVE operation successfully", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440036",
			ChainType:      "ETHEREUM",
			OperationType:  "APPROVE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{"amount": "1000"},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440037",
		}

		mockRepo.On("GetByIdempotencyKey", mock.Anything, req.IdempotencyKey).Return(nil, errors.New("not found"))
		mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*entities.EVMTransaction")).Return(nil).Times(2)
		mockRPC.On("GetNonce", mock.Anything, mock.AnythingOfType("string")).Return(uint64(20), nil)
		mockRPC.On("GetGasPrice", mock.Anything).Return(big.NewInt(25000000000), nil)

		resp, err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		mockRepo.AssertExpectations(t)
		mockRPC.AssertExpectations(t)
	})

	t.Run("fail_with_invalid_operation_type", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440038",
			ChainType:      "ETHEREUM",
			OperationType:  "INVALID_OPERATION",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440039",
		}

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("fail_with_invalid_from_address", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440040",
			ChainType:      "ETHEREUM",
			OperationType:  "GET_BALANCE",
			FromAddress:    "invalid_address",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440041",
		}

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("fail_with_invalid_to_address", func(t *testing.T) {
		mockRPC := new(MockRPCClient)
		mockRepo := new(MockTransactionRepository)

		rpcClients := map[string]rpc.RPCClient{"ETHEREUM": mockRPC}
		useCase := NewExecuteEVMTransactionUseCase(rpcClients, mockRepo, logger)

		req := &dtos.ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440042",
			ChainType:      "ETHEREUM",
			OperationType:  "TRANSFER",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "invalid_address",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440043",
		}

		resp, err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Nil(t, resp)
	})
}
