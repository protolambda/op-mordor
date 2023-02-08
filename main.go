package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"io"
	"os"
)

var _ derive.Engine = (*L2Engine)(nil)
var _ derive.L1Fetcher = (*L1Chain)(nil)

func StateFn(logger log.Logger, l1Hash, l2Hash common.Hash) (outputRoot common.Hash, err error) {

	oracle := &PreimageOracle{}

	getHeader := func(blockHash common.Hash) (eth.BlockInfo, error) {
		headInfoRrlp, err := oracle.GetPreimage(blockHash)
		if err != nil {
			return nil, fmt.Errorf("l1 head preimage err: %w", err)
		}
		var h types.Header
		if err := rlp.Decode(bytes.NewReader(headInfoRrlp), &h); err != nil {
			return nil, fmt.Errorf("l1 head decode err: %w", err)
		}
		return eth.HeaderBlockInfo(&h), nil
	}

	l1Head, err := getHeader(l1Hash)
	// TODO
	l2Head, err := getHeader(l2Hash)
	// TODO
	l1Fetcher := NewL1Chain(oracle, l1Head)
	l2BlockFetcher := NewL2BlockFetcher(oracle)
	l2Genesis := &core.Genesis{} // TODO
	l2Engine := NewL2Engine(logger, l2Genesis, l2BlockFetcher, l2Head)

	cfg := &chaincfg.Goerli
	pipeline := derive.NewDerivationPipeline(logger, cfg, l1Fetcher, l2Engine, metrics.NoopMetrics)
	pipeline.Reset()
	for {
		if err := pipeline.Step(context.Background()); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return common.Hash{}, fmt.Errorf("pipeline err: %w", err)
		}
	}

	//rollup.ComputeL2OutputRoot(eth.Bytes32{}, out)
	return common.Hash{}, nil
}

func main() {
	// TODO: maybe parse from arg
	l1Hash := common.Hash{}
	l2Hash := common.Hash{}

	logger := log.New()
	logger.SetHandler(log.StdoutHandler)

	out, err := StateFn(logger, l1Hash, l2Hash)
	if err != nil {
		logger.Error("state fn crit err", "err", err)
		os.Exit(1)
	}
	print(out.String())
}
