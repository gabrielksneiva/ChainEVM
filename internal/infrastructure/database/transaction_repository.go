package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/entities"
	"github.com/gabrielksneiva/ChainEVM/internal/domain/valueobjects"
	"go.uber.org/zap"
)

// TransactionRepository interface para persistência de transações
type TransactionRepository interface {
	Save(ctx context.Context, tx *entities.EVMTransaction) error
	GetByOperationID(ctx context.Context, operationID string) (*entities.EVMTransaction, error)
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entities.EVMTransaction, error)
	UpdateStatus(ctx context.Context, operationID string, status entities.TransactionStatus) error
}

// DynamoDBTransactionRepository implementação usando DynamoDB
type DynamoDBTransactionRepository struct {
	dynamoDBClient DynamoDBClient
	tableName      string
	logger         *zap.Logger
}

// NewDynamoDBTransactionRepository cria um novo repositório DynamoDB
func NewDynamoDBTransactionRepository(
	dynamoDBClient DynamoDBClient,
	tableName string,
	logger *zap.Logger,
) TransactionRepository {
	return &DynamoDBTransactionRepository{
		dynamoDBClient: dynamoDBClient,
		tableName:      tableName,
		logger:         logger,
	}
}

// TransactionItem estrutura para armazenar no DynamoDB
type TransactionItem struct {
	OperationID     string  `dynamodbav:"operation_id"`
	IdempotencyKey  string  `dynamodbav:"idempotency_key"`
	ChainType       string  `dynamodbav:"chain_type"`
	OperationType   string  `dynamodbav:"operation_type"`
	FromAddress     string  `dynamodbav:"from_address"`
	ToAddress       string  `dynamodbav:"to_address"`
	Status          string  `dynamodbav:"status"`
	TransactionHash string  `dynamodbav:"transaction_hash,omitempty"`
	BlockNumber     *int64  `dynamodbav:"block_number,omitempty"`
	GasUsed         *int64  `dynamodbav:"gas_used,omitempty"`
	GasPrice        *string `dynamodbav:"gas_price,omitempty"`
	ErrorMessage    string  `dynamodbav:"error_message,omitempty"`
	CreatedAt       string  `dynamodbav:"created_at"`
	ExecutedAt      *string `dynamodbav:"executed_at,omitempty"`
	TTL             int64   `dynamodbav:"ttl"`
}

