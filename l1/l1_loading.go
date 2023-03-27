package l1

import (
	"context"
	"fmt"
	"op-mordor/oracle"
	"op-mordor/store"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// LoadingL1Oracle is an implementation of oracle.L1Oracle that loads content from another node via JSON-RPC API.
// Loaded data is written to a store.Store to make the pre-image data available for later execution without needing another node.
type LoadingL1Oracle struct {
	logger log.Logger
	client *ethclient.Client
	store  store.Store
}

var _ oracle.L1Oracle = (*LoadingL1Oracle)(nil)

func (l *LoadingL1Oracle) FetchL1Header(ctx context.Context, blockHash common.Hash) (*types.Header, error) {
	h, err := l.client.HeaderByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	err = l.store.StoreHeader(blockHash, h)
	if err != nil {
		return nil, fmt.Errorf("storing header: %w", err)
	}
	l.logger.Info("Fetched L1 header", "num", h.Number)
	return h, nil
}

func (l *LoadingL1Oracle) FetchL1BlockTransactions(ctx context.Context, blockHash common.Hash) (types.Transactions, error) {
	bl, err := l.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	txs := bl.Transactions()
	err = l.store.StoreTransactions(bl.TxHash(), txs)
	if err != nil {
		return nil, fmt.Errorf("storing transactions: %w", err)
	}
	l.logger.Info("Fetched L1 tx", "num", bl.Number(), "count", txs.Len())
	return txs, nil
}

func (l *LoadingL1Oracle) FetchL1BlockReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	transactions, err := l.FetchL1BlockTransactions(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	var receipts []*types.Receipt
	for _, transaction := range transactions {
		receipt, err := l.client.TransactionReceipt(ctx, transaction.Hash())
		if err != nil {
			return nil, fmt.Errorf("loading receipt for tx %s: %w", transaction.Hash(), err)
		}
		receipts = append(receipts, receipt)
	}
	err = l.store.StoreReceipts(receipts)
	if err != nil {
		return nil, fmt.Errorf("storing receipts: %w", err)
	}
	l.logger.Info("Fetched L1 receipts", "hash", blockHash, "count", len(receipts))
	return receipts, nil
}

var _ oracle.L1Oracle = (*LoadingL1Oracle)(nil)

func NewLoadingL1Chain(logger log.Logger, client *ethclient.Client, store store.Store) oracle.L1Oracle {
	return &LoadingL1Oracle{
		logger: logger,
		client: client,
		store:  store,
	}
}
