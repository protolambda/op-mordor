package main

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type L2Sugar struct {
	head eth.BlockInfo

	oracle L2PreimageOracle
}

func NewL2Sugar(head eth.BlockInfo, oracle L2PreimageOracle) *L2Sugar {
	return &L2Sugar{
		head:   head,
		oracle: oracle,
	}
}

func (l *L2Sugar) SetHead(head eth.BlockInfo) {
	l.head = head
}

func (l *L2Sugar) CurrentBlock() *types.Block {
	// TODO
	return nil
}

func (l *L2Sugar) getBlockByHash(hash common.Hash) *types.Block {
	// TODO
	panic("TODO")
}

func (l *L2Sugar) getHeaderByHash(hash common.Hash) *types.Header {
	// TODO
	panic("TODO")
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
