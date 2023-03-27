package l2

import (
	"context"
	"log"
	"op-mordor/oracle"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// OracleBackedL2Chain is a wrapper around a oracle.L2Oracle that provides "sugar" to make working with the L2 chain
// data in the oracle easier.
type OracleBackedL2Chain struct {
	oracle  oracle.L2Oracle
	cfg     *rollup.Config
	genesis *rollup.Genesis
	ctx     context.Context

	head   eth.BlockInfo
	blocks map[common.Hash]*types.Block
}

func NewOracleBackedL2Chain(
	head eth.BlockInfo,
	oracle oracle.L2Oracle,
	cfg *rollup.Config,
) *OracleBackedL2Chain {
	return &OracleBackedL2Chain{
		oracle: oracle,
		cfg:    cfg,
		ctx:    context.TODO(),
		head:   head,

		blocks: make(map[common.Hash]*types.Block),
	}
}

func (l *OracleBackedL2Chain) handleErr(err error) {
	log.Fatalf("OracleBackedL2Chain fatal error: %v", err)
}

func (l *OracleBackedL2Chain) SetHead(head eth.BlockInfo) {
	l.head = head
}

func (l *OracleBackedL2Chain) currentBlock() eth.BlockInfo {
	return l.head
}

func (l *OracleBackedL2Chain) getBlockByHash(hash common.Hash) *types.Block {
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
func (l *OracleBackedL2Chain) getBlockByNumber(u uint64) *types.Block {
	block := l.getBlockByHash(l.head.Hash())
	for block.NumberU64() > u {
		block = l.getBlockByHash(block.ParentHash())
	}
	return block
}

func (l *OracleBackedL2Chain) getBlockInfoByHash(hash common.Hash) eth.BlockInfo {
	return eth.HeaderBlockInfo(l.getHeaderByHash(hash))
}

func (l *OracleBackedL2Chain) getHeaderByHash(hash common.Hash) *types.Header {
	return l.getBlockByHash(hash).Header()
}

func (l *OracleBackedL2Chain) getBlockHashByNumber(u uint64) common.Hash {
	return l.getBlockByNumber(u).Hash()
}

// used by geth chain context
func (l *OracleBackedL2Chain) getHeader(hash common.Hash, _ uint64) *types.Header {
	return l.getHeaderByHash(hash)
}

func (l *OracleBackedL2Chain) PayloadByHash(_ context.Context, hash common.Hash) (*eth.ExecutionPayload, error) {
	block := l.getBlockByHash(hash)
	return eth.BlockAsPayload(block)
}

func (l *OracleBackedL2Chain) PayloadByNumber(ctx context.Context, u uint64) (*eth.ExecutionPayload, error) {
	return l.PayloadByHash(ctx, l.getBlockHashByNumber(u))
}

func (l *OracleBackedL2Chain) L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error) {
	return l.L2BlockRefByHash(ctx, l.head.Hash())
}

func (l *OracleBackedL2Chain) L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error) {
	payload, err := l.PayloadByHash(ctx, l2Hash)
	if err != nil {
		return eth.L2BlockRef{}, err
	}
	return derive.PayloadToBlockRef(payload, &l.cfg.Genesis)
}

func (l *OracleBackedL2Chain) SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error) {
	payload, err := l.PayloadByHash(ctx, hash)
	if err != nil {
		return eth.SystemConfig{}, err
	}
	return derive.PayloadToSystemConfig(payload, l.cfg)
}
