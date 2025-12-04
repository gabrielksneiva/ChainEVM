package dtos

// ExecuteTransactionRequest representa a requisição para executar uma operação EVM
type ExecuteTransactionRequest struct {
	OperationID    string                 `json:"operation_id" validate:"required,uuid"`
	ChainType      string                 `json:"chain_type" validate:"required,oneof=ETHEREUM POLYGON BSC ARBITRUM OPTIMISM AVALANCHE"`
	OperationType  string                 `json:"operation_type" validate:"required"`
	FromAddress    string                 `json:"from_address" validate:"required"`
	ToAddress      string                 `json:"to_address" validate:"required"`
	Payload        map[string]interface{} `json:"payload" validate:"required"`
	IdempotencyKey string                 `json:"idempotency_key" validate:"required,uuid"`
}

// ExecuteTransactionResponse resposta quando transação é executada
type ExecuteTransactionResponse struct {
	OperationID     string  `json:"operation_id"`
	ChainType       string  `json:"chain_type"`
	TransactionHash string  `json:"transaction_hash,omitempty"`
	Status          string  `json:"status"`
	BlockNumber     *int64  `json:"block_number,omitempty"`
	GasUsed         *int64  `json:"gas_used,omitempty"`
	GasPrice        *string `json:"gas_price,omitempty"`
	ErrorMessage    string  `json:"error_message,omitempty"`
	CreatedAt       string  `json:"created_at"`
	ExecutedAt      *string `json:"executed_at,omitempty"`
}

// QueryResultResponse resposta para operações de leitura
type QueryResultResponse struct {
	OperationID string      `json:"operation_id"`
	ChainType   string      `json:"chain_type"`
	Result      interface{} `json:"result"`
	Status      string      `json:"status"`
	CreatedAt   string      `json:"created_at"`
}
