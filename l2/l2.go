package l2

import (
	"context"
	"op-mordor/derivation"
	"op-mordor/oracle"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

type L2Engine struct {
	*EngineAPI
	*OracleBackedL2Chain
}

var _ derivation.L2Access = (*L2Engine)(nil)

func NewL2Engine(ctx context.Context, log log.Logger, cfg *params.ChainConfig, l2Hash common.Hash, l2Oracle oracle.L2Oracle, rollupCfg *rollup.Config) (*L2Engine, error) {

	l2HeadBlock, err := l2Oracle.FetchL2Block(ctx, l2Hash)
	if err != nil {
		return nil, err
	}

	l2Head := eth.HeaderBlockInfo(l2HeadBlock.Header())

	l2Chain := NewOracleBackedL2Chain(l2Head, l2Oracle, rollupCfg)
	preDB := NewOracleBackedDB(l2Oracle)
	return &L2Engine{
		EngineAPI:           NewEngineAPI(log, cfg, l2Chain, preDB),
		OracleBackedL2Chain: l2Chain,
	}, nil
}