// Save persiste uma transação
func (r *DynamoDBTransactionRepository) Save(ctx context.Context, tx *entities.EVMTransaction) error {
	item := TransactionItem{
		OperationID:     tx.OperationID().String(),
		IdempotencyKey:  tx.IdempotencyKey(),
		ChainType:       tx.ChainType().String(),
		OperationType:   tx.OperationType().String(),
		FromAddress:     tx.FromAddress().String(),
		ToAddress:       tx.ToAddress().String(),
		Status:          string(tx.Status()),
		TransactionHash: tx.TxHash().String(),
		BlockNumber:     tx.BlockNumber(),
		GasUsed:         tx.GasUsed(),
		GasPrice:        tx.GasPrice(),
		ErrorMessage:    tx.ErrorMessage(),
		CreatedAt:       tx.CreatedAt().Format("2006-01-02T15:04:05Z"),
		ExecutedAt:      nil,
		TTL:             7776000, // 90 dias em segundos
	}

	if tx.ExecutedAt() != nil {
		executedAt := tx.ExecutedAt().Format("2006-01-02T15:04:05Z")
		item.ExecutedAt = &executedAt
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		r.logger.Error("failed to marshal transaction item", zap.Error(err))
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: &r.tableName,
	}

	_, err = r.dynamoDBClient.PutItem(ctx, input)
	if err != nil {
		r.logger.Error("failed to save transaction to DynamoDB",
			zap.String("operation_id", tx.OperationID().String()),
			zap.Error(err))
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	r.logger.Info("transaction saved successfully",
		zap.String("operation_id", tx.OperationID().String()))
	return nil
}

// GetByOperationID recupera uma transação pelo operation ID
func (r *DynamoDBTransactionRepository) GetByOperationID(ctx context.Context, operationID string) (*entities.EVMTransaction, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"operation_id": &types.AttributeValueMemberS{Value: operationID},
		},
		TableName: &r.tableName,
	}

	result, err := r.dynamoDBClient.GetItem(ctx, input)
	if err != nil {
		r.logger.Error("failed to get transaction from DynamoDB",
			zap.String("operation_id", operationID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("transaction not found")
	}

	var item TransactionItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		r.logger.Error("failed to unmarshal transaction item", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return unmarshalTransactionItem(item, r.logger)
} // GetByIdempotencyKey recupera uma transação pelo idempotency key
func (r *DynamoDBTransactionRepository) GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (*entities.EVMTransaction, error) {
	input := &dynamodb.QueryInput{
		TableName:              &r.tableName,
		IndexName:              stringPtr("idempotency_key-index"),
		KeyConditionExpression: stringPtr("idempotency_key = :key"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":key": &types.AttributeValueMemberS{Value: idempotencyKey},
		},
	}

	result, err := r.dynamoDBClient.Query(ctx, input)
	if err != nil {
		r.logger.Error("failed to query transaction by idempotency key",
			zap.String("idempotency_key", idempotencyKey),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query transaction: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("transaction not found")
	}

	var item TransactionItem
	err = attributevalue.UnmarshalMap(result.Items[0], &item)
	if err != nil {
		r.logger.Error("failed to unmarshal transaction item", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return unmarshalTransactionItem(item, r.logger)
}

// UpdateStatus atualiza o status de uma transação
func (r *DynamoDBTransactionRepository) UpdateStatus(ctx context.Context, operationID string, status entities.TransactionStatus) error {
	input := &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"operation_id": &types.AttributeValueMemberS{Value: operationID},
		},
		UpdateExpression: stringPtr("SET #status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
		TableName: &r.tableName,
	}

	_, err := r.dynamoDBClient.UpdateItem(ctx, input)
	if err != nil {
		r.logger.Error("failed to update transaction status",
			zap.String("operation_id", operationID),
			zap.Error(err))
		return fmt.Errorf("failed to update status: %w", err)
	}

	r.logger.Info("transaction status updated",
		zap.String("operation_id", operationID),
		zap.String("status", string(status)))
	return nil
}

func stringPtr(s string) *string {
	return &s
}

// unmarshalTransactionItem converte um TransactionItem em EVMTransaction
func unmarshalTransactionItem(item TransactionItem, logger *zap.Logger) (*entities.EVMTransaction, error) {
	operationID, err := valueobjects.NewOperationID(item.OperationID)
	if err != nil {
		logger.Error("failed to parse operation ID", zap.Error(err))
		return nil, err
	}

	chainType, err := valueobjects.NewChainType(item.ChainType)
	if err != nil {
		logger.Error("failed to parse chain type", zap.Error(err))
		return nil, err
	}

	operationType, err := valueobjects.NewOperationType(item.OperationType)
	if err != nil {
		logger.Error("failed to parse operation type", zap.Error(err))
		return nil, err
	}

	fromAddr, err := valueobjects.NewEVMAddress(item.FromAddress)
	if err != nil {
		logger.Error("failed to parse from address", zap.Error(err))
		return nil, err
	}

	toAddr, err := valueobjects.NewEVMAddress(item.ToAddress)
	if err != nil {
		logger.Error("failed to parse to address", zap.Error(err))
		return nil, err
	}

	tx := entities.NewEVMTransaction(
		operationID,
		chainType,
		operationType,
		fromAddr,
		toAddr,
		make(map[string]interface{}),
		item.IdempotencyKey,
	)

	// Restaurar estado
	status := entities.TransactionStatus(item.Status)
	switch status {
	case entities.TransactionStatusPending:
		// Estado inicial
	case entities.TransactionStatusProcessing:
		tx.MarkAsProcessing()
	case entities.TransactionStatusSuccess:
		if item.TransactionHash != "" {
			txHash, _ := valueobjects.NewTransactionHash(item.TransactionHash)
			blockNum := int64(0)
			if item.BlockNumber != nil {
				blockNum = *item.BlockNumber
			}
			gasUsedVal := int64(0)
			if item.GasUsed != nil {
				gasUsedVal = *item.GasUsed
			}
			tx.MarkAsSuccess(txHash, blockNum, gasUsedVal)
		}
	case entities.TransactionStatusConfirmed:
		tx.MarkAsConfirmed()
	case entities.TransactionStatusFailed:
		tx.MarkAsFailed(item.ErrorMessage)
	}

	return tx, nil
}
