package main

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type OracleL1Chain struct {
	oracle L1PreimageOracle

	head eth.BlockInfo

	headers      map[common.Hash]eth.BlockInfo
	numbers      map[uint64]common.Hash
	transactions map[common.Hash]types.Transactions
	receipts     map[common.Hash]types.Receipts
}

func NewOracleL1Chain(oracle L1PreimageOracle, head eth.BlockInfo) *OracleL1Chain {
	return &OracleL1Chain{
		oracle:       oracle,
		headers:      make(map[common.Hash]eth.BlockInfo),
		transactions: make(map[common.Hash]types.Transactions),
		receipts:     make(map[common.Hash]types.Receipts),
		head:         head,
	}
}

func (l *OracleL1Chain) L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error) {
	return eth.InfoToL1BlockRef(l.head), nil
}

func (l *OracleL1Chain) L1BlockRefByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error) {
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

func (l *OracleL1Chain) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
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

func (l *OracleL1Chain) L1BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L1BlockRef, error) {
	info, err := l.InfoByHash(ctx, hash)
	if err != nil {
		return eth.L1BlockRef{}, fmt.Errorf("l1 block ref err: %w", err)
	}
	return eth.InfoToL1BlockRef(info), nil
}

func (l *OracleL1Chain) InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
	return l.headerByHash(ctx, hash)
}

func (l *OracleL1Chain) headerByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
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

func (l *OracleL1Chain) InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error) {
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
