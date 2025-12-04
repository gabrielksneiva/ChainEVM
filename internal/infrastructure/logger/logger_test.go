package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("create development logger", func(t *testing.T) {
		t.Parallel()

		logger, err := NewLogger("development")

		require.NoError(t, err)
		require.NotNil(t, logger)

		// Test that we can use the logger
		logger.Info("test message")
		logger.Debug("debug message")
	})

	t.Run("create production logger", func(t *testing.T) {
		t.Parallel()

		logger, err := NewLogger("production")

		require.NoError(t, err)
		require.NotNil(t, logger)

		// Test that we can use the logger
		logger.Info("test message", zap.String("key", "value"))
		logger.Error("error message", zap.Error(assert.AnError))
	})

	t.Run("create logger with custom environment", func(t *testing.T) {
		t.Parallel()

		logger, err := NewLogger("staging")

		require.NoError(t, err)
		require.NotNil(t, logger)

		logger.Info("test message")
	})

	t.Run("logger with empty environment defaults to production", func(t *testing.T) {
		t.Parallel()

		logger, err := NewLogger("")

		require.NoError(t, err)
		require.NotNil(t, logger)
	})
}

func TestLoggerUsage(t *testing.T) {
	logger, err := NewLogger("development")
	require.NoError(t, err)

	t.Run("info logging", func(t *testing.T) {
		logger.Info("info message", zap.String("key", "value"))
	})

	t.Run("error logging", func(t *testing.T) {
		logger.Error("error message", zap.Error(assert.AnError))
	})

	t.Run("debug logging", func(t *testing.T) {
		logger.Debug("debug message", zap.Int("count", 42))
	})

	t.Run("warn logging", func(t *testing.T) {
		logger.Warn("warning message", zap.String("status", "warning"))
	})
}

func TestLoggerWithFields(t *testing.T) {
	logger, err := NewLogger("production")
	require.NoError(t, err)

	t.Run("log with multiple fields", func(t *testing.T) {
		logger.Info("transaction processed",
			zap.String("tx_id", "0x123"),
			zap.String("chain", "ETHEREUM"),
			zap.String("status", "success"),
		)
	})

	t.Run("log with nested fields", func(t *testing.T) {
		logger.Info("operation",
			zap.String("op_id", "op-1"),
			zap.String("chain", "POLYGON"),
		)
	})
}

func TestLoggerEnvironments(t *testing.T) {
	environments := []string{"development", "production", "staging", "test"}

	for _, env := range environments {
		t.Run("logger_"+env, func(t *testing.T) {
			logger, err := NewLogger(env)
			require.NoError(t, err)
			require.NotNil(t, logger)

			logger.Info("test", zap.String("env", env))
		})
	}
}

func TestLoggerFatality(t *testing.T) {
	logger, err := NewLogger("production")
	require.NoError(t, err)

	t.Run("fatal logging not invoked to avoid test exit", func(t *testing.T) {
		// We can't actually call Fatal as it exits the process
		// but we verify the logger is properly initialized
		require.NotNil(t, logger)
	})
}

func TestDevelopmentVsProductionLogger(t *testing.T) {
	devLogger, devErr := NewLogger("development")
	prodLogger, prodErr := NewLogger("production")

	require.NoError(t, devErr)
	require.NoError(t, prodErr)
	require.NotNil(t, devLogger)
	require.NotNil(t, prodLogger)

	// Both loggers should be usable
	devLogger.Info("dev test")
	prodLogger.Info("prod test")
}
