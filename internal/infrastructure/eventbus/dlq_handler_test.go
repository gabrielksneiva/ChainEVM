package eventbus

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// mockSQSClient para testes de DLQ
type mockSQSClient struct {
	mock.Mock
}

func (m *mockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *mockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *mockSQSClient) ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ChangeMessageVisibilityOutput), args.Error(1)
}

func (m *mockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (m *mockSQSClient) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.GetQueueAttributesOutput), args.Error(1)
}

// TestDLQHandler_SendMessageSuccess testa envio bem-sucedido para DLQ
func TestDLQHandler_SendMessageSuccess(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	originalMessage := &Message{
		OperationID:    "op-123",
		ChainType:      "sepolia",
		OperationType:  "TRANSFER",
		FromAddress:    "0x1234",
		ToAddress:      "0x5678",
		IdempotencyKey: "idempotency-123",
	}

	msgID := "msg-123"
	mockSQS.On("SendMessage", mock.Anything, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
		return input.QueueUrl != nil && *input.QueueUrl == dlqURL
	})).Return(&sqs.SendMessageOutput{
		MessageId: &msgID,
	}, nil)

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	err := dlqHandler.SendMessage(context.Background(), originalMessage, "max retries exceeded")

	// Assert
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

// TestDLQHandler_SendMessageFailure testa falha ao enviar para DLQ
func TestDLQHandler_SendMessageFailure(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	originalMessage := &Message{
		OperationID:   "op-123",
		ChainType:     "sepolia",
		OperationType: "TRANSFER",
		FromAddress:   "0x1234",
		ToAddress:     "0x5678",
	}

	mockSQS.On("SendMessage", mock.Anything, mock.Anything).Return(nil, errors.New("DLQ send failed"))

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	err2 := dlqHandler.SendMessage(context.Background(), originalMessage, "some reason")

	// Assert
	assert.Error(t, err2)
	assert.Equal(t, "failed to send message to DLQ: DLQ send failed", err2.Error())
}

// TestDLQHandler_GetDeadLetterMessages testa recuperação de mensagens DLQ
func TestDLQHandler_GetDeadLetterMessages(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	messageID1 := "msg-1"
	messageID2 := "msg-2"
	body1 := `{"operation_id":"op-1","chain_type":"sepolia"}`
	body2 := `{"operation_id":"op-2","chain_type":"ethereum"}`
	receiptHandle1 := "receipt-1"
	receiptHandle2 := "receipt-2"

	dlqMessages := []types.Message{
		{
			MessageId:     &messageID1,
			Body:          &body1,
			ReceiptHandle: &receiptHandle1,
			Attributes:    map[string]string{"ApproximateReceiveCount": "5"},
		},
		{
			MessageId:     &messageID2,
			Body:          &body2,
			ReceiptHandle: &receiptHandle2,
			Attributes:    map[string]string{"ApproximateReceiveCount": "10"},
		},
	}

	mockSQS.On("ReceiveMessage", mock.Anything, mock.MatchedBy(func(input *sqs.ReceiveMessageInput) bool {
		return input.QueueUrl != nil && *input.QueueUrl == dlqURL
	})).Return(&sqs.ReceiveMessageOutput{
		Messages: dlqMessages,
	}, nil)

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	messages, err := dlqHandler.GetDeadLetterMessages(context.Background(), 10)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, len(messages))
	assert.Equal(t, messageID1, *messages[0].MessageId)
	assert.Equal(t, messageID2, *messages[1].MessageId)
}

// TestDLQHandler_GetDeadLetterMessages_ReceiveError testa erro ao receber mensagens
func TestDLQHandler_GetDeadLetterMessages_ReceiveError(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	mockSQS.On("ReceiveMessage", mock.Anything, mock.Anything).Return(nil, errors.New("receive failed"))

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	messages, err := dlqHandler.GetDeadLetterMessages(context.Background(), 10)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, messages)
	assert.Equal(t, "failed to receive messages from DLQ: receive failed", err.Error())
}

// TestDLQHandler_DeleteDeadLetterMessage testa exclusão de mensagem da DLQ
func TestDLQHandler_DeleteDeadLetterMessage(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	receiptHandle := "receipt-handle-123"

	mockSQS.On("DeleteMessage", mock.Anything, mock.MatchedBy(func(input *sqs.DeleteMessageInput) bool {
		return input.QueueUrl != nil && *input.QueueUrl == dlqURL && input.ReceiptHandle != nil && *input.ReceiptHandle == receiptHandle
	})).Return(&sqs.DeleteMessageOutput{}, nil)

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	err := dlqHandler.DeleteDeadLetterMessage(context.Background(), &receiptHandle)

	// Assert
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

// TestDLQHandler_DeleteDeadLetterMessage_Error testa erro ao deletar mensagem DLQ
func TestDLQHandler_DeleteDeadLetterMessage_Error(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	receiptHandle := "receipt-handle-123"

	mockSQS.On("DeleteMessage", mock.Anything, mock.Anything).Return(nil, errors.New("delete failed"))

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	err := dlqHandler.DeleteDeadLetterMessage(context.Background(), &receiptHandle)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "failed to delete message from DLQ: delete failed", err.Error())
}

// TestDLQHandler_GetDeadLetterMessageCount testa contagem de mensagens DLQ
func TestDLQHandler_GetDeadLetterMessageCount(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	mockSQS.On("GetQueueAttributes", mock.Anything, mock.MatchedBy(func(input *sqs.GetQueueAttributesInput) bool {
		return input.QueueUrl != nil && *input.QueueUrl == dlqURL
	})).Return(&sqs.GetQueueAttributesOutput{
		Attributes: map[string]string{
			"ApproximateNumberOfMessages": "42",
		},
	}, nil)

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	count, err := dlqHandler.GetDeadLetterMessageCount(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int32(42), count)
}

// TestDLQHandler_GetDeadLetterMessageCount_Error testa erro ao obter contagem
func TestDLQHandler_GetDeadLetterMessageCount_Error(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	mockSQS.On("GetQueueAttributes", mock.Anything, mock.Anything).Return(nil, errors.New("get attributes failed"))

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	count, err := dlqHandler.GetDeadLetterMessageCount(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Equal(t, int32(0), count)
}

// TestDLQHandler_GetDeadLetterMessageCount_InvalidCount testa valor inválido de contagem
func TestDLQHandler_GetDeadLetterMessageCount_InvalidCount(t *testing.T) {
	// Arrange
	mockSQS := new(mockSQSClient)
	dlqURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq"
	logger := zap.NewNop()

	mockSQS.On("GetQueueAttributes", mock.Anything, mock.Anything).Return(&sqs.GetQueueAttributesOutput{
		Attributes: map[string]string{
			"ApproximateNumberOfMessages": "invalid",
		},
	}, nil)

	dlqHandler := NewDLQHandler(mockSQS, dlqURL, logger)

	// Act
	count, err := dlqHandler.GetDeadLetterMessageCount(context.Background())

	// Assert
	assert.Error(t, err)
	assert.Equal(t, int32(0), count)
}
