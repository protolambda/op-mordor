package main

import (
	"context"
	"op-mordor/store"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type LoadingL2Chain struct {
	logger    log.Logger
	rpcClient *rpc.Client
	client    *ethclient.Client
	store     store.BlockStore
}

func NewLoadingL2Chain(logger log.Logger, l2RpcClient *rpc.Client, sstore store.Store) *LoadingL2Chain {
	return &LoadingL2Chain{
		logger:    logger,
		rpcClient: l2RpcClient,
		client:    ethclient.NewClient(l2RpcClient),
		store:     store.BlockStore{Store: sstore},
	}
}

// FetchL2MPTNode fetches L2 state MPT node
func (l *LoadingL2Chain) FetchL2MPTNode(ctx context.Context, nodeHash common.Hash) ([]byte, error) {
	var node hexutil.Bytes
	err := l.rpcClient.CallContext(ctx, &node, "debug_dbGet", nodeHash.Hex())
	if err != nil {
		return nil, err
	}
	err = l.store.StoreNode(nodeHash, node)
	return node, err
}

// FetchL2Block fetches L2 block with transactions
func (l *LoadingL2Chain) FetchL2Block(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	block, err := l.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	err = l.store.StoreBlock(block)
	l.logger.Info("Fetch L2 block", "num", block.NumberU64())
	return block, err
}
