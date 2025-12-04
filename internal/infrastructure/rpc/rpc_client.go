package rpc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// RPCClient interface para operações de RPC
type RPCClient interface {
	GetBalance(ctx context.Context, address string) (*big.Int, error)
	GetNonce(ctx context.Context, address string) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	EstimateGas(ctx context.Context, msg interface{}) (uint64, error)
	GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error)
	GetChainID(ctx context.Context) (*big.Int, error)
	GetGasPrice(ctx context.Context) (*big.Int, error)
	Close() error
}

// EVMRPCClient implementação do RPCClient para Ethereum
type EVMRPCClient struct {
	client  EthClient
	rpcURL  string
	timeout time.Duration
	logger  *zap.Logger
}

// NewEVMRPCClient cria uma nova instância do cliente EVM
func NewEVMRPCClient(rpcURL string, timeout time.Duration, logger *zap.Logger) (RPCClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		logger.Error("failed to connect to EVM RPC", zap.String("rpc_url", rpcURL), zap.Error(err))
		return nil, fmt.Errorf("failed to connect to EVM RPC: %w", err)
	}

	logger.Info("connected to EVM RPC", zap.String("rpc_url", rpcURL))

	return &EVMRPCClient{
		client:  NewEthClientAdapter(client),
		rpcURL:  rpcURL,
		timeout: timeout,
		logger:  logger,
	}, nil
}

// GetBalance retorna o saldo de um endereço
func (c *EVMRPCClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, nil)
	if err != nil {
		c.logger.Error("failed to get balance", zap.String("address", address), zap.Error(err))
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetNonce retorna o nonce de um endereço
func (c *EVMRPCClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	addr := common.HexToAddress(address)
	nonce, err := c.client.PendingNonceAt(ctx, addr)
	if err != nil {
		c.logger.Error("failed to get nonce", zap.String("address", address), zap.Error(err))
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	return nonce, nil
}

// SendTransaction envia uma transação
func (c *EVMRPCClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	err := c.client.SendTransaction(ctx, tx)
	if err != nil {
		c.logger.Error("failed to send transaction", zap.Error(err))
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	c.logger.Info("transaction sent", zap.String("tx_hash", tx.Hash().Hex()))
	return nil
}

// EstimateGas estima o gas necessário para uma transação
func (c *EVMRPCClient) EstimateGas(ctx context.Context, msg interface{}) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// msg deve ser um types.CallMsg convertido
	callMsg, ok := msg.(interface {
		GetFrom() common.Address
		GetTo() *common.Address
		GetData() []byte
		GetValue() *big.Int
	})
	if !ok {
		return 0, fmt.Errorf("invalid message type for gas estimation")
	}

	// Simplificação para demo - em produção seria mais complexo
	gas := uint64(21000)
	if callMsg.GetData() != nil {
		gas += uint64(len(callMsg.GetData()) * 4)
	}

	return gas, nil
}

// GetTransactionReceipt retorna o recebimento de uma transação
func (c *EVMRPCClient) GetTransactionReceipt(ctx context.Context, txHash string) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	hash := common.HexToHash(txHash)
	receipt, err := c.client.TransactionReceipt(ctx, hash)
	if err != nil {
		c.logger.Error("failed to get transaction receipt", zap.String("tx_hash", txHash), zap.Error(err))
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	return receipt, nil
}

// GetChainID retorna o ID da chain
func (c *EVMRPCClient) GetChainID(ctx context.Context) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		c.logger.Error("failed to get chain ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID, nil
}

// GetGasPrice retorna o preço do gas
func (c *EVMRPCClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		c.logger.Error("failed to get gas price", zap.Error(err))
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	return gasPrice, nil
}

// Close fecha a conexão com o RPC
func (c *EVMRPCClient) Close() error {
	c.client.Close()
	c.logger.Info("RPC client closed")
	return nil
}
