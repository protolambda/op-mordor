package main

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type L1Chain struct {
}

func (l *L1Chain) L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L1Chain) L1BlockRefByNumber(ctx context.Context, u uint64) (eth.L1BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L1Chain) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L1Chain) L1BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L1BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L1Chain) InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L1Chain) InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error) {
	//TODO implement me
	panic("implement me")
}
