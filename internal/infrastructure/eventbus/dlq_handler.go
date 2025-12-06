package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
)

// DLQSender interface para enviar mensagens para Dead Letter Queue
type DLQSender interface {
	SendMessage(ctx context.Context, message *Message, reason string) error
}

// DLQHandler gerencia a Dead Letter Queue para mensagens com falha
type DLQHandler struct {
	sqsClient SQSClient
	dlqURL    string
	logger    *zap.Logger
}

// NewDLQHandler cria um novo gerenciador de DLQ
func NewDLQHandler(sqsClient SQSClient, dlqURL string, logger *zap.Logger) *DLQHandler {
	return &DLQHandler{
		sqsClient: sqsClient,
		dlqURL:    dlqURL,
		logger:    logger,
	}
}

// DeadLetterMessage representa uma mensagem na DLQ com metadados
type DeadLetterMessage struct {
	OriginalMessage *Message
	ReceiptHandle   *string
	Reason          string
	RetryCount      int
	Timestamp       string
}

// SendMessage envia uma mensagem para a Dead Letter Queue
func (h *DLQHandler) SendMessage(ctx context.Context, message *Message, reason string) error {
	// Enriquecer com informação de falha
	enrichedPayload := map[string]interface{}{
		"original_message": message,
		"failure_reason":   reason,
	}

	bodyBytes, err := json.Marshal(enrichedPayload)
	if err != nil {
		h.logger.Error("failed to marshal DLQ message", zap.Error(err))
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    &h.dlqURL,
		MessageBody: func() *string { s := string(bodyBytes); return &s }(),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"OperationID": {
				DataType:    func() *string { s := "String"; return &s }(),
				StringValue: &message.OperationID,
			},
			"FailureReason": {
				DataType:    func() *string { s := "String"; return &s }(),
				StringValue: func() *string { return &reason }(),
			},
		},
	}

	_, err = h.sqsClient.SendMessage(ctx, input)
	if err != nil {
		h.logger.Error("failed to send message to DLQ",
			zap.String("operation_id", message.OperationID),
			zap.Error(err))
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	h.logger.Info("message sent to DLQ",
		zap.String("operation_id", message.OperationID),
		zap.String("reason", reason))

	return nil
}

// GetDeadLetterMessages recupera mensagens da DLQ
func (h *DLQHandler) GetDeadLetterMessages(ctx context.Context, maxMessages int32) ([]types.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:              &h.dlqURL,
		MaxNumberOfMessages:   maxMessages,
		WaitTimeSeconds:       20,
		VisibilityTimeout:     300,
		MessageAttributeNames: []string{"All"},
	}

	result, err := h.sqsClient.ReceiveMessage(ctx, input)
	if err != nil {
		h.logger.Error("failed to receive messages from DLQ", zap.Error(err))
		return nil, fmt.Errorf("failed to receive messages from DLQ: %w", err)
	}

	h.logger.Info("received messages from DLQ", zap.Int("count", len(result.Messages)))
	return result.Messages, nil
}

// DeleteDeadLetterMessage remove uma mensagem da DLQ
func (h *DLQHandler) DeleteDeadLetterMessage(ctx context.Context, receiptHandle *string) error {
	if receiptHandle == nil {
		return fmt.Errorf("receipt handle cannot be nil")
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      &h.dlqURL,
		ReceiptHandle: receiptHandle,
	}

	_, err := h.sqsClient.DeleteMessage(ctx, input)
	if err != nil {
		h.logger.Error("failed to delete message from DLQ", zap.Error(err))
		return fmt.Errorf("failed to delete message from DLQ: %w", err)
	}

	h.logger.Info("message deleted from DLQ")
	return nil
}

// GetDeadLetterMessageCount retorna o número aproximado de mensagens na DLQ
func (h *DLQHandler) GetDeadLetterMessageCount(ctx context.Context) (int32, error) {
	input := &sqs.GetQueueAttributesInput{
		QueueUrl:       &h.dlqURL,
		AttributeNames: []types.QueueAttributeName{"All"},
	}

	result, err := h.sqsClient.GetQueueAttributes(ctx, input)
	if err != nil {
		h.logger.Error("failed to get DLQ attributes", zap.Error(err))
		return 0, fmt.Errorf("failed to get DLQ attributes: %w", err)
	}

	countStr, ok := result.Attributes["ApproximateNumberOfMessages"]
	if !ok {
		h.logger.Warn("ApproximateNumberOfMessages attribute not found in response")
		return 0, fmt.Errorf("ApproximateNumberOfMessages attribute not found")
	}

	count, err := strconv.ParseInt(countStr, 10, 32)
	if err != nil {
		h.logger.Error("failed to parse message count", zap.Error(err))
		return 0, fmt.Errorf("failed to parse message count: %w", err)
	}

	return int32(count), nil
}
