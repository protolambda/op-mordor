package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
)

var _ derive.Engine = (*L2Engine)(nil)
var _ derive.L1Fetcher = (*OracleL1Chain)(nil)

//go:embed l2config.json
var l2config []byte

func StateFn(logger log.Logger, l1Hash, l2Hash common.Hash, rpcMode bool) (outputRoot eth.Bytes32, err error) {
	cfg := &chaincfg.Goerli
	// l2 config - genesis.json -
	var l1Oracle L1PreimageOracle
	var l2Oracle L2PreimageOracle
	// TODO instantiate one of the two oracle modes
	if rpcMode {
		l1Oracle, l2Oracle, err = setupRpcOracles(logger)
		if err != nil {
			return eth.Bytes32{}, fmt.Errorf("setting up oracles: %w", err)
		}
	} else {
		// TODO disk-mode (or future memory-mapped oracle)
		panic("non-rpc oracles not implemented yet")
	}

	l1Header, err := l1Oracle.FetchL1Header(context.Background(), l1Hash)
	l2HeadBlock, err := l2Oracle.FetchL2Block(context.Background(), l2Hash)

	l1Head := eth.HeaderBlockInfo(l1Header)
	l2Head := eth.HeaderBlockInfo(l2HeadBlock.Header())

	l1Fetcher := NewOracleL1Chain(l1Oracle, l1Head)
	preDB := NewPreimageBackedDB(l2Oracle)
	var conf params.ChainConfig
	err = json.Unmarshal(l2config, &conf)
	if err != nil {
		panic(fmt.Errorf("invalid l2config json: %w", err))
	}
	l2Chain := NewL2Sugar(l2Head, l2Oracle, cfg)
	l2Engine := NewL2Engine(logger, &conf, l2Chain, preDB)

	pipeline := derive.NewDerivationPipeline(logger, cfg, l1Fetcher, l2Engine, metrics.NoopMetrics)
	pipeline.Reset()
	for {
		if err := pipeline.Step(context.Background()); errors.Is(err, io.EOF) {
			break
		} else if errors.Is(err, derive.ErrTemporary) {
			logger.Warn("Temporary error in pipeline", "err", err)
			time.Sleep(5 * time.Second)
			continue
		} else if errors.Is(err, derive.NotEnoughData) {
			logger.Debug("Data is lacking")
			continue
		} else if err != nil {
			return eth.Bytes32{}, fmt.Errorf("pipeline err: %w", err)
		}
	}

	l2OutputVersion := eth.Bytes32{}
	outBlock := l2Engine.sugar.CurrentBlock()
	stateDB, err := state.New(outBlock.Root(), state.NewDatabase(preDB), nil)
	if err != nil {
		return eth.Bytes32{}, fmt.Errorf("failed to open L2 state db at block %s: %w", outBlock.Hash(), err)
	}
	withdrawalsTrie := stateDB.StorageTrie(predeploys.L2ToL1MessagePasserAddr)
	return rollup.ComputeL2OutputRoot(l2OutputVersion, outBlock.Hash(), outBlock.Root(), withdrawalsTrie.Hash()), nil
}

func main() {
	setupEnv()
	logger := log.New()
	logger.SetHandler(log.StderrHandler)

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

	out, err := StateFn(logger, l1Hash, l2Hash, true) // TODO switch between modes
	if err != nil {
		logger.Error("state fn crit err", "err", err)
		os.Exit(1)
	}
	print(out.String())
	os.Exit(0)
}
