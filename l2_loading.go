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
}

func NewLoadingL2Chain(rpcClient *rpc.Client) *LoadingL2Chain {
	return &LoadingL2Chain{
		rpcClient: rpcClient,
		client:    ethclient.NewClient(rpcClient),
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
	// TODO: persister block pre-image to disk
	return block, nil
}
