package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type LoadingL1Chain struct {
	client *ethclient.Client
	store  Store
}

func (l *LoadingL1Chain) FetchL1Header(ctx context.Context, blockHash common.Hash) (*types.Header, error) {
	h, err := l.client.HeaderByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	err = l.store.StoreHeader(blockHash, h)
	if err != nil {
		return nil, fmt.Errorf("storing header: %w", err)
	}
	return h, nil
}

func (l *LoadingL1Chain) FetchL1BlockTransactions(ctx context.Context, blockHash common.Hash) (types.Transactions, error) {
	bl, err := l.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	txs := bl.Transactions()
	err = l.store.StoreTransactions(bl.TxHash(), txs)
	if err != nil {
		return nil, fmt.Errorf("storing transactions: %w", err)
	}
	return txs, nil
}

func (l *LoadingL1Chain) FetchL1BlockReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
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
	return receipts, nil
}

var _ L1PreimageOracle = (*LoadingL1Chain)(nil)

func NewLoadingL1Chain(client *ethclient.Client, store Store) L1PreimageOracle {
	return &LoadingL1Chain{
		client: client,
		store:  store,
	}
}
