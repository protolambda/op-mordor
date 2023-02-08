package main

import (
	"context"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
)

type L2Engine struct {
	oracle *PreimageOracle
	head   eth.BlockInfo
}

func NewL2Engine(oracle *PreimageOracle, head eth.BlockInfo) *L2Engine {
	return &L2Engine{
		oracle: oracle,
		head:   head,
	}
}

func (l *L2Engine) GetPayload(ctx context.Context, payloadId eth.PayloadID) (*eth.ExecutionPayload, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) ForkchoiceUpdate(ctx context.Context, state *eth.ForkchoiceState, attr *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) NewPayload(ctx context.Context, payload *eth.ExecutionPayload) (*eth.PayloadStatusV1, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) PayloadByHash(ctx context.Context, hash common.Hash) (*eth.ExecutionPayload, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) PayloadByNumber(ctx context.Context, u uint64) (*eth.ExecutionPayload, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	//TODO implement me
	panic("implement me")
}

func (l *L2Engine) SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error) {
	//TODO implement me
	panic("implement me")
}
