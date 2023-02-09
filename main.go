package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/client"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

var _ derive.Engine = (*L2Engine)(nil)
var _ derive.L1Fetcher = (*OracleL1Chain)(nil)

func StateFn(logger log.Logger, l1Hash, l2Hash common.Hash, rpcMode bool) (outputRoot eth.Bytes32, err error) {
	cfg := &chaincfg.Goerli
	var l1Oracle L1PreimageOracle
	var l2Oracle L2PreimageOracle
	// TODO instantiate one of the two oracle modes
	if rpcMode {
		rpcClient, err := rpc.Dial("http://infura:8545") // TODO: Support setting url
		if err != nil {
			return eth.Bytes32{}, fmt.Errorf("l1 rpc unavailable: %w", err)
		}
		ethClient := ethclient.NewClient(rpcClient)

		opRpc := client.NewBaseRPCClient(rpcClient)
		clientConfig := sources.L1ClientDefaultConfig(cfg, false, sources.RPCKindBasic)
		receiptsFetcher, err := sources.NewEthClient(opRpc, logger, nil, &clientConfig.EthClientConfig)
		if err != nil {
			return eth.Bytes32{}, fmt.Errorf("l1 eth client init: %w", err)
		}

		l1Oracle = NewLoadingL1Chain(ethClient, receiptsFetcher)
		l2Oracle = nil // TODO
	} else {
		// TODO disk-mode (or future memory-mapped oracle)
	}

	l1Header, err := l1Oracle.FetchL1Header(context.Background(), l1Hash)
	l2HeadBlock, err := l2Oracle.FetchL2Block(context.Background(), l2Hash)

	l1Head := eth.HeaderBlockInfo(l1Header)
	l2Head := eth.HeaderBlockInfo(l2HeadBlock.Header())

	l1Fetcher := NewOracleL1Chain(l1Oracle, l1Head)
	preDB := NewPreimageBackedDB(nil) // TODO
	l2Genesis := &core.Genesis{}      // TODO
	l2Chain := NewL2Sugar(l2Head, l2Oracle)
	l2Engine := NewL2Engine(logger, l2Genesis, l2Chain, preDB)

	pipeline := derive.NewDerivationPipeline(logger, cfg, l1Fetcher, l2Engine, metrics.NoopMetrics)
	pipeline.Reset()
	for {
		if err := pipeline.Step(context.Background()); errors.Is(err, io.EOF) {
			break
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

	out, err := StateFn(logger, l1Hash, l2Hash, false) // TODO switch between modes
	if err != nil {
		logger.Error("state fn crit err", "err", err)
		os.Exit(1)
	}
	print(out.String())
	os.Exit(0)
}
