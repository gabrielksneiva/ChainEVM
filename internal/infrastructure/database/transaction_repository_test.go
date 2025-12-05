package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/entities"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockDynamoDBClient mock do cliente DynamoDB
type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func TestNewDynamoDBTransactionRepository(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()

	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	assert.NotNil(t, repo)
}

func TestDynamoDBTransactionRepository_Save_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	opID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	opType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	toAddr, _ := valueobjects.NewEVMAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")

	tx := entities.NewEVMTransaction(opID, chainType, opType, fromAddr, toAddr, nil, "idem123")

	mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(&dynamodb.PutItemOutput{}, nil)

	err := repo.Save(context.Background(), tx)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_Save_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	opID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	opType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	toAddr, _ := valueobjects.NewEVMAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")

	tx := entities.NewEVMTransaction(opID, chainType, opType, fromAddr, toAddr, nil, "idem123")

	mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(nil, errors.New("dynamodb error"))

	err := repo.Save(context.Background(), tx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save transaction")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByOperationID_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	now := time.Now()
	mockOutput := &dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"operation_id":    &types.AttributeValueMemberS{Value: "550e8400-e29b-41d4-a716-446655440000"},
			"chain_type":      &types.AttributeValueMemberS{Value: "ETHEREUM"},
			"operation_type":  &types.AttributeValueMemberS{Value: "TRANSFER"},
			"from_address":    &types.AttributeValueMemberS{Value: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"},
			"to_address":      &types.AttributeValueMemberS{Value: "0x8ba1f109551bD432803012645Ac136ddd64DBA72"},
			"status":          &types.AttributeValueMemberS{Value: "PENDING"},
			"created_at":      &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"idempotency_key": &types.AttributeValueMemberS{Value: "idem123"},
		},
	}

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(mockOutput, nil)

	tx, err := repo.GetByOperationID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", tx.OperationID().String())
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByOperationID_NotFound(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockOutput := &dynamodb.GetItemOutput{
		Item: nil,
	}

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(mockOutput, nil)

	tx, err := repo.GetByOperationID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "transaction not found")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByOperationID_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(nil, errors.New("dynamodb error"))

	tx, err := repo.GetByOperationID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	assert.Error(t, err)
	assert.Nil(t, tx)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByIdempotencyKey_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	now := time.Now()
	mockOutput := &dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"operation_id":    &types.AttributeValueMemberS{Value: "550e8400-e29b-41d4-a716-446655440000"},
				"chain_type":      &types.AttributeValueMemberS{Value: "ETHEREUM"},
				"operation_type":  &types.AttributeValueMemberS{Value: "TRANSFER"},
				"from_address":    &types.AttributeValueMemberS{Value: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0"},
				"to_address":      &types.AttributeValueMemberS{Value: "0x8ba1f109551bD432803012645Ac136ddd64DBA72"},
				"status":          &types.AttributeValueMemberS{Value: "PENDING"},
				"created_at":      &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
				"idempotency_key": &types.AttributeValueMemberS{Value: "idem123"},
			},
		},
	}

	mockClient.On("Query", mock.Anything, mock.AnythingOfType("*dynamodb.QueryInput")).
		Return(mockOutput, nil)

	tx, err := repo.GetByIdempotencyKey(context.Background(), "idem123")

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", tx.OperationID().String())
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByIdempotencyKey_NotFound(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockOutput := &dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{},
	}

	mockClient.On("Query", mock.Anything, mock.AnythingOfType("*dynamodb.QueryInput")).
		Return(mockOutput, nil)

	tx, err := repo.GetByIdempotencyKey(context.Background(), "idem123")

	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.Contains(t, err.Error(), "transaction not found")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByIdempotencyKey_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockClient.On("Query", mock.Anything, mock.AnythingOfType("*dynamodb.QueryInput")).
		Return(nil, errors.New("dynamodb error"))

	tx, err := repo.GetByIdempotencyKey(context.Background(), "idem123")

	assert.Error(t, err)
	assert.Nil(t, tx)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_UpdateStatus_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).
		Return(&dynamodb.UpdateItemOutput{}, nil)

	err := repo.UpdateStatus(context.Background(), "550e8400-e29b-41d4-a716-446655440000", entities.TransactionStatusSuccess)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_UpdateStatus_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput")).
		Return(nil, errors.New("dynamodb error"))

	err := repo.UpdateStatus(context.Background(), "550e8400-e29b-41d4-a716-446655440000", entities.TransactionStatusSuccess)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update status")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_Save_WithExecutedAt(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	opID, _ := valueobjects.NewOperationID("550e8400-e29b-41d4-a716-446655440000")
	chainType, _ := valueobjects.NewChainType("ETHEREUM")
	opType, _ := valueobjects.NewOperationType("TRANSFER")
	fromAddr, _ := valueobjects.NewEVMAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	toAddr, _ := valueobjects.NewEVMAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")
	txHash, _ := valueobjects.NewTransactionHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")

	tx := entities.NewEVMTransaction(opID, chainType, opType, fromAddr, toAddr, nil, "idem123")
	tx.MarkAsSuccess(txHash, 100000, 2100000)

	mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(&dynamodb.PutItemOutput{}, nil)

	err := repo.Save(context.Background(), tx)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByOperationID_UnmarshalError(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockOutput := &dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"operation_id": &types.AttributeValueMemberS{Value: "invalid-uuid"},
		},
	}

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(mockOutput, nil)

	tx, err := repo.GetByOperationID(context.Background(), "550e8400-e29b-41d4-a716-446655440000")

	assert.Error(t, err)
	assert.Nil(t, tx)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBTransactionRepository_GetByIdempotencyKey_UnmarshalError(t *testing.T) {
	t.Parallel()

	mockClient := new(MockDynamoDBClient)
	logger, _ := zap.NewDevelopment()
	repo := NewDynamoDBTransactionRepository(mockClient, "test-table", logger)

	mockOutput := &dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"operation_id": &types.AttributeValueMemberS{Value: "invalid-uuid"},
			},
		},
	}

	mockClient.On("Query", mock.Anything, mock.AnythingOfType("*dynamodb.QueryInput")).
		Return(mockOutput, nil)

	tx, err := repo.GetByIdempotencyKey(context.Background(), "idem123")

	assert.Error(t, err)
	assert.Nil(t, tx)
	mockClient.AssertExpectations(t)
}
