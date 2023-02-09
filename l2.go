package main

import (
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

type L2Engine struct {
	*EngineAPI
	*L2Sugar
}

var _ derive.Engine = (*L2Engine)(nil)

func NewL2Engine(log log.Logger, cfg *params.ChainConfig, ch *L2Sugar, preDB *PreimageBackedDB) *L2Engine {
	return &L2Engine{
		EngineAPI: NewEngineAPI(log, cfg, ch, preDB),
		L2Sugar:   ch,
	}
}
