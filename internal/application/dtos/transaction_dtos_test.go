package dtos

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteTransactionRequest(t *testing.T) {
	t.Parallel()

	t.Run("create valid request", func(t *testing.T) {
		t.Parallel()

		req := &ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440000",
			ChainType:      "ETHEREUM",
			OperationType:  "TRANSFER",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "0x0987654321098765432109876543210987654321",
			Payload:        map[string]interface{}{"amount": "1.5"},
			IdempotencyKey: "key-123",
		}

		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", req.OperationID)
		assert.Equal(t, "ETHEREUM", req.ChainType)
		assert.Equal(t, "TRANSFER", req.OperationType)
		assert.Equal(t, "0x1234567890123456789012345678901234567890", req.FromAddress)
		assert.Equal(t, "0x0987654321098765432109876543210987654321", req.ToAddress)
		assert.NotNil(t, req.Payload)
		assert.Equal(t, "key-123", req.IdempotencyKey)
	})

	t.Run("create request with empty payload", func(t *testing.T) {
		t.Parallel()

		req := &ExecuteTransactionRequest{
			OperationID:    "550e8400-e29b-41d4-a716-446655440001",
			ChainType:      "POLYGON",
			OperationType:  "GET_BALANCE",
			FromAddress:    "0x1234567890123456789012345678901234567890",
			ToAddress:      "",
			Payload:        map[string]interface{}{},
			IdempotencyKey: "key-124",
		}

		assert.Equal(t, "GET_BALANCE", req.OperationType)
		assert.Empty(t, req.ToAddress)
		assert.Empty(t, req.Payload)
	})
}

func TestExecuteTransactionResponse(t *testing.T) {
	t.Parallel()

	t.Run("create successful response", func(t *testing.T) {
		t.Parallel()

		resp := &ExecuteTransactionResponse{
			OperationID:     "550e8400-e29b-41d4-a716-446655440000",
			Status:          "success",
			TransactionHash: "0xabc123...",
			ChainType:       "ETHEREUM",
			CreatedAt:       "2024-01-01T00:00:00Z",
		}

		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", resp.OperationID)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, "0xabc123...", resp.TransactionHash)
		assert.Equal(t, "ETHEREUM", resp.ChainType)
	})

	t.Run("create error response", func(t *testing.T) {
		t.Parallel()

		resp := &ExecuteTransactionResponse{
			OperationID:  "550e8400-e29b-41d4-a716-446655440001",
			Status:       "failed",
			ErrorMessage: "Invalid address",
			ChainType:    "ETHEREUM",
			CreatedAt:    "2024-01-01T00:00:00Z",
		}

		assert.Equal(t, "failed", resp.Status)
		assert.NotEmpty(t, resp.ErrorMessage)
	})
}

func TestQueryResultResponse(t *testing.T) {
	t.Parallel()

	t.Run("create query result response", func(t *testing.T) {
		t.Parallel()

		result := map[string]interface{}{
			"balance": "1000000000000000000",
			"nonce":   "42",
		}

		resp := &QueryResultResponse{
			OperationID: "550e8400-e29b-41d4-a716-446655440000",
			Status:      "success",
			Result:      result,
		}

		require.NotNil(t, resp.Result)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", resp.OperationID)
		assert.Equal(t, "success", resp.Status)

		resultMap, ok := resp.Result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "1000000000000000000", resultMap["balance"])
		assert.Equal(t, "42", resultMap["nonce"])
	})

	t.Run("create query result with nil result", func(t *testing.T) {
		t.Parallel()

		resp := &QueryResultResponse{
			OperationID: "550e8400-e29b-41d4-a716-446655440001",
			Status:      "pending",
			Result:      nil,
		}

		assert.Nil(t, resp.Result)
		assert.Equal(t, "pending", resp.Status)
	})
}
