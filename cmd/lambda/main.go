package main

// Lambda handler for EVM transaction execution
// Updated: 2025-12-04 23:10 - Testing E2E with full IAM permissions

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gabrielksneiva/ChainEVM/internal/application/dtos"
	"github.com/gabrielksneiva/ChainEVM/internal/application/usecases"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/database"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/eventbus"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/logger"
	"github.com/gabrielksneiva/ChainEVM/internal/infrastructure/rpc"
	pkgconfig "github.com/gabrielksneiva/ChainEVM/pkg/config"
	"go.uber.org/zap"
)

var (
	cfg            *pkgconfig.Config
	log            *zap.Logger
	executeUseCase *usecases.ExecuteEVMTransactionUseCase
	sqsConsumer    *eventbus.SQSConsumer
)

func init() {
	var err error

	// Load config
	cfg = pkgconfig.LoadConfig()

	// Initialize logger
	log, err = logger.NewLogger(cfg.Environment)
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	// Initialize AWS config
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal("failed to load AWS config", zap.Error(err))
	}

	// Initialize SQS client
	sqsClient := sqs.NewFromConfig(awsCfg)
	sqsAdapter := eventbus.NewSQSAdapter(sqsClient)
	sqsConsumer = eventbus.NewSQSConsumer(sqsAdapter, cfg.SQSQueueURL, log)

	// Initialize DynamoDB client
	dynamoDBClient := dynamodb.NewFromConfig(awsCfg)
	dynamoDBAdapter := database.NewDynamoDBAdapter(dynamoDBClient)
	transactionRepo := database.NewDynamoDBTransactionRepository(dynamoDBAdapter, cfg.DynamoDBTableName, log)

	// Initialize RPC clients for each chain
	rpcClients := make(map[string]rpc.RPCClient)
	for chainName, rpcURL := range cfg.EVMRPCURLs {
		if rpcURL != "" {
			client, err := rpc.NewEVMRPCClient(rpcURL, cfg.RPCTimeout, log)
			if err != nil {
				log.Warn("failed to initialize RPC client for chain",
					zap.String("chain", chainName),
					zap.Error(err))
				continue
			}
			rpcClients[chainName] = client
		}
	}

	// Initialize use cases
	executeUseCase = usecases.NewExecuteEVMTransactionUseCase(rpcClients, transactionRepo, log)

	log.Info("Lambda function initialized successfully",
		zap.String("environment", cfg.Environment),
		zap.String("sqs_queue_url", cfg.SQSQueueURL),
		zap.String("dynamodb_table", cfg.DynamoDBTableName),
		zap.Int("rpc_clients_initialized", len(rpcClients)),
	)
}

func main() {
	lambda.Start(handler)
}

// handler processa eventos SQS
func handler(ctx context.Context, event events.SQSEvent) error {
	log.Info("processing SQS event", zap.Int("message_count", len(event.Records)))

	for _, record := range event.Records {
		if err := processMessage(ctx, record); err != nil {
			log.Error("failed to process message",
				zap.String("message_id", record.MessageId),
				zap.Error(err))
			// Não retorna erro para não descartar a mensagem automaticamente
			// A visibilidade será alterada para retry
		}
	}

	return nil
}

// processMessage processa uma única mensagem
func processMessage(ctx context.Context, record events.SQSMessage) error {
	log.Info("processing SQS message",
		zap.String("message_id", record.MessageId),
		zap.String("receipt_handle", record.ReceiptHandle))

	// Parse mensagem
	var msgBody eventbus.Message
	if err := json.Unmarshal([]byte(record.Body), &msgBody); err != nil {
		log.Error("failed to unmarshal message body", zap.Error(err))
		return err
	}

	// Converter para DTO
	req := &dtos.ExecuteTransactionRequest{
		OperationID:    msgBody.OperationID,
		ChainType:      msgBody.ChainType,
		OperationType:  msgBody.OperationType,
		FromAddress:    msgBody.FromAddress,
		ToAddress:      msgBody.ToAddress,
		Payload:        msgBody.Payload,
		IdempotencyKey: msgBody.IdempotencyKey,
	}

	// Executar transação
	response, err := executeUseCase.Execute(ctx, req)
	if err != nil {
		log.Error("failed to execute transaction",
			zap.String("operation_id", req.OperationID),
			zap.Error(err))
		return err
	}

	log.Info("transaction executed successfully",
		zap.String("operation_id", response.OperationID),
		zap.String("status", response.Status))

	// Deletar mensagem da fila após processamento bem-sucedido
	receiptHandle := record.ReceiptHandle
	if err := sqsConsumer.DeleteMessage(ctx, &receiptHandle); err != nil {
		log.Error("failed to delete message from SQS",
			zap.String("message_id", record.MessageId),
			zap.Error(err))
		return err
	}

	return nil
}
