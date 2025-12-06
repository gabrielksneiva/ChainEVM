package rpc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go.uber.org/zap"
)

// SignedTransactionClient interface para operações de assinatura e envio
type SignedTransactionClient interface {
	SignAndSendTransaction(ctx context.Context, tx *types.Transaction, privateKey string) (string, error)
	WaitForConfirmations(ctx context.Context, txHash string, requiredConfirmations int) (*types.Receipt, error)
}

// TransactionSigner implementa assinatura de transações
type TransactionSigner struct {
	client       EthClient
	chainID      *big.Int
	logger       *zap.Logger
	timeout      time.Duration
	pollInterval time.Duration
}

// NewTransactionSigner cria um novo signer
func NewTransactionSigner(client EthClient, chainID *big.Int, logger *zap.Logger, timeout time.Duration) *TransactionSigner {
	return &TransactionSigner{
		client:       client,
		chainID:      chainID,
		logger:       logger,
		timeout:      timeout,
		pollInterval: 3 * time.Second,
	}
}

// SignAndSendTransaction assina e envia uma transação
func (s *TransactionSigner) SignAndSendTransaction(ctx context.Context, tx *types.Transaction, privateKeyHex string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Parse private key
	pk, err := parsePrivateKey(privateKeyHex)
	if err != nil {
		s.logger.Error("failed to parse private key", zap.Error(err))
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(s.chainID), pk)
	if err != nil {
		s.logger.Error("failed to sign transaction", zap.Error(err))
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = s.client.SendTransaction(ctx, signedTx)
	if err != nil {
		s.logger.Error("failed to send signed transaction", zap.Error(err))
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	txHash := signedTx.Hash().Hex()
	s.logger.Info("transaction sent", zap.String("tx_hash", txHash))
	return txHash, nil
}

// WaitForConfirmations aguarda N confirmações
func (s *TransactionSigner) WaitForConfirmations(ctx context.Context, txHashHex string, requiredConfirmations int) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Parse hash
	txHash := common.HexToHash(txHashHex)

	for {
		select {
		case <-ctx.Done():
			s.logger.Warn("timeout waiting for confirmations", zap.String("tx_hash", txHashHex))
			return nil, fmt.Errorf("timeout waiting for confirmations")
		case <-time.After(s.pollInterval):
			receipt, err := s.client.TransactionReceipt(ctx, txHash)
			if err != nil {
				s.logger.Debug("transaction not yet mined", zap.String("tx_hash", txHashHex))
				continue
			}

			currentBlockNum, err := s.client.BlockNumber(ctx)
			if err != nil {
				s.logger.Error("failed to get block number", zap.Error(err))
				continue
			}

			confirmations := int(currentBlockNum) - int(receipt.BlockNumber.Uint64()) + 1
			s.logger.Info("checking confirmations",
				zap.String("tx_hash", txHashHex),
				zap.Int("confirmations", confirmations),
				zap.Int("required", requiredConfirmations))

			if confirmations >= requiredConfirmations {
				s.logger.Info("transaction confirmed",
					zap.String("tx_hash", txHashHex),
					zap.Int("confirmations", confirmations))
				return receipt, nil
			}
		}
	}
}

// parsePrivateKey parses hex private key to ECDSA private key
func parsePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	// Remove '0x' prefix if present
	if len(hexKey) > 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}

	pk, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	return pk, nil
}
