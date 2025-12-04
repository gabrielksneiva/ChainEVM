package metrics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	m := NewMetrics(logger)

	require.NotNil(t, m)

	stats := m.GetStats(context.Background())
	assert.Equal(t, int64(0), stats["transaction_count"])
	assert.Equal(t, int64(0), stats["transaction_success"])
}

func TestIncrementMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	m := NewMetrics(logger)

	m.IncrementTransactionCount()
	m.IncrementTransactionCount()
	m.IncrementTransactionSuccess()
	m.IncrementRPCCall()
	m.IncrementRPCCall()
	m.IncrementRPCCallError()

	stats := m.GetStats(context.Background())
	assert.Equal(t, int64(2), stats["transaction_count"])
	assert.Equal(t, int64(1), stats["transaction_success"])
	assert.Equal(t, int64(2), stats["rpc_call_count"])
	assert.Equal(t, int64(1), stats["rpc_call_errors"])
}

func TestResetMetrics(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	m := NewMetrics(logger)

	m.IncrementTransactionCount()
	m.IncrementTransactionSuccess()
	m.IncrementRPCCall()

	stats := m.GetStats(context.Background())
	assert.True(t, stats["transaction_count"] > 0)

	m.Reset()

	stats = m.GetStats(context.Background())
	assert.Equal(t, int64(0), stats["transaction_count"])
	assert.Equal(t, int64(0), stats["transaction_success"])
	assert.Equal(t, int64(0), stats["rpc_call_count"])
}
