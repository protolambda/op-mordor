package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type L1Chain struct {
	oracle *PreimageOracle
	head   eth.BlockInfo

	cache map[common.Hash]*PreimageBlockInfo
}

func NewL1Chain(oracle *PreimageOracle, head eth.BlockInfo) *L1Chain {
	return &L1Chain{
		oracle: oracle,
		cache:  make(map[common.Hash]*PreimageBlockInfo),
		head:   head,
	}
}

func (l *L1Chain) L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error) {
	return eth.InfoToL1BlockRef(l.head), nil
}

func (l *L1Chain) L1BlockRefByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error) {
	block := l.head
	for block.NumberU64() > number {
		parent, err := l.InfoByHash(ctx, block.ParentHash())
		if err != nil {
			return eth.L1BlockRef{}, fmt.Errorf("l1 block ref by number: %w", err)
		}
		block = parent
	}
	return eth.InfoToL1BlockRef(block), nil
}

func (l *L1Chain) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	info, err := l.InfoByHash(ctx, blockHash)
	if err != nil {
		return nil, nil, err
	}

	receiptRlp, err := l.oracle.GetPreimage(info.ReceiptHash())
	if err != nil {
		return nil, nil, err
	}

	receipts := types.Receipts{}

	if err = rlp.Decode(bytes.NewReader(receiptRlp), receipts); err != nil {
		return nil, nil, err
	}
	return info, receipts, nil
}

func (l *L1Chain) L1BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L1BlockRef, error) {
	info, err := l.InfoByHash(ctx, hash)
	if err != nil {
		return eth.L1BlockRef{}, fmt.Errorf("l1 block ref err: %w", err)
	}
	return eth.InfoToL1BlockRef(info), nil
}

func (l *L1Chain) InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
	header, err := l.headerByHash(hash)
	if err != nil {
		return nil, err
	}
	return eth.HeaderBlockInfo(&header), nil
}

func (l *L1Chain) headerByHash(hash common.Hash) (types.Header, error) {
	headInfoRlp, err := l.oracle.GetPreimage(hash)
	if err != nil {
		return types.Header{}, fmt.Errorf("l1 head preimage err: %w", err)
	}
	var h types.Header
	if err := rlp.Decode(bytes.NewReader(headInfoRlp), &h); err != nil {
		return types.Header{}, fmt.Errorf("l1 head decode err: %w", err)
	}
	return h, nil
}

func (l *L1Chain) InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error) {
	header, err := l.headerByHash(hash)
	if err != nil {
		return nil, nil, err
	}
	txRlp, err := l.oracle.GetPreimage(header.TxHash)
	if err != nil {
		return nil, nil, err
	}
	txs := types.Transactions{}
	err = rlp.Decode(bytes.NewReader(txRlp), txs)
	if err != nil {
		return nil, nil, err
	}
	return eth.HeaderBlockInfo(&header), txs, nil
}
