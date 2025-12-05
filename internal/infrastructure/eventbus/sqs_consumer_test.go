package eventbus

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSQSClient mock do cliente SQS
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ChangeMessageVisibilityOutput), args.Error(1)
}

func stringPtr(s string) *string {
	return &s
}

func TestNewSQSConsumer(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()

	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	assert.NotNil(t, consumer)
}

func TestSQSConsumer_ReceiveMessages_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	msg := Message{
		OperationID:   "op123",
		ChainType:     "ETHEREUM",
		OperationType: "TRANSFER",
		FromAddress:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:     "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
	}
	msgBody, _ := json.Marshal(msg)

	mockOutput := &sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				MessageId:     stringPtr("msg1"),
				ReceiptHandle: stringPtr("receipt1"),
				Body:          stringPtr(string(msgBody)),
			},
		},
	}

	mockClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(mockOutput, nil)

	messages, err := consumer.ReceiveMessages(context.Background(), 10)

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	mockClient.AssertExpectations(t)
}

func TestSQSConsumer_ReceiveMessages_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	mockClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(nil, errors.New("sqs error"))

	messages, err := consumer.ReceiveMessages(context.Background(), 10)

	assert.Error(t, err)
	assert.Nil(t, messages)
	mockClient.AssertExpectations(t)
}

func TestSQSConsumer_ReceiveMessages_Empty(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	mockOutput := &sqs.ReceiveMessageOutput{
		Messages: []types.Message{},
	}

	mockClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(mockOutput, nil)

	messages, err := consumer.ReceiveMessages(context.Background(), 10)

	assert.NoError(t, err)
	assert.Len(t, messages, 0)
	mockClient.AssertExpectations(t)
}

func TestSQSConsumer_ParseMessage_ValidJSON(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	msg := Message{
		OperationID:   "op123",
		ChainType:     "ETHEREUM",
		OperationType: "TRANSFER",
		FromAddress:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:     "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
	}
	msgBody, _ := json.Marshal(msg)

	sqsMsg := types.Message{
		Body: stringPtr(string(msgBody)),
	}

	parsedMsg, err := consumer.ParseMessage(sqsMsg)

	assert.NoError(t, err)
	assert.Equal(t, "op123", parsedMsg.OperationID)
	assert.Equal(t, "ETHEREUM", parsedMsg.ChainType)
}

func TestSQSConsumer_ParseMessage_InvalidJSON(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	sqsMsg := types.Message{
		Body: stringPtr("invalid json"),
	}

	parsedMsg, err := consumer.ParseMessage(sqsMsg)

	assert.Error(t, err)
	assert.Nil(t, parsedMsg)
}

func TestSQSConsumer_DeleteMessage_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	mockClient.On("DeleteMessage", mock.Anything, mock.AnythingOfType("*sqs.DeleteMessageInput")).
		Return(&sqs.DeleteMessageOutput{}, nil)

	err := consumer.DeleteMessage(context.Background(), stringPtr("receipt1"))

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSQSConsumer_DeleteMessage_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	mockClient.On("DeleteMessage", mock.Anything, mock.AnythingOfType("*sqs.DeleteMessageInput")).
		Return(nil, errors.New("sqs error"))

	err := consumer.DeleteMessage(context.Background(), stringPtr("receipt1"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete message")
	mockClient.AssertExpectations(t)
}

func TestSQSConsumer_ChangeMessageVisibility_Success(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	mockClient.On("ChangeMessageVisibility", mock.Anything, mock.AnythingOfType("*sqs.ChangeMessageVisibilityInput")).
		Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

	err := consumer.ChangeMessageVisibility(context.Background(), stringPtr("receipt1"), 30)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestSQSConsumer_ChangeMessageVisibility_Error(t *testing.T) {
	t.Parallel()

	mockClient := new(MockSQSClient)
	logger, _ := zap.NewDevelopment()
	consumer := NewSQSConsumer(mockClient, "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue", logger)

	mockClient.On("ChangeMessageVisibility", mock.Anything, mock.AnythingOfType("*sqs.ChangeMessageVisibilityInput")).
		Return(nil, errors.New("sqs error"))

	err := consumer.ChangeMessageVisibility(context.Background(), stringPtr("receipt1"), 30)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to change message visibility")
	mockClient.AssertExpectations(t)
}

func TestMessage_Marshalling(t *testing.T) {
	t.Parallel()

	msg := Message{
		OperationID:   "op123",
		ChainType:     "ETHEREUM",
		OperationType: "TRANSFER",
		FromAddress:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
		ToAddress:     "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
	}

	data, err := json.Marshal(msg)
	assert.NoError(t, err)

	var parsed Message
	err = json.Unmarshal(data, &parsed)
	assert.NoError(t, err)
	assert.Equal(t, msg.OperationID, parsed.OperationID)
}
