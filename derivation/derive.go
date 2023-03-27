package derivation

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/log"
)

type L2Access interface {
	derive.Engine
	L2OutputRoot() (eth.Bytes32, error)
}

type Derivation struct {
	logger    log.Logger
	l1        derive.L1Fetcher
	l2        L2Access
	rollupCfg *rollup.Config
}

func NewDerivation(logger log.Logger, rollupCfg *rollup.Config, l1 derive.L1Fetcher, l2 L2Access) *Derivation {
	return &Derivation{
		logger:    logger,
		rollupCfg: rollupCfg,
		l1:        l1,
		l2:        l2,
	}
}

func (d *Derivation) Run() (eth.Bytes32, error) {
	err := d.runDerivation()
	if err != nil {
		return eth.Bytes32{}, err
	}
	return d.l2.L2OutputRoot()
}

func (d *Derivation) runDerivation() error {
	pipeline := derive.NewDerivationPipeline(d.logger, d.rollupCfg, d.l1, d.l2, metrics.NoopMetrics)
	pipeline.Reset()
	for {
		if err := pipeline.Step(context.Background()); errors.Is(err, io.EOF) {
			return nil
		} else if errors.Is(err, derive.ErrTemporary) {
			d.logger.Warn("Temporary error in pipeline", "err", err)
			time.Sleep(5 * time.Second)
			continue
		} else if errors.Is(err, derive.NotEnoughData) {
			d.logger.Debug("Data is lacking")
			continue
		} else if err != nil {
			return fmt.Errorf("pipeline err: %w", err)
		}
	}
}
