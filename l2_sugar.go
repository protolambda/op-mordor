package main

import (
	"context"
	"log"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type L2Sugar struct {
	oracle  L2PreimageOracle
	cfg     *rollup.Config
	genesis *rollup.Genesis
	ctx     context.Context

	head   eth.BlockInfo
	blocks map[common.Hash]*types.Block
}

func NewL2Sugar(
	head eth.BlockInfo,
	oracle L2PreimageOracle,
	cfg *rollup.Config,
) *L2Sugar {
	return &L2Sugar{
		oracle: oracle,
		cfg:    cfg,
		ctx:    context.TODO(),
		head:   head,

		blocks: make(map[common.Hash]*types.Block),
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

func (l *L2Sugar) getBlockByHash(hash common.Hash) *types.Block {
	block, ok := l.blocks[hash]
	if ok {
		return block
	}

	block, err := l.oracle.FetchL2Block(l.ctx, hash)
	if err != nil {
		l.handleErr(err)
		return nil
	}
	l.blocks[hash] = block

	return block
}

// getBlockByNumber iterates back from the head until the block with number u is
// found. It uses getBlockByHash so all blocks retrieved this way are locally
// cached.
func (l *L2Sugar) getBlockByNumber(u uint64) *types.Block {
	block := l.getBlockByHash(l.head.Hash())
	for block.NumberU64() > u {
		block = l.getBlockByHash(block.ParentHash())
	}
	return block
}

func (l *L2Sugar) getBlockInfoByHash(hash common.Hash) eth.BlockInfo {
	return eth.HeaderBlockInfo(l.getHeaderByHash(hash))
}

func (l *L2Sugar) getHeaderByHash(hash common.Hash) *types.Header {
	return l.getBlockByHash(hash).Header()
}

func (l *L2Sugar) getBlockHashByNumber(u uint64) common.Hash {
	return l.getBlockByNumber(u).Hash()
}

// used by geth chain context
func (l *L2Sugar) getHeader(hash common.Hash, _ uint64) *types.Header {
	return l.getHeaderByHash(hash)
}

func (l *L2Sugar) PayloadByHash(_ context.Context, hash common.Hash) (*eth.ExecutionPayload, error) {
	block := l.getBlockByHash(hash)
	return eth.BlockAsPayload(block)
}

func (l *L2Sugar) PayloadByNumber(ctx context.Context, u uint64) (*eth.ExecutionPayload, error) {
	return l.PayloadByHash(ctx, l.getBlockHashByNumber(u))
}

func (l *L2Sugar) L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	return l.L2BlockRefByHash(ctx, l.head.Hash())
}

func (l *L2Sugar) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	payload, err := l.PayloadByHash(ctx, l2Hash)
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	return derive.PayloadToBlockRef(payload, &l.cfg.Genesis)
}

func (l *L2Sugar) SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error) {
	payload, err := l.PayloadByHash(ctx, hash)
	if err != nil {
		return eth.SystemConfig{}, err
	}
	return derive.PayloadToSystemConfig(payload, l.cfg)
}
