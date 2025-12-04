package metrics

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
)

// Metrics gerencia métricas da aplicação
type Metrics struct {
	logger             *zap.Logger
	transactionCount   int64
	transactionSuccess int64
	transactionFailed  int64
	rpcCallCount       int64
	rpcCallErrors      int64
}

// NewMetrics cria uma nova instância de Metrics
func NewMetrics(logger *zap.Logger) *Metrics {
	return &Metrics{
		logger: logger,
	}
}

// IncrementTransactionCount incrementa o contador de transações
func (m *Metrics) IncrementTransactionCount() {
	atomic.AddInt64(&m.transactionCount, 1)
}

// IncrementTransactionSuccess incrementa o contador de sucessos
func (m *Metrics) IncrementTransactionSuccess() {
	atomic.AddInt64(&m.transactionSuccess, 1)
}

// IncrementTransactionFailed incrementa o contador de falhas
func (m *Metrics) IncrementTransactionFailed() {
	atomic.AddInt64(&m.transactionFailed, 1)
}

// IncrementRPCCall incrementa o contador de chamadas RPC
func (m *Metrics) IncrementRPCCall() {
	atomic.AddInt64(&m.rpcCallCount, 1)
}

// IncrementRPCCallError incrementa o contador de erros RPC
func (m *Metrics) IncrementRPCCallError() {
	atomic.AddInt64(&m.rpcCallErrors, 1)
}

// GetStats retorna as estatísticas atuais
func (m *Metrics) GetStats(ctx context.Context) map[string]int64 {
	return map[string]int64{
		"transaction_count":   atomic.LoadInt64(&m.transactionCount),
		"transaction_success": atomic.LoadInt64(&m.transactionSuccess),
		"transaction_failed":  atomic.LoadInt64(&m.transactionFailed),
		"rpc_call_count":      atomic.LoadInt64(&m.rpcCallCount),
		"rpc_call_errors":     atomic.LoadInt64(&m.rpcCallErrors),
	}
}

// Reset reseta todas as métricas
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.transactionCount, 0)
	atomic.StoreInt64(&m.transactionSuccess, 0)
	atomic.StoreInt64(&m.transactionFailed, 0)
	atomic.StoreInt64(&m.rpcCallCount, 0)
	atomic.StoreInt64(&m.rpcCallErrors, 0)
	m.logger.Info("metrics reset")
}
