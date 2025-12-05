package events

import "time"

// TransactionCreatedEvent evento disparado quando uma transação é criada
type TransactionCreatedEvent struct {
	*BaseDomainEvent
	OperationID    string
	ChainType      string
	OperationType  string
	FromAddress    string
	ToAddress      string
	IdempotencyKey string
}

// NewTransactionCreatedEvent cria um novo evento de transação criada
func NewTransactionCreatedEvent(
	operationID string,
	chainType string,
	operationType string,
	fromAddress string,
	toAddress string,
	idempotencyKey string,
) *TransactionCreatedEvent {
	return &TransactionCreatedEvent{
		BaseDomainEvent: NewBaseDomainEvent("transaction.created", operationID),
		OperationID:     operationID,
		ChainType:       chainType,
		OperationType:   operationType,
		FromAddress:     fromAddress,
		ToAddress:       toAddress,
		IdempotencyKey:  idempotencyKey,
	}
}

// TransactionProcessingEvent evento disparado quando uma transação entra em processamento
type TransactionProcessingEvent struct {
	*BaseDomainEvent
	OperationID string
	ChainType   string
}

// NewTransactionProcessingEvent cria um novo evento de processamento
func NewTransactionProcessingEvent(operationID, chainType string) *TransactionProcessingEvent {
	return &TransactionProcessingEvent{
		BaseDomainEvent: NewBaseDomainEvent("transaction.processing", operationID),
		OperationID:     operationID,
		ChainType:       chainType,
	}
}

// TransactionSucceededEvent evento disparado quando uma transação é bem-sucedida
type TransactionSucceededEvent struct {
	*BaseDomainEvent
	OperationID     string
	ChainType       string
	TransactionHash string
	BlockNumber     int64
	GasUsed         int64
	ExecutedAt      time.Time
}

// NewTransactionSucceededEvent cria um novo evento de sucesso
func NewTransactionSucceededEvent(
	operationID string,
	chainType string,
	txHash string,
	blockNumber int64,
	gasUsed int64,
) *TransactionSucceededEvent {
	return &TransactionSucceededEvent{
		BaseDomainEvent: NewBaseDomainEvent("transaction.succeeded", operationID),
		OperationID:     operationID,
		ChainType:       chainType,
		TransactionHash: txHash,
		BlockNumber:     blockNumber,
		GasUsed:         gasUsed,
		ExecutedAt:      time.Now(),
	}
}

// TransactionFailedEvent evento disparado quando uma transação falha
type TransactionFailedEvent struct {
	*BaseDomainEvent
	OperationID  string
	ChainType    string
	ErrorMessage string
	FailedAt     time.Time
}

// NewTransactionFailedEvent cria um novo evento de falha
func NewTransactionFailedEvent(operationID, chainType, errorMessage string) *TransactionFailedEvent {
	return &TransactionFailedEvent{
		BaseDomainEvent: NewBaseDomainEvent("transaction.failed", operationID),
		OperationID:     operationID,
		ChainType:       chainType,
		ErrorMessage:    errorMessage,
		FailedAt:        time.Now(),
	}
}

// TransactionConfirmedEvent evento disparado quando uma transação é confirmada
type TransactionConfirmedEvent struct {
	*BaseDomainEvent
	OperationID   string
	ChainType     string
	Confirmations int
}

// NewTransactionConfirmedEvent cria um novo evento de confirmação
func NewTransactionConfirmedEvent(operationID, chainType string, confirmations int) *TransactionConfirmedEvent {
	return &TransactionConfirmedEvent{
		BaseDomainEvent: NewBaseDomainEvent("transaction.confirmed", operationID),
		OperationID:     operationID,
		ChainType:       chainType,
		Confirmations:   confirmations,
	}
}
