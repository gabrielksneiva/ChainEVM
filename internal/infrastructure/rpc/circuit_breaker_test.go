package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCircuitBreaker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cb := NewCircuitBreaker(5, 2, 30*time.Second, logger)

	require.NotNil(t, cb)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cb := NewCircuitBreaker(2, 1, 100*time.Millisecond, logger)

	// Estado inicial: CLOSED
	assert.Equal(t, StateClosed, cb.State())

	// Simular 2 falhas -> deve ir para OPEN
	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State())

	// Aguardar timeout
	time.Sleep(150 * time.Millisecond)

	// Deve transicionar para HALF_OPEN
	assert.Equal(t, StateHalfOpen, cb.State())

	// Sucesso em HALF_OPEN -> voltar para CLOSED
	cb.RecordSuccess()
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cb := NewCircuitBreaker(2, 1, 100*time.Millisecond, logger)

	callCount := 0
	fn := func() error {
		callCount++
		if callCount <= 2 {
			return fmt.Errorf("error")
		}
		return nil
	}

	// Primeira chamada falha
	err := cb.Call(fn)
	require.Error(t, err)
	assert.Equal(t, StateClosed, cb.State())

	// Segunda chamada falha -> CB vai para OPEN
	err = cb.Call(fn)
	require.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())

	// Terceira chamada é rejeitada pelo CB
	err = cb.Call(fn)
	require.Error(t, err)
	assert.Equal(t, StateOpen, cb.State())

	// Aguardar timeout
	time.Sleep(150 * time.Millisecond)

	// Agora deve estar em HALF_OPEN e permitir a chamada
	assert.Equal(t, StateHalfOpen, cb.State())
	err = cb.Call(fn)
	require.NoError(t, err)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerReset(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cb := NewCircuitBreaker(2, 1, 30*time.Second, logger)

	cb.RecordFailure()
	cb.RecordFailure()
	assert.Equal(t, StateOpen, cb.State())

	cb.Reset()
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreakerRecordSuccessWhenClosed(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cb := NewCircuitBreaker(2, 1, 30*time.Second, logger)

	// Record a failure while closed
	cb.RecordFailure()
	assert.Equal(t, StateClosed, cb.State())

	// Then record success should reset failures
	cb.RecordSuccess()
	assert.Equal(t, StateClosed, cb.State())

	// Record another failure - should still be at 1
	cb.RecordFailure()
	assert.Equal(t, StateClosed, cb.State())
}
