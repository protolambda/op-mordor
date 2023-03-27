package oracle

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type L1Oracle interface {
	// FetchL1Header fetches L1 header
	FetchL1Header(ctx context.Context, blockHash common.Hash) (*types.Header, error)
	// FetchL1BlockTransactions fetches L1 block transactions
	FetchL1BlockTransactions(ctx context.Context, blockHash common.Hash) (types.Transactions, error)
	// FetchL1BlockReceipts fetches L1 block receipts
	FetchL1BlockReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error)
}

type L2StateOracle interface {
	// FetchL2MPTNode fetches L2 state MPT node
	FetchL2MPTNode(ctx context.Context, nodeHash common.Hash) ([]byte, error)
}

type L2Oracle interface {
	L2StateOracle
	// FetchL2Block fetches L2 block with transactions
	FetchL2Block(ctx context.Context, blockHash common.Hash) (*types.Block, error)
}
