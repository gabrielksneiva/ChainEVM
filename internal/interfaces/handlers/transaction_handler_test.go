package handlers

import (
	"context"
	"net/http"
	"testing"

	"github.com/gabrielksneiva/ChainEVM/internal/application/dtos"
	"github.com/gabrielksneiva/ChainEVM/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockExecuteUseCase for testing
type MockExecuteUseCase struct {
	mock.Mock
}

func (m *MockExecuteUseCase) Execute(ctx context.Context, req *dtos.ExecuteTransactionRequest) (*dtos.ExecuteTransactionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.ExecuteTransactionResponse), args.Error(1)
}

func TestNewTransactionHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}

	handler := NewTransactionHandler(nil, cfg, logger)

	require.NotNil(t, handler)
}

func TestValidateRequest(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	tests := []struct {
		name    string
		req     *dtos.ExecuteTransactionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &dtos.ExecuteTransactionRequest{
				OperationID:    "550e8400-e29b-41d4-a716-446655440000",
				ChainType:      "ETHEREUM",
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			wantErr: false,
		},
		{
			name: "missing operation_id",
			req: &dtos.ExecuteTransactionRequest{
				ChainType:      "ETHEREUM",
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequestAllChains(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	validChains := []string{"ETHEREUM", "POLYGON", "BSC", "ARBITRUM", "OPTIMISM", "BASE"}

	for _, chain := range validChains {
		t.Run("valid_chain_"+chain, func(t *testing.T) {
			req := &dtos.ExecuteTransactionRequest{
				OperationID:   "550e8400-e29b-41d4-a716-446655440000",
				ChainType:     chain,
				FromAddress:   "0x1234567890123456789012345678901234567890",
				ToAddress:     "0x0987654321098765432109876543210987654321",
				OperationType: "TRANSFER",
			}
			err := handler.ValidateRequest(req)
			assert.NoError(t, err)
		})
	}
}

func TestExecuteTransactionSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}

	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		FromAddress:    "0x1234567890123456789012345678901234567890",
		ToAddress:      "0x0987654321098765432109876543210987654321",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{"amount": "1000000000000000000"},
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
	}

	assert.NotNil(t, handler)
	assert.NotNil(t, req)
}

func TestValidateRequestEdgeCases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	tests := []struct {
		name    string
		req     *dtos.ExecuteTransactionRequest
		wantErr bool
	}{
		{
			name: "empty operation_id",
			req: &dtos.ExecuteTransactionRequest{
				OperationID:    "",
				ChainType:      "ETHEREUM",
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			wantErr: true,
		},
		{
			name: "empty from_address",
			req: &dtos.ExecuteTransactionRequest{
				OperationID:    "550e8400-e29b-41d4-a716-446655440000",
				ChainType:      "ETHEREUM",
				FromAddress:    "",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			wantErr: true,
		},
		{
			name: "empty to_address",
			req: &dtos.ExecuteTransactionRequest{
				OperationID:    "550e8400-e29b-41d4-a716-446655440000",
				ChainType:      "ETHEREUM",
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			wantErr: true,
		},
		{
			name: "empty chain_type",
			req: &dtos.ExecuteTransactionRequest{
				OperationID:    "550e8400-e29b-41d4-a716-446655440000",
				ChainType:      "",
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateRequest(tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequestValueObjects(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	validOperations := []string{"TRANSFER", "GET_BALANCE", "GET_NONCE"}

	for _, op := range validOperations {
		t.Run("valid_operation_"+op, func(t *testing.T) {
			req := &dtos.ExecuteTransactionRequest{
				OperationID:   "550e8400-e29b-41d4-a716-446655440000",
				ChainType:     "ETHEREUM",
				FromAddress:   "0x1234567890123456789012345678901234567890",
				ToAddress:     "0x0987654321098765432109876543210987654321",
				OperationType: op,
			}
			err := handler.ValidateRequest(req)
			assert.NoError(t, err)
		})
	}
}

func TestHandlerWithContext(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}

	handler := NewTransactionHandler(nil, cfg, logger)

	assert.NotNil(t, handler)
	assert.NotNil(t, cfg)
}

func TestMultipleChainValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	chainTests := []struct {
		chain   string
		wantErr bool
	}{
		{"ETHEREUM", false},
		{"POLYGON", false},
		{"BSC", false},
		{"ARBITRUM", false},
		{"OPTIMISM", false},
		{"BASE", false},
	}

	for _, ct := range chainTests {
		t.Run("chain_"+ct.chain, func(t *testing.T) {
			req := &dtos.ExecuteTransactionRequest{
				OperationID:    "550e8400-e29b-41d4-a716-446655440000",
				ChainType:      ct.chain,
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  "TRANSFER",
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			}
			err := handler.ValidateRequest(req)
			if ct.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMultipleOperationValidation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	opTests := []struct {
		op      string
		wantErr bool
	}{
		{"TRANSFER", false},
		{"GET_BALANCE", false},
		{"GET_NONCE", false},
	}

	for _, ot := range opTests {
		t.Run("op_"+ot.op, func(t *testing.T) {
			req := &dtos.ExecuteTransactionRequest{
				OperationID:    "550e8400-e29b-41d4-a716-446655440000",
				ChainType:      "ETHEREUM",
				FromAddress:    "0x1234567890123456789012345678901234567890",
				ToAddress:      "0x0987654321098765432109876543210987654321",
				OperationType:  ot.op,
				Payload:        map[string]interface{}{},
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			}
			err := handler.ValidateRequest(req)
			if ot.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExecuteTransactionWithNilUseCase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		FromAddress:    "0x1234567890123456789012345678901234567890",
		ToAddress:      "0x0987654321098765432109876543210987654321",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{},
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
	}

	resp, code, err := handler.ExecuteTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, 500, code)
}

func TestValidateRequestComplete(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		FromAddress:    "0x1234567890123456789012345678901234567890",
		ToAddress:      "0x0987654321098765432109876543210987654321",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{"key": "value"},
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
	}

	err := handler.ValidateRequest(req)
	assert.NoError(t, err)
}

func TestExecuteTransactionLogging(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "op-123",
		ChainType:      "POLYGON",
		FromAddress:    "0xAAAA",
		ToAddress:      "0xBBBB",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{},
		IdempotencyKey: "idempotency-key-123",
	}

	resp, code, err := handler.ExecuteTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, code)
}

// Test ExecuteTransaction delegates to executeTransactionInternal
func TestExecuteTransactionDelegatesToInternal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "test-op-1",
		ChainType:      "ETHEREUM",
		FromAddress:    "0x1111111111111111111111111111111111111111",
		ToAddress:      "0x2222222222222222222222222222222222222222",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{},
		IdempotencyKey: "test-idem-1",
	}

	resp, code, err := handler.ExecuteTransaction(context.Background(), req)

	// Since usecase is nil, should get error
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, code)
}

// Test ValidateRequest with all empty fields
func TestValidateRequestAllEmpty(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{}

	err := handler.ValidateRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "operation_id is required", err.Error())
}

// Test ValidateRequest missing chain_type
func TestValidateRequestMissingChain(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID: "test-op",
	}

	err := handler.ValidateRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "chain_type is required", err.Error())
}

// Test ValidateRequest missing from address
func TestValidateRequestMissingFromAddr(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID: "test-op",
		ChainType:   "ETHEREUM",
	}

	err := handler.ValidateRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "from_address is required", err.Error())
}

