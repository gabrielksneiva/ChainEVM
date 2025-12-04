package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Run("load config with defaults", func(t *testing.T) {
		// Save current environment
		currentEnv := make(map[string]string)
		for _, e := range []string{"ENVIRONMENT", "AWS_REGION", "SQS_QUEUE_URL", "DYNAMODB_TABLE_NAME",
			"REQUEST_TIMEOUT_SECONDS", "RPC_TIMEOUT_SECONDS", "REQUIRED_CONFIRMATIONS"} {
			currentEnv[e] = os.Getenv(e)
			os.Unsetenv(e)
		}
		defer func() {
			for k, v := range currentEnv {
				if v != "" {
					os.Setenv(k, v)
				}
			}
		}()

		cfg := LoadConfig()

		require.NotNil(t, cfg)
		assert.Equal(t, "development", cfg.Environment)
		assert.Equal(t, "us-east-1", cfg.AWSRegion)
		assert.Equal(t, "evm-transactions", cfg.DynamoDBTableName)
		assert.Equal(t, 30*time.Second, cfg.RequestTimeout)
		assert.Equal(t, 10*time.Second, cfg.RPCTimeout)
		assert.Equal(t, 12, cfg.RequiredConfirmations)
	})

	t.Run("load config with environment variables", func(t *testing.T) {
		// Save and set environment variables
		currentEnv := make(map[string]string)
		envVars := map[string]string{
			"ENVIRONMENT":             "production",
			"AWS_REGION":              "us-west-2",
			"SQS_QUEUE_URL":           "https://sqs.us-west-2.amazonaws.com/123456789/test-queue",
			"DYNAMODB_TABLE_NAME":     "test-transactions",
			"REQUEST_TIMEOUT_SECONDS": "60",
			"RPC_TIMEOUT_SECONDS":     "20",
			"REQUIRED_CONFIRMATIONS":  "6",
			"RPC_URL_ETHEREUM":        "https://eth.example.com",
		}

		for k := range envVars {
			currentEnv[k] = os.Getenv(k)
		}
		for k, v := range envVars {
			os.Setenv(k, v)
		}

		defer func() {
			for k, v := range currentEnv {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}
		}()

		cfg := LoadConfig()

		require.NotNil(t, cfg)
		assert.Equal(t, "production", cfg.Environment)
		assert.Equal(t, "us-west-2", cfg.AWSRegion)
		assert.Equal(t, "https://sqs.us-west-2.amazonaws.com/123456789/test-queue", cfg.SQSQueueURL)
		assert.Equal(t, "test-transactions", cfg.DynamoDBTableName)
		assert.Equal(t, 60*time.Second, cfg.RequestTimeout)
		assert.Equal(t, 20*time.Second, cfg.RPCTimeout)
		assert.Equal(t, 6, cfg.RequiredConfirmations)
		assert.Equal(t, "https://eth.example.com", cfg.EVMRPCURLs["ETHEREUM"])
	})

	t.Run("load config with invalid timeout values", func(t *testing.T) {
		currentEnv := make(map[string]string)
		envVars := []string{"REQUEST_TIMEOUT_SECONDS", "RPC_TIMEOUT_SECONDS", "REQUIRED_CONFIRMATIONS"}

		for _, e := range envVars {
			currentEnv[e] = os.Getenv(e)
			os.Setenv(e, "invalid")
		}

		defer func() {
			for k, v := range currentEnv {
				if v != "" {
					os.Setenv(k, v)
				} else {
					os.Unsetenv(k)
				}
			}
		}()

		cfg := LoadConfig()

		// Should use zero values for invalid inputs
		assert.Equal(t, time.Duration(0), cfg.RequestTimeout)
		assert.Equal(t, time.Duration(0), cfg.RPCTimeout)
		assert.Equal(t, 0, cfg.RequiredConfirmations)
	})

	t.Run("verify all RPC URLs are mapped", func(t *testing.T) {
		cfg := LoadConfig()

		require.NotNil(t, cfg.EVMRPCURLs)
		assert.Contains(t, cfg.EVMRPCURLs, "ETHEREUM")
		assert.Contains(t, cfg.EVMRPCURLs, "POLYGON")
		assert.Contains(t, cfg.EVMRPCURLs, "BSC")
		assert.Contains(t, cfg.EVMRPCURLs, "ARBITRUM")
		assert.Contains(t, cfg.EVMRPCURLs, "OPTIMISM")
		assert.Contains(t, cfg.EVMRPCURLs, "AVALANCHE")
	})
}
