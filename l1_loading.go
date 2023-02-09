package main

import (
	"context"
	"fmt"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type LoadingL1Chain struct {
	oracle OracleL1Chain
	client ethclient.Client
}

func (l LoadingL1Chain) L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error) {
	return l.oracle.L1BlockRefByLabel(ctx, label)
}

func (l LoadingL1Chain) L1BlockRefByNumber(ctx context.Context, number uint64) (eth.L1BlockRef, error) {
	ref, err := l.oracle.L1BlockRefByNumber(ctx, number)
	if err == ErrMissingData {
		block := l.oracle.head
		for block.NumberU64() > number {
			parent, err := l.InfoByHash(ctx, block.ParentHash())
			if err != nil {
				return eth.L1BlockRef{}, fmt.Errorf("l1 block ref by number: %w", err)
			}
			block = parent
		}
		return eth.InfoToL1BlockRef(block), nil
	} else if err != nil {
		return eth.L1BlockRef{}, nil
	}
	return ref, nil
}

func (l LoadingL1Chain) L1BlockRefByHash(ctx context.Context, hash common.Hash) (eth.L1BlockRef, error) {
	info, err := l.InfoByHash(ctx, hash)
	if err != nil {
		return eth.L1BlockRef{}, err
	}
	return eth.InfoToL1BlockRef(info), nil
}

func (l LoadingL1Chain) InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error) {
	ref, err := l.oracle.InfoByHash(ctx, hash)
	if err == ErrMissingData {
		header, err := l.client.HeaderByHash(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("fetch failed: %w", err)
		}
		err = l.oracle.StoreBlock(ctx, hash, header)
		if err != nil {
			return nil, fmt.Errorf("fetch info by hash: %w", err)
		}
		return eth.HeaderBlockInfo(header), err
	} else if err != nil {
		return nil, fmt.Errorf("oracle error: %w", err)
	}
	return ref, nil
}

func (l LoadingL1Chain) InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error) {
	info, txs, err := l.oracle.InfoAndTxsByHash(ctx, hash)
	if err == ErrMissingData {
		block, err := l.client.BlockByHash(ctx, hash)
		if err != nil {
			return nil, nil, fmt.Errorf("fetch failed: %w", err)
		}
		err = l.oracle.StoreTransactions(ctx, hash, block.Header(), block.Transactions())
		if err != nil {
			return nil, nil, fmt.Errorf("fetch info and txs by hash: %w", err)
		}
		return eth.HeaderBlockInfo(block.Header()), block.Transactions(), err
	} else if err != nil {
		return nil, nil, fmt.Errorf("oracle error: %w", err)
	}
	return info, txs, nil
}

func (l LoadingL1Chain) FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error) {
	//TODO implement me
	panic("implement me")
}

func NewLoadingL1Chain(oracle OracleL1Chain, client ethclient.Client) derive.L1Fetcher {
	return &LoadingL1Chain{
		oracle: oracle,
		client: client,
	}
}
