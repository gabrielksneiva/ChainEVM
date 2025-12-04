package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDBClient interface para permitir mocking
type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// DynamoDBAdapter implementa DynamoDBClient wrapping o cliente real
type DynamoDBAdapter struct {
	client *dynamodb.Client
}

// NewDynamoDBAdapter cria um novo adapter
func NewDynamoDBAdapter(client *dynamodb.Client) *DynamoDBAdapter {
	return &DynamoDBAdapter{client: client}
}

// PutItem delega ao cliente real
func (a *DynamoDBAdapter) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	return a.client.PutItem(ctx, params, optFns...)
}

// GetItem delega ao cliente real
func (a *DynamoDBAdapter) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return a.client.GetItem(ctx, params, optFns...)
}

// UpdateItem delega ao cliente real
func (a *DynamoDBAdapter) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return a.client.UpdateItem(ctx, params, optFns...)
}

// Query delega ao cliente real
func (a *DynamoDBAdapter) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	return a.client.Query(ctx, params, optFns...)
}
