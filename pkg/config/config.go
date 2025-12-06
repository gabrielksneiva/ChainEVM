package config

import (
	"os"
	"strconv"
	"time"
)

// Config configuração da aplicação
type Config struct {
	Environment string
	AWSRegion   string

	// AWS SQS
	SQSQueueURL    string
	SQSQueueDLQURL string

	// AWS DynamoDB
	DynamoDBTableName string

	// EVM RPC URLs
	EVMRPCURLs map[string]string

	// Timeouts
	RequestTimeout time.Duration
	RPCTimeout     time.Duration

	// Blockchain confirmations
	RequiredConfirmations int
}

// LoadConfig carrega configuração a partir de variáveis de ambiente
func LoadConfig() *Config {
	requestTimeout, _ := strconv.Atoi(getEnv("REQUEST_TIMEOUT_SECONDS", "30"))
	rpcTimeout, _ := strconv.Atoi(getEnv("RPC_TIMEOUT_SECONDS", "10"))
	requiredConfirmations, _ := strconv.Atoi(getEnv("REQUIRED_CONFIRMATIONS", "12"))

	evmRPCURLs := map[string]string{
		"ETHEREUM":  getEnv("RPC_URL_ETHEREUM", "https://eth-mainnet.g.alchemy.com/v2/demo"),
		"POLYGON":   getEnv("RPC_URL_POLYGON", "https://polygon-mainnet.g.alchemy.com/v2/demo"),
		"BSC":       getEnv("RPC_URL_BSC", "https://bsc-mainnet.infura.io/v3/demo"),
		"ARBITRUM":  getEnv("RPC_URL_ARBITRUM", "https://arb-mainnet.g.alchemy.com/v2/demo"),
		"OPTIMISM":  getEnv("RPC_URL_OPTIMISM", "https://opt-mainnet.g.alchemy.com/v2/demo"),
		"AVALANCHE": getEnv("RPC_URL_AVALANCHE", "https://avax-mainnet.g.alchemy.com/v2/demo"),
	}

	return &Config{
		Environment:           getEnv("ENVIRONMENT", "development"),
		AWSRegion:             getEnv("AWS_REGION", "us-east-1"),
		SQSQueueURL:           getEnv("SQS_QUEUE_URL", ""),
		SQSQueueDLQURL:        getEnv("SQS_QUEUE_DLQ_URL", ""),
		DynamoDBTableName:     getEnv("DYNAMODB_TABLE_NAME", "evm-transactions"),
		EVMRPCURLs:            evmRPCURLs,
		RequestTimeout:        time.Duration(requestTimeout) * time.Second,
		RPCTimeout:            time.Duration(rpcTimeout) * time.Second,
		RequiredConfirmations: requiredConfirmations,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
