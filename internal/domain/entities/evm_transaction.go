package entities

import (
	"time"

	"github.com/gabrielksneiva/ChainEVM/internal/domain/valueobjects"
)

// EVMTransaction representa uma transação EVM no domínio
type EVMTransaction struct {
	operationID    valueobjects.OperationID
	chainType      valueobjects.ChainType
	operationType  valueobjects.OperationType
	fromAddress    valueobjects.EVMAddress
	toAddress      valueobjects.EVMAddress
	payload        map[string]interface{}
	txHash         valueobjects.TransactionHash
	status         TransactionStatus
	createdAt      time.Time
	executedAt     *time.Time
	blockNumber    *int64
	gasUsed        *int64
	gasPrice       *string
	nonce          *int64
	errorMessage   string
	idempotencyKey string
}

// TransactionStatus status da transação
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusProcessing TransactionStatus = "PROCESSING"
	TransactionStatusSuccess    TransactionStatus = "SUCCESS"
	TransactionStatusFailed     TransactionStatus = "FAILED"
	TransactionStatusConfirmed  TransactionStatus = "CONFIRMED"
)

// NewEVMTransaction cria uma nova transação EVM
func NewEVMTransaction(
	operationID valueobjects.OperationID,
	chainType valueobjects.ChainType,
	operationType valueobjects.OperationType,
	fromAddress valueobjects.EVMAddress,
	toAddress valueobjects.EVMAddress,
	payload map[string]interface{},
	idempotencyKey string,
) *EVMTransaction {
	return &EVMTransaction{
		operationID:    operationID,
		chainType:      chainType,
		operationType:  operationType,
		fromAddress:    fromAddress,
		toAddress:      toAddress,
		payload:        payload,
		status:         TransactionStatusPending,
		createdAt:      time.Now(),
		idempotencyKey: idempotencyKey,
	}
}

// Getters
func (t *EVMTransaction) OperationID() valueobjects.OperationID {
	return t.operationID
}

func (t *EVMTransaction) ChainType() valueobjects.ChainType {
	return t.chainType
}

func (t *EVMTransaction) OperationType() valueobjects.OperationType {
	return t.operationType
}

func (t *EVMTransaction) FromAddress() valueobjects.EVMAddress {
	return t.fromAddress
}

func (t *EVMTransaction) ToAddress() valueobjects.EVMAddress {
	return t.toAddress
}

func (t *EVMTransaction) Payload() map[string]interface{} {
	return t.payload
}

func (t *EVMTransaction) TxHash() valueobjects.TransactionHash {
	return t.txHash
}

func (t *EVMTransaction) Status() TransactionStatus {
	return t.status
}

func (t *EVMTransaction) CreatedAt() time.Time {
	return t.createdAt
}

func (t *EVMTransaction) ExecutedAt() *time.Time {
	return t.executedAt
}

func (t *EVMTransaction) BlockNumber() *int64 {
	return t.blockNumber
}

func (t *EVMTransaction) GasUsed() *int64 {
	return t.gasUsed
}

func (t *EVMTransaction) GasPrice() *string {
	return t.gasPrice
}

func (t *EVMTransaction) Nonce() *int64 {
	return t.nonce
}

func (t *EVMTransaction) ErrorMessage() string {
	return t.errorMessage
}

func (t *EVMTransaction) IdempotencyKey() string {
	return t.idempotencyKey
}

// Setters para atualização de estado
func (t *EVMTransaction) MarkAsProcessing() {
	t.status = TransactionStatusProcessing
}

func (t *EVMTransaction) MarkAsSuccess(txHash valueobjects.TransactionHash, blockNumber int64, gasUsed int64) {
	t.status = TransactionStatusSuccess
	t.txHash = txHash
	t.blockNumber = &blockNumber
	t.gasUsed = &gasUsed
	now := time.Now()
	t.executedAt = &now
}

func (t *EVMTransaction) MarkAsConfirmed() {
	t.status = TransactionStatusConfirmed
}

func (t *EVMTransaction) MarkAsFailed(errorMsg string) {
	t.status = TransactionStatusFailed
	t.errorMessage = errorMsg
	now := time.Now()
	t.executedAt = &now
}

func (t *EVMTransaction) SetTxMetadata(gasPrice string, nonce int64) {
	t.gasPrice = &gasPrice
	t.nonce = &nonce
}
