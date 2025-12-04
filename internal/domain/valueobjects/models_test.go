package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for ChainType
func TestNewChainType(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
		expected  ChainType
	}{
		{"Valid Ethereum", "ETHEREUM", false, ChainTypeEthereum},
		{"Valid Polygon", "POLYGON", false, ChainTypePolygon},
		{"Valid BSC", "BSC", false, ChainTypeBSC},
		{"Invalid chain", "INVALID", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewChainType(tt.value)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestChainTypeIsValid(t *testing.T) {
	assert.True(t, ChainTypeEthereum.IsValid())
	assert.True(t, ChainTypePolygon.IsValid())
	assert.False(t, ChainType("INVALID").IsValid())
}

func TestChainTypeString(t *testing.T) {
	assert.Equal(t, "ETHEREUM", ChainTypeEthereum.String())
}

func TestChainTypeEquals(t *testing.T) {
	ct1, _ := NewChainType("ETHEREUM")
	ct2, _ := NewChainType("ETHEREUM")
	ct3, _ := NewChainType("POLYGON")

	assert.True(t, ct1.Equals(ct2))
	assert.False(t, ct1.Equals(ct3))
}

// Tests for OperationType
func TestNewOperationType(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
		expected  OperationType
	}{
		{"Valid Transfer", "TRANSFER", false, OperationTypeTransfer},
		{"Valid Deploy", "DEPLOY", false, OperationTypeDeploy},
		{"Valid Call", "CALL", false, OperationTypeCall},
		{"Invalid operation", "INVALID_OP", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewOperationType(tt.value)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestOperationTypeIsWriteOperation(t *testing.T) {
	writeOps := []OperationType{
		OperationTypeTransfer,
		OperationTypeDeploy,
		OperationTypeCall,
		OperationTypeApprove,
	}

	readOps := []OperationType{
		OperationTypeGetBalance,
		OperationTypeGetNonce,
		OperationTypeQuery,
	}

	for _, op := range writeOps {
		assert.True(t, op.IsWriteOperation(), "expected %s to be write operation", op)
	}

	for _, op := range readOps {
		assert.False(t, op.IsWriteOperation(), "expected %s to be read operation", op)
	}
}

func TestOperationTypeIsValid(t *testing.T) {
	assert.True(t, OperationTypeTransfer.IsValid())
	assert.True(t, OperationTypeGetBalance.IsValid())
	assert.False(t, OperationType("INVALID").IsValid())
}

func TestOperationTypeString(t *testing.T) {
	assert.Equal(t, "TRANSFER", OperationTypeTransfer.String())
	assert.Equal(t, "GET_BALANCE", OperationTypeGetBalance.String())
}

func TestOperationTypeEquals(t *testing.T) {
	op1, _ := NewOperationType("TRANSFER")
	op2, _ := NewOperationType("TRANSFER")
	op3, _ := NewOperationType("DEPLOY")

	assert.True(t, op1.Equals(op2))
	assert.False(t, op1.Equals(op3))
}

// Tests for OperationID
func TestNewOperationID(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	invalidID := "not-a-uuid"

	id, err := NewOperationID(validUUID)
	require.NoError(t, err)
	assert.Equal(t, OperationID(validUUID), id)

	_, err = NewOperationID(invalidID)
	require.Error(t, err)
}

func TestOperationIDString(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	id, _ := NewOperationID(validUUID)
	assert.Equal(t, validUUID, id.String())
}

func TestOperationIDEquals(t *testing.T) {
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	id1, _ := NewOperationID(uuid)
	id2, _ := NewOperationID(uuid)
	id3, _ := NewOperationID("550e8400-e29b-41d4-a716-446655440001")

	assert.True(t, id1.Equals(id2))
	assert.False(t, id1.Equals(id3))
}

// Tests for EVMAddress
func TestNewEVMAddress(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		shouldErr bool
	}{
		{"Valid address with 0x", "0x1234567890123456789012345678901234567890", false},
		{"Valid address without 0x", "1234567890123456789012345678901234567890", false},
		{"Invalid address", "invalid", true},
		{"Too short", "0x123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEVMAddress(tt.value)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEVMAddressString(t *testing.T) {
	addr := "0x1234567890123456789012345678901234567890"
	evmAddr, _ := NewEVMAddress(addr)
	assert.Equal(t, addr, evmAddr.String())
}

func TestEVMAddressEquals(t *testing.T) {
	addr1, _ := NewEVMAddress("0x1234567890123456789012345678901234567890")
	addr2, _ := NewEVMAddress("0x1234567890123456789012345678901234567890")
	addr3, _ := NewEVMAddress("0x0987654321098765432109876543210987654321")

	assert.True(t, addr1.Equals(addr2))
	assert.False(t, addr1.Equals(addr3))
}

// Tests for TransactionHash
func TestNewTransactionHash(t *testing.T) {
	validHash := "0x1234567890123456789012345678901234567890123456789012345678901234"
	invalidHash := "invalid"

	hash, err := NewTransactionHash(validHash)
	require.NoError(t, err)
	assert.Equal(t, TransactionHash(validHash), hash)

	_, err = NewTransactionHash(invalidHash)
	require.Error(t, err)
}

func TestTransactionHashString(t *testing.T) {
	validHash := "0x1234567890123456789012345678901234567890123456789012345678901234"
	hash, _ := NewTransactionHash(validHash)
	assert.Equal(t, validHash, hash.String())
}

func TestTransactionHashEquals(t *testing.T) {
	hash1Str := "0x1234567890123456789012345678901234567890123456789012345678901234"
	hash2Str := "0x5678901234567890123456789012345678901234567890123456789012345678"

	hash1, _ := NewTransactionHash(hash1Str)
	hash2, _ := NewTransactionHash(hash1Str)
	hash3, _ := NewTransactionHash(hash2Str)

	assert.True(t, hash1.Equals(hash2))
	assert.False(t, hash1.Equals(hash3))
}
