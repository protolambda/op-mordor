package main

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type LoadingL1Chain struct {
	client          *ethclient.Client
	receiptsFetcher derive.L1ReceiptsFetcher
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
	_, rcpts, err := l.receiptsFetcher.FetchReceipts(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	// TODO: persist receipts to disk
	return rcpts, nil
}

var _ L1PreimageOracle = (*LoadingL1Chain)(nil)

func NewLoadingL1Chain(client *ethclient.Client, receiptsFetcher derive.L1ReceiptsFetcher) L1PreimageOracle {
	return &LoadingL1Chain{
		client:          client,
		receiptsFetcher: receiptsFetcher,
	}
}
