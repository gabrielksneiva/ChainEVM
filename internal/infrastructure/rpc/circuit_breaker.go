package rpc

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CircuitBreakerState representa o estado do circuit breaker
type CircuitBreakerState string

const (
	// StateClosed estado normal - requisições passam
	StateClosed CircuitBreakerState = "CLOSED"
	// StateOpen estado de falha - requisições são rejeitadas
	StateOpen CircuitBreakerState = "OPEN"
	// StateHalfOpen estado de teste - permite requisição de teste
	StateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreaker implementa o padrão Circuit Breaker para RPC
type CircuitBreaker struct {
	mu                 sync.RWMutex
	state              CircuitBreakerState
	failures           int
	successes          int
	lastFailureTime    time.Time
	failureThreshold   int
	successThreshold   int
	timeout            time.Duration
	halfOpenMaxRetries int
	logger             *zap.Logger
}

// NewCircuitBreaker cria um novo circuit breaker
func NewCircuitBreaker(
	failureThreshold int,
	successThreshold int,
	timeout time.Duration,
	logger *zap.Logger,
) *CircuitBreaker {
	return &CircuitBreaker{
		state:              StateClosed,
		failures:           0,
		successes:          0,
		failureThreshold:   failureThreshold,
		successThreshold:   successThreshold,
		timeout:            timeout,
		halfOpenMaxRetries: 1,
		logger:             logger,
	}
}

// State retorna o estado atual do circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	state := cb.state
	lastFailureTime := cb.lastFailureTime
	timeout := cb.timeout
	cb.mu.RUnlock()

	// Se está OPEN e passou o timeout, transicionar para HALF_OPEN
	if state == StateOpen && time.Since(lastFailureTime) > timeout {
		cb.mu.Lock()
		defer cb.mu.Unlock()

		cb.state = StateHalfOpen
		cb.successes = 0
		cb.logger.Info("circuit breaker transitioned to HALF_OPEN")
		return StateHalfOpen
	}

	return state
}

// Call executa uma função respeitando o estado do circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	state := cb.State()

	if state == StateOpen {
		cb.logger.Warn("circuit breaker is OPEN, rejecting call")
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()

	if err != nil {
		cb.RecordFailure()
		return err
	}

	cb.RecordSuccess()
	return nil
}

// RecordSuccess registra uma execução bem-sucedida
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successes++
		cb.logger.Info("circuit breaker half-open success", zap.Int("successes", cb.successes))

		if cb.successes >= cb.successThreshold {
			cb.state = StateClosed
			cb.failures = 0
			cb.successes = 0
			cb.logger.Info("circuit breaker CLOSED after successful recovery")
		}
	} else if cb.state == StateClosed {
		cb.failures = 0
	}
}

// RecordFailure registra uma falha
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailureTime = time.Now()

	cb.logger.Warn("circuit breaker failure recorded",
		zap.Int("failures", cb.failures),
		zap.Int("threshold", cb.failureThreshold))

	if cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
		cb.logger.Error("circuit breaker OPEN after too many failures")
	}
}

// Reset reseta o circuit breaker para o estado fechado
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.failures = 0
	cb.successes = 0
	cb.logger.Info("circuit breaker reset to CLOSED")
}
