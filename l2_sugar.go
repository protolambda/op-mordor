package main

import (
	"context"
	"log"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type L2Sugar struct {
	oracle L2PreimageOracle
	ctx    context.Context

	head   eth.BlockInfo
	blocks map[common.Hash]*types.Block
}

func NewL2Sugar(head eth.BlockInfo, oracle L2PreimageOracle) *L2Sugar {
	return &L2Sugar{
		ctx:    context.TODO(),
		head:   head,
		oracle: oracle,
	}
}

func (l *L2Sugar) handleErr(err error) {
	log.Fatalf("L2Sugar fatal error: %v", err)
}

func (l *L2Sugar) SetHead(head eth.BlockInfo) {
	l.head = head
}

func (l *L2Sugar) CurrentBlock() eth.BlockInfo {
	return l.head
}

func (l *L2Sugar) getBlockInfoByHash(hash common.Hash) eth.BlockInfo {
	return eth.HeaderBlockInfo(l.getHeaderByHash(hash))
}

func (l *L2Sugar) getHeaderByHash(hash common.Hash) *types.Header {
	block, ok := l.blocks[hash]
	if ok {
		return block.Header()
	}

	block, err := l.oracle.FetchL2Block(l.ctx, hash)
	if err != nil {
		l.handleErr(err)
		return nil
	}
	l.blocks[hash] = block

	return block.Header()
}

func (l *L2Sugar) getBlockHashByNumber(u uint64) common.Hash {
	// TODO
	panic("TODO")
}

// used by geth chain context
func (l *L2Sugar) getHeader(hash common.Hash, u uint64) *types.Header {
	// TODO
	panic("TODO")
}

func (l *L2Sugar) PayloadByHash(ctx context.Context, hash common.Hash) (*eth.ExecutionPayload, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Sugar) PayloadByNumber(ctx context.Context, u uint64) (*eth.ExecutionPayload, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Sugar) L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Sugar) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Sugar) SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error) {
	//TODO implement me
	panic("implement me")
}
