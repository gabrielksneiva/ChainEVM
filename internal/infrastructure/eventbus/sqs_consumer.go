package eventbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// SQSConsumer consome mensagens da fila SQS
type SQSConsumer struct {
	sqsClient SQSClient
	queueURL  string
	logger    *zap.Logger
}

// NewSQSConsumer cria um novo consumidor SQS
func NewSQSConsumer(sqsClient SQSClient, queueURL string, logger *zap.Logger) *SQSConsumer {
	return &SQSConsumer{
		sqsClient: sqsClient,
		queueURL:  queueURL,
		logger:    logger,
	}
}

// Message estrutura das mensagens da fila
type Message struct {
	OperationID    string                 `json:"operation_id"`
	ChainType      string                 `json:"chain_type"`
	OperationType  string                 `json:"operation_type"`
	FromAddress    string                 `json:"from_address"`
	ToAddress      string                 `json:"to_address"`
	Payload        map[string]interface{} `json:"payload"`
	IdempotencyKey string                 `json:"idempotency_key"`
}

// ReceiveMessages recebe mensagens da fila SQS
func (c *SQSConsumer) ReceiveMessages(ctx context.Context, maxMessages int32) ([]types.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &c.queueURL,
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     20,
		VisibilityTimeout:   300,
	}

	result, err := c.sqsClient.ReceiveMessage(ctx, input)
	if err != nil {
		c.logger.Error("failed to receive messages from SQS", zap.Error(err))
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	c.logger.Info("received messages from SQS", zap.Int("count", len(result.Messages)))
	return result.Messages, nil
}

// ParseMessage transforma a mensagem em estrutura utilizável
func (c *SQSConsumer) ParseMessage(message types.Message) (*Message, error) {
	var msg Message
	if err := json.Unmarshal([]byte(*message.Body), &msg); err != nil {
		c.logger.Error("failed to parse SQS message", zap.Error(err))
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	return &msg, nil
}

// DeleteMessage deleta uma mensagem da fila
func (c *SQSConsumer) DeleteMessage(ctx context.Context, receiptHandle *string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      &c.queueURL,
		ReceiptHandle: receiptHandle,
	}

	_, err := c.sqsClient.DeleteMessage(ctx, input)
	if err != nil {
		c.logger.Error("failed to delete message from SQS", zap.Error(err))
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// ChangeMessageVisibility altera a visibilidade da mensagem (para retry)
func (c *SQSConsumer) ChangeMessageVisibility(ctx context.Context, receiptHandle *string, visibilityTimeout int32) error {
	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          &c.queueURL,
		ReceiptHandle:     receiptHandle,
		VisibilityTimeout: visibilityTimeout,
	}

	_, err := c.sqsClient.ChangeMessageVisibility(ctx, input)
	if err != nil {
		c.logger.Error("failed to change message visibility", zap.Error(err))
		return fmt.Errorf("failed to change message visibility: %w", err)
	}

	return nil
}
