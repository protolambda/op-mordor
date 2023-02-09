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
}

func (l *LoadingL1Chain) FetchL1Header(ctx context.Context, blockHash common.Hash) (*types.Header, error) {
	h, err := l.client.HeaderByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	// TODO: persist header pre-image to disk
	return h, nil
}

func (l *LoadingL1Chain) FetchL1BlockTransactions(ctx context.Context, blockHash common.Hash) (types.Transactions, error) {
	bl, err := l.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	// TODO: persist transactions to disk
	return bl.Transactions(), nil
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
	// TODO: persist receipts to disk
	return receipts, nil
}

var _ L1PreimageOracle = (*LoadingL1Chain)(nil)

func NewLoadingL1Chain(client *ethclient.Client) L1PreimageOracle {
	return &LoadingL1Chain{
		client: client,
	}
}
