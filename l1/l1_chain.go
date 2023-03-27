package l1

import (
	"context"
	"fmt"
	"op-mordor/oracle"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// OracleBackedL1Chain is a wrapper around a oracle.L1Oracle that provides "sugar" to make working with the L1 chain
// data in the oracle easier.
type OracleBackedL1Chain struct {
	oracle oracle.L1Oracle

	head eth.BlockInfo

	headers      map[common.Hash]eth.BlockInfo
	numbers      map[uint64]common.Hash
	transactions map[common.Hash]types.Transactions
	receipts     map[common.Hash]types.Receipts
}

var _ derive.L1Fetcher = (*OracleBackedL1Chain)(nil)

func NewOracleBackedL1Chain(ctx context.Context, oracle oracle.L1Oracle, headHash common.Hash) (*OracleBackedL1Chain, error) {
	l1Header, err := oracle.FetchL1Header(ctx, headHash)
	if err != nil {
		return nil, err
	}
	head := eth.HeaderBlockInfo(l1Header)
	return &OracleBackedL1Chain{
		oracle:       oracle,
		headers:      make(map[common.Hash]eth.BlockInfo),
		transactions: make(map[common.Hash]types.Transactions),
		receipts:     make(map[common.Hash]types.Receipts),
		numbers:      make(map[uint64]common.Hash),
		head:         head,
	}, nil
}

func (l *OracleBackedL1Chain) L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error) {
	return eth.InfoToL1BlockRef(l.head), nil
}

func (l *OracleBackedL1Chain) L1BlockRefByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error) {
	hash, ok := l.numbers[number]
	if ok {
		return l.L1BlockRefByHash(ctx, hash)
	}
	block := l.head
	for block.NumberU64() > number {
		parent, err := l.InfoByHash(ctx, block.ParentHash())
		if err != nil {
			return eth.L1BlockRef{}, fmt.Errorf("l1 block ref by number: %w", err)
		}
		block = parent
		l.numbers[block.NumberU64()] = block.Hash()
	}
	return eth.InfoToL1BlockRef(block), nil
}

func (l *OracleBackedL1Chain) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	info, err := l.InfoByHash(ctx, blockHash)
	if err != nil {
		return nil, nil, err
	}

	receipts, ok := l.receipts[blockHash]
	if ok {
		return info, receipts, nil
	}
	receipts, err = l.oracle.FetchL1BlockReceipts(ctx, blockHash)
	if err != nil {
		return nil, nil, err
	}
	l.receipts[blockHash] = receipts

	return info, receipts, nil
}

func (l *OracleBackedL1Chain) L1BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L1BlockRef, error) {
	info, err := l.InfoByHash(ctx, hash)
	if err != nil {
		return eth.L1BlockRef{}, fmt.Errorf("l1 block ref err: %w", err)
	}
	return eth.InfoToL1BlockRef(info), nil
}

func (l *OracleBackedL1Chain) InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
	return l.headerByHash(ctx, hash)
}

func (l *OracleBackedL1Chain) headerByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
	info, ok := l.headers[hash]
	if ok {
		return info, nil
	}
	header, err := l.oracle.FetchL1Header(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("l1 header err: %w", err)
	}
	info = eth.HeaderBlockInfo(header)
	l.headers[hash] = info
	return info, nil
}

func (l *OracleBackedL1Chain) InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error) {
	header, err := l.headerByHash(ctx, hash)
	if err != nil {
		return nil, nil, err
	}
	txs, ok := l.transactions[hash]
	if ok {
		return header, txs, nil
	}
	txs, err = l.oracle.FetchL1BlockTransactions(ctx, hash)
	if err != nil {
		return nil, nil, err
	}
	l.transactions[hash] = txs
	return header, txs, nil
}
