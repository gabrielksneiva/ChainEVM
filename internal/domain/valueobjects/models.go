package valueobjects

import (
	"fmt"
	"regexp"
)

// ChainType representa o tipo de blockchain EVM suportada
type ChainType string

const (
	ChainTypeEthereum  ChainType = "ETHEREUM"
	ChainTypePolygon   ChainType = "POLYGON"
	ChainTypeBSC       ChainType = "BSC"
	ChainTypeArbitrum  ChainType = "ARBITRUM"
	ChainTypeOptimism  ChainType = "OPTIMISM"
	ChainTypeAvalanche ChainType = "AVALANCHE"
)

// NewChainType cria e valida um novo ChainType
func NewChainType(value string) (ChainType, error) {
	ct := ChainType(value)
	if !ct.IsValid() {
		return "", fmt.Errorf("invalid chain type: %s", value)
	}
	return ct, nil
}

// IsValid verifica se o chain type é válido
func (c ChainType) IsValid() bool {
	switch c {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBSC,
		ChainTypeArbitrum, ChainTypeOptimism, ChainTypeAvalanche:
		return true
	default:
		return false
	}
}

// String retorna a representação em string
func (c ChainType) String() string {
	return string(c)
}

// Equals verifica igualdade
func (c ChainType) Equals(other ChainType) bool {
	return c == other
}

// EVMAddress representa um endereço Ethereum válido
type EVMAddress string

// NewEVMAddress cria e valida um novo endereço EVM
func NewEVMAddress(value string) (EVMAddress, error) {
	if !isValidEVMAddress(value) {
		return "", fmt.Errorf("invalid EVM address: %s", value)
	}
	return EVMAddress(value), nil
}

// isValidEVMAddress valida um endereço Ethereum (40 caracteres hex com ou sem prefixo 0x)
func isValidEVMAddress(address string) bool {
	re := regexp.MustCompile(`^(0x)?[0-9a-fA-F]{40}$`)
	return re.MatchString(address)
}

// String retorna a representação em string
func (a EVMAddress) String() string {
	return string(a)
}

// Equals verifica igualdade
func (a EVMAddress) Equals(other EVMAddress) bool {
	return a == other
}

// OperationType representa o tipo de operação EVM
type OperationType string

const (
	OperationTypeTransfer    OperationType = "TRANSFER"
	OperationTypeDeploy      OperationType = "DEPLOY"
	OperationTypeCall        OperationType = "CALL"
	OperationTypeApprove     OperationType = "APPROVE"
	OperationTypeSwap        OperationType = "SWAP"
	OperationTypeStake       OperationType = "STAKE"
	OperationTypeUnstake     OperationType = "UNSTAKE"
	OperationTypeWithdraw    OperationType = "WITHDRAW"
	OperationTypeMint        OperationType = "MINT"
	OperationTypeBurn        OperationType = "BURN"
	OperationTypeQuery       OperationType = "QUERY"
	OperationTypeGetBalance  OperationType = "GET_BALANCE"
	OperationTypeGetNonce    OperationType = "GET_NONCE"
	OperationTypeEstimateGas OperationType = "ESTIMATE_GAS"
)

// NewOperationType cria e valida um novo OperationType
func NewOperationType(value string) (OperationType, error) {
	ot := OperationType(value)
	if !ot.IsValid() {
		return "", fmt.Errorf("invalid operation type: %s", value)
	}
	return ot, nil
}

// IsValid verifica se o operation type é válido
func (o OperationType) IsValid() bool {
	switch o {
	case OperationTypeTransfer, OperationTypeDeploy, OperationTypeCall,
		OperationTypeApprove, OperationTypeSwap, OperationTypeStake,
		OperationTypeUnstake, OperationTypeWithdraw, OperationTypeMint,
		OperationTypeBurn, OperationTypeQuery, OperationTypeGetBalance,
		OperationTypeGetNonce, OperationTypeEstimateGas:
		return true
	default:
		return false
	}
}

// String retorna a representação em string
func (o OperationType) String() string {
	return string(o)
}

// Equals verifica igualdade
func (o OperationType) Equals(other OperationType) bool {
	return o == other
}

// IsWriteOperation verifica se a operação modifica o estado da blockchain
func (o OperationType) IsWriteOperation() bool {
	switch o {
	case OperationTypeTransfer, OperationTypeDeploy, OperationTypeCall,
		OperationTypeApprove, OperationTypeSwap, OperationTypeStake,
		OperationTypeUnstake, OperationTypeWithdraw, OperationTypeMint,
		OperationTypeBurn:
		return true
	default:
		return false
	}
}

// OperationID representa o ID único da operação
type OperationID string

// NewOperationID cria e valida um novo OperationID
func NewOperationID(value string) (OperationID, error) {
	if !isValidOperationID(value) {
		return "", fmt.Errorf("invalid operation ID: %s", value)
	}
	return OperationID(value), nil
}

// isValidOperationID valida um OperationID (UUID format)
func isValidOperationID(id string) bool {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return re.MatchString(id)
}

// String retorna a representação em string
func (o OperationID) String() string {
	return string(o)
}

// Equals verifica igualdade
func (o OperationID) Equals(other OperationID) bool {
	return o == other
}

// TransactionHash representa o hash de uma transação EVM
type TransactionHash string

// NewTransactionHash cria e valida um novo TransactionHash
func NewTransactionHash(value string) (TransactionHash, error) {
	if !isValidTransactionHash(value) {
		return "", fmt.Errorf("invalid transaction hash: %s", value)
	}
	return TransactionHash(value), nil
}

// isValidTransactionHash valida um hash de transação (32 bytes hex, com ou sem 0x)
func isValidTransactionHash(hash string) bool {
	re := regexp.MustCompile(`^(0x)?[0-9a-fA-F]{64}$`)
	return re.MatchString(hash)
}

// String retorna a representação em string
func (t TransactionHash) String() string {
	return string(t)
}

// Equals verifica igualdade
func (t TransactionHash) Equals(other TransactionHash) bool {
	return t == other
}
