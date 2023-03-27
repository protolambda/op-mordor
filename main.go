package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"op-mordor/derivation"
	"op-mordor/l1"
	"op-mordor/l2"
	"op-mordor/oracle"
	"os"

	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

var _ derive.Engine = (*l2.L2Engine)(nil)
var _ derive.L1Fetcher = (*l1.OracleBackedL1Chain)(nil)

//go:embed l2config.json
var l2config []byte

func main() {
	setupEnv()
	logger := log.New()
	logger.SetHandler(log.StderrHandler)

	l1Hash, l2Hash := parseCLIArgs(logger)

	ctx := context.Background()

	var conf params.ChainConfig
	err := json.Unmarshal(l2config, &conf)
	if err != nil {
		panic(fmt.Errorf("invalid l2config json: %w", err))
	}

	// l2 config - genesis.json -
	cfg := &chaincfg.Goerli
	cfg.SeqWindowSize = 20
	cfg.ChannelTimeout = 20

	var l1Oracle oracle.L1Oracle
	var l2Oracle oracle.L2Oracle
	rpcMode := true // TODO switch between modes
	// Instantiate one of the two oracle modes
	if rpcMode {
		l1Oracle, l2Oracle, err = setupRpcOracles(logger)
		if err != nil {
			panic(fmt.Errorf("setting up oracles: %w", err))
		}
	} else {
		// TODO disk-mode (or future memory-mapped oracle)
		panic("non-rpc oracles not implemented yet")
	}

	l1Fetcher, err := l1.NewOracleBackedL1Chain(ctx, l1Oracle, l1Hash)
	if err != nil {
		panic(fmt.Errorf("creating L1: %w", err))
	}
	l2Engine, err := l2.NewL2Engine(ctx, logger, &conf, l2Hash, l2Oracle, cfg)
	if err != nil {
		panic(fmt.Errorf("creating L2: %w", err))
	}

	d := derivation.NewDerivation(logger, cfg, l1Fetcher, l2Engine)
	out, err := d.Run()
	if err != nil {
		logger.Error("state fn crit err", "err", err)
		os.Exit(1)
	}
	print(out.String())
	os.Exit(0)
}

func parseCLIArgs(logger log.Logger) (common.Hash, common.Hash) {
	if len(os.Args) != 3 {
		logger.Error("unexpected number of arguments", "args", len(os.Args))
		os.Exit(1)
	}
	var l1Hash, l2Hash common.Hash
	if err := l1Hash.UnmarshalText([]byte(os.Args[1])); err != nil {
		logger.Error("bad l1 hash input", "err", err)
		os.Exit(1)
	}
	if err := l2Hash.UnmarshalText([]byte(os.Args[2])); err != nil {
		logger.Error("bad l2 hash input", "err", err)
		os.Exit(1)
	}
	return l1Hash, l2Hash
}
