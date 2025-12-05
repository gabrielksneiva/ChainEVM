package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseDomainEvent(t *testing.T) {
	event := NewBaseDomainEvent("test.event", "agg-123")

	require.NotNil(t, event)
	assert.Equal(t, "test.event", event.EventType())
	assert.Equal(t, "agg-123", event.AggregateID())
	assert.NotNil(t, event.OccurredAt())
}

func TestNewTransactionCreatedEvent(t *testing.T) {
	event := NewTransactionCreatedEvent(
		"op-123",
		"ETHEREUM",
		"TRANSFER",
		"0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222",
		"idempotency-123",
	)

	require.NotNil(t, event)
	assert.Equal(t, "transaction.created", event.EventType())
	assert.Equal(t, "op-123", event.AggregateID())
	assert.Equal(t, "ETHEREUM", event.ChainType)
	assert.Equal(t, "TRANSFER", event.OperationType)
	assert.Equal(t, "0x1111111111111111111111111111111111111111", event.FromAddress)
	assert.Equal(t, "0x2222222222222222222222222222222222222222", event.ToAddress)
	assert.Equal(t, "idempotency-123", event.IdempotencyKey)
}

func TestNewTransactionProcessingEvent(t *testing.T) {
	event := NewTransactionProcessingEvent("op-234", "POLYGON")

	require.NotNil(t, event)
	assert.Equal(t, "transaction.processing", event.EventType())
	assert.Equal(t, "op-234", event.AggregateID())
	assert.Equal(t, "POLYGON", event.ChainType)
}

func TestNewTransactionSucceededEvent(t *testing.T) {
	event := NewTransactionSucceededEvent(
		"op-456",
		"POLYGON",
		"0x1234567890123456789012345678901234567890123456789012345678901234",
		100,
		21000,
	)

	require.NotNil(t, event)
	assert.Equal(t, "transaction.succeeded", event.EventType())
	assert.Equal(t, "op-456", event.AggregateID())
	assert.Equal(t, "POLYGON", event.ChainType)
	assert.Equal(t, "0x1234567890123456789012345678901234567890123456789012345678901234", event.TransactionHash)
	assert.Equal(t, int64(100), event.BlockNumber)
	assert.Equal(t, int64(21000), event.GasUsed)
}

func TestNewTransactionFailedEvent(t *testing.T) {
	event := NewTransactionFailedEvent(
		"op-789",
		"BSC",
		"insufficient funds",
	)

	require.NotNil(t, event)
	assert.Equal(t, "transaction.failed", event.EventType())
	assert.Equal(t, "op-789", event.AggregateID())
	assert.Equal(t, "BSC", event.ChainType)
	assert.Equal(t, "insufficient funds", event.ErrorMessage)
}

func TestNewTransactionConfirmedEvent(t *testing.T) {
	event := NewTransactionConfirmedEvent(
		"op-999",
		"ETHEREUM",
		15,
	)

	require.NotNil(t, event)
	assert.Equal(t, "transaction.confirmed", event.EventType())
	assert.Equal(t, "op-999", event.AggregateID())
	assert.Equal(t, "ETHEREUM", event.ChainType)
	assert.Equal(t, 15, event.Confirmations)
}

func TestBaseDomainEventAttributes(t *testing.T) {
	event := NewBaseDomainEvent("custom.event", "custom-agg")

	assert.NotEmpty(t, event.OccurredAt())
	assert.Equal(t, "custom.event", event.EventType())
	assert.Equal(t, "custom-agg", event.AggregateID())
}

func TestMultipleEventsWithDifferentChains(t *testing.T) {
	chains := []string{"ETHEREUM", "POLYGON", "BSC", "ARBITRUM", "OPTIMISM", "AVALANCHE"}

	for _, chain := range chains {
		t.Run("chain_"+chain, func(t *testing.T) {
			event := NewTransactionProcessingEvent("op-123", chain)
			assert.Equal(t, chain, event.ChainType)
			assert.NotNil(t, event.OccurredAt())
		})
	}
}

func TestEventImmutability(t *testing.T) {
	event := NewTransactionCreatedEvent(
		"op-123",
		"ETHEREUM",
		"TRANSFER",
		"0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222",
		"key-123",
	)

	assert.Equal(t, "op-123", event.AggregateID())
	assert.Equal(t, "ETHEREUM", event.ChainType)
}

func TestTransactionEventsTimestamps(t *testing.T) {
	event1 := NewTransactionFailedEvent("op-1", "ETH", "error")
	event2 := NewTransactionSucceededEvent("op-2", "POLYGON", "0xhash", 100, 21000)

	assert.NotNil(t, event1.FailedAt)
	assert.NotNil(t, event2.ExecutedAt)
}
