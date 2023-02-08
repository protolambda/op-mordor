package main

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type L1Chain struct {
	oracle *PreimageOracle
	head   eth.BlockInfo

	cache map[common.Hash]*PreimageBlockInfo

	// oracle
	// rpc client
}

func NewL1Chain(oracle *PreimageOracle, head eth.BlockInfo) *L1Chain {
	return &L1Chain{
		oracle: oracle,
		cache:  make(map[common.Hash]*PreimageBlockInfo),
		head:   head,
	}
}

// TODO method to iterate back from head to number N

// TODO method to request block header from pre-image oracle

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
