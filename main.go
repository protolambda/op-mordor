package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"os"
)

var _ derive.Engine = (*L2Engine)(nil)
var _ derive.L1Fetcher = (*L1Chain)(nil)

func StateFn(logger log.Logger, l1Hash, l2Hash common.Hash) (outputRoot common.Hash, err error) {
	l1Fetcher := &L1Chain{}
	l2Engine := &L2Engine{}

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