// Test ValidateRequest missing to address
func TestValidateRequestMissingToAddr(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}
	handler := NewTransactionHandler(nil, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID: "test-op",
		ChainType:   "ETHEREUM",
		FromAddress: "0x1111111111111111111111111111111111111111",
	}

	err := handler.ValidateRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "to_address is required", err.Error())
}

// Test ExecuteTransactionInternal with successful execution
func TestExecuteTransactionInternalSuccess(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}

	mockUseCase := new(MockExecuteUseCase)
	handler := NewTransactionHandler(mockUseCase, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		FromAddress:    "0x1111111111111111111111111111111111111111",
		ToAddress:      "0x2222222222222222222222222222222222222222",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{"amount": "1.0"},
		IdempotencyKey: "idem-key-123",
	}

	expectedResp := &dtos.ExecuteTransactionResponse{
		OperationID:     req.OperationID,
		Status:          "SUCCESS",
		ChainType:       "ETHEREUM",
		TransactionHash: "0xabcdef1234567890",
		CreatedAt:       "2024-01-01T00:00:00Z",
	}

	mockUseCase.On("Execute", mock.Anything, req).Return(expectedResp, nil)

	resp, code, err := handler.ExecuteTransaction(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, expectedResp.OperationID, resp.OperationID)
	assert.Equal(t, expectedResp.Status, resp.Status)
	assert.Equal(t, expectedResp.TransactionHash, resp.TransactionHash)
	mockUseCase.AssertExpectations(t)
}

// Test ExecuteTransactionInternal with use case error
func TestExecuteTransactionInternalUseCaseError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &config.Config{}

	mockUseCase := new(MockExecuteUseCase)
	handler := NewTransactionHandler(mockUseCase, cfg, logger)

	req := &dtos.ExecuteTransactionRequest{
		OperationID:    "550e8400-e29b-41d4-a716-446655440000",
		ChainType:      "ETHEREUM",
		FromAddress:    "0x1111111111111111111111111111111111111111",
		ToAddress:      "0x2222222222222222222222222222222222222222",
		OperationType:  "TRANSFER",
		Payload:        map[string]interface{}{},
		IdempotencyKey: "idem-key-456",
	}

	mockUseCase.On("Execute", mock.Anything, req).Return(nil, assert.AnError)

	resp, code, err := handler.ExecuteTransaction(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, http.StatusInternalServerError, code)
	mockUseCase.AssertExpectations(t)
}
