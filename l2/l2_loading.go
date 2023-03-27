package l2

import (
	"context"
	"fmt"
	"op-mordor/oracle"
	"op-mordor/store"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

// LoadingL2Oracle is an implementation of oracle.L2Oracle that loads content from another node via JSON-RPC API.
// Loaded data is written to a store.BlockStore to make the pre-image data available for later execution without needing another node.
type LoadingL2Oracle struct {
	logger    log.Logger
	rpcClient *rpc.Client
	client    *ethclient.Client
	store     store.BlockStore
	source    store.BlockSource
}

var _ oracle.L2Oracle = (*LoadingL2Oracle)(nil)

func NewLoadingL2Chain(logger log.Logger, l2RpcClient *rpc.Client, sstore store.Store, source store.Source) *LoadingL2Oracle {
	return &LoadingL2Oracle{
		logger:    logger,
		rpcClient: l2RpcClient,
		client:    ethclient.NewClient(l2RpcClient),
		store:     store.BlockStore{Store: sstore},
		source:    store.BlockSource{Source: source},
	}
}

// FetchL2MPTNode fetches L2 state MPT node
func (l *LoadingL2Oracle) FetchL2MPTNode(ctx context.Context, nodeHash common.Hash) ([]byte, error) {
	snode, err := l.source.ReadNode(nodeHash)
	if err == nil {
		return snode, nil
	} else if !store.IsNoDataError(err) {
		return nil, fmt.Errorf("restoring node: %w", err)
	}
	var node hexutil.Bytes
	err = l.rpcClient.CallContext(ctx, &node, "debug_dbGet", nodeHash.Hex())
	if err != nil {
		return nil, err
	}
	err = l.store.StoreNode(nodeHash, node)
	l.logger.Debug("Loaded node", "key", nodeHash, "val", node)
	return node, err
}

// FetchL2Block fetches L2 block with transactions
func (l *LoadingL2Oracle) FetchL2Block(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	// TODO: try blocksource first; after ReadTransactions is implemented
	block, err := l.client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, err
	}
	err = l.store.StoreBlock(block)
	l.logger.Info("Fetch L2 block", "num", block.NumberU64())
	return block, err
}
