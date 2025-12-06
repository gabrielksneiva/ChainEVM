package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthClient interface para permitir mocking
type EthClient interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	ChainID(ctx context.Context) (*big.Int, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	Close()
}

// EthClientAdapter implementa EthClient wrapping o cliente real
type EthClientAdapter struct {
	client *ethclient.Client
}

// NewEthClientAdapter cria um novo adapter
func NewEthClientAdapter(client *ethclient.Client) *EthClientAdapter {
	return &EthClientAdapter{client: client}
}

// BalanceAt delega ao cliente real
func (a *EthClientAdapter) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return a.client.BalanceAt(ctx, account, blockNumber)
}

// NonceAt delega ao cliente real
func (a *EthClientAdapter) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	return a.client.NonceAt(ctx, account, blockNumber)
}

// PendingNonceAt delega ao cliente real
func (a *EthClientAdapter) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return a.client.PendingNonceAt(ctx, account)
}

// SendTransaction delega ao cliente real
func (a *EthClientAdapter) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return a.client.SendTransaction(ctx, tx)
}

// EstimateGas delega ao cliente real
func (a *EthClientAdapter) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return a.client.EstimateGas(ctx, msg)
}

// TransactionReceipt delega ao cliente real
func (a *EthClientAdapter) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return a.client.TransactionReceipt(ctx, txHash)
}

// ChainID delega ao cliente real
func (a *EthClientAdapter) ChainID(ctx context.Context) (*big.Int, error) {
	return a.client.ChainID(ctx)
}

// SuggestGasPrice delega ao cliente real
func (a *EthClientAdapter) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return a.client.SuggestGasPrice(ctx)
}

// BlockNumber delega ao cliente real
func (a *EthClientAdapter) BlockNumber(ctx context.Context) (uint64, error) {
	return a.client.BlockNumber(ctx)
}

// Close delega ao cliente real
func (a *EthClientAdapter) Close() {
	a.client.Close()
}
