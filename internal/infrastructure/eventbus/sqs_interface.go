package eventbus

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSClient interface para permitir mocking
type SQSClient interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error)
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

// SQSAdapter implementa SQSClient wrapping o cliente real
type SQSAdapter struct {
	client *sqs.Client
}

// NewSQSAdapter cria um novo adapter
func NewSQSAdapter(client *sqs.Client) *SQSAdapter {
	return &SQSAdapter{client: client}
}

// ReceiveMessage delega ao cliente real
func (a *SQSAdapter) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	return a.client.ReceiveMessage(ctx, params, optFns...)
}

// DeleteMessage delega ao cliente real
func (a *SQSAdapter) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return a.client.DeleteMessage(ctx, params, optFns...)
}

// ChangeMessageVisibility delega ao cliente real
func (a *SQSAdapter) ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error) {
	return a.client.ChangeMessageVisibility(ctx, params, optFns...)
}

// SendMessage delega ao cliente real
func (a *SQSAdapter) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return a.client.SendMessage(ctx, params, optFns...)
}

// GetQueueAttributes delega ao cliente real
func (a *SQSAdapter) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	return a.client.GetQueueAttributes(ctx, params, optFns...)
}
