package main

import (
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
)

type L2Engine struct {
	*EngineAPI
	*L2Sugar
}

var _ derive.Engine = (*L2Engine)(nil)

func NewL2Engine(log log.Logger, cfg *core.Genesis, fetcher *L2BlockFetcher, preDB *PreimageBackedDB, head eth.BlockInfo) *L2Engine {
	sugar := NewL2Sugar(head, fetcher)
	return &L2Engine{
		EngineAPI: NewEngineAPI(log, cfg, sugar, preDB),
		L2Sugar:   sugar,
	}
}
