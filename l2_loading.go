package main

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type LoadingL2Chain struct {
	rpcClient *rpc.Client
	client    *ethclient.Client
	store     BlockStore
}

func NewLoadingL2Chain(l2RpcClient *rpc.Client, store BlockStore) *LoadingL2Chain {
	return &LoadingL2Chain{
		rpcClient: l2RpcClient,
		client:    ethclient.NewClient(l2RpcClient),
		store:     store,
	}
}

// FetchL2MPTNode fetches L2 state MPT node
func (l *LoadingL2Chain) FetchL2MPTNode(ctx context.Context, nodeHash common.Hash) ([]byte, error) {
	var resp hexutil.Bytes
	err := l.rpcClient.CallContext(ctx, &resp, "debug_dbGet", nodeHash.Hex())
	// TODO: persist node pre-image to disk
	return resp, err
}

// FetchL2Block fetches L2 block with transactions
func (l *LoadingL2Chain) FetchL2Block(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	block, err := l.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	err = l.store.StoreBlock(block)
	return block, err
}

// BlockStore augments a Store with a method StoreBlock to store blocks.
// It reuses header and transactions storing of the Store.
type BlockStore struct {
	Store
}

func (s BlockStore) StoreBlock(block *types.Block) error {
	if err := s.StoreHeader(block.Hash(), block.Header()); err != nil {
		return err
	}
	return s.StoreTransactions(block.TxHash(), block.Transactions())
}
