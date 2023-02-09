package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	beaconConsensus "github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/beacon"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trie"
)

type HeaderInterface interface {
	ByHash(hash common.Hash, r [32]byte) *types.Header
	ByNum(hash common.Hash, u uint64) *types.Header
}

type EngineChainContext struct {
	eng       consensus.Engine
	getHeader func(hash common.Hash, u uint64) *types.Header
}

func (e EngineChainContext) Engine() consensus.Engine {
	return e.eng
}

func (e EngineChainContext) GetHeader(hash common.Hash, u uint64) *types.Header {
	return e.getHeader(hash, u)
}

var _ core.ChainContext = (*EngineChainContext)(nil)

type EngineAPI struct {
	log log.Logger

	sugar    *L2Sugar
	chainCtx *EngineChainContext
	vmCfg    vm.Config

	safe      common.Hash
	finalized eth.BlockID

	// L2 evm / chain
	l2Database ethdb.Database
	l2Cfg      *core.Genesis

	// L2 block building data
	l2BuildingHeader *types.Header             // block header that we add txs to for block building
	l2BuildingState  *state.StateDB            // state used for block building
	l2GasPool        *core.GasPool             // track gas used of ongoing building
	pendingIndices   map[common.Address]uint64 // per account, how many txs from the pool were already included in the block, since the pool is lagging behind block mining.
	l2Transactions   []*types.Transaction      // collects txs that were successfully included into current block build
	l2Receipts       []*types.Receipt          // collect receipts of ongoing building
	l2ForceEmpty     bool                      // when no additional txs may be processed (i.e. when sequencer drift runs out)
	l2TxFailed       []*types.Transaction      // log of failed transactions which could not be included

	payloadID beacon.PayloadID // ID of payload that is currently being built
}

func NewEngineAPI(log log.Logger, cfg *core.Genesis, sugar *L2Sugar, preDB *PreimageBackedDB) *EngineAPI {
	cons := beaconConsensus.New(nil)

	return &EngineAPI{
		log:   log,
		sugar: sugar,
		chainCtx: &EngineChainContext{
			eng:       cons,
			getHeader: sugar.getHeader,
		},
		vmCfg:      vm.Config{},
		safe:       sugar.head.Hash(),
		finalized:  eth.BlockID{Hash: sugar.head.Hash(), Number: sugar.head.NumberU64()},
		l2Database: preDB,
		l2Cfg:      cfg,
		// building state starts nil
	}
}

var (
	STATUS_INVALID         = &eth.ForkchoiceUpdatedResult{PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionInvalid}, PayloadID: nil}
	STATUS_SYNCING         = &eth.ForkchoiceUpdatedResult{PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionSyncing}, PayloadID: nil}
	INVALID_TERMINAL_BLOCK = eth.PayloadStatusV1{Status: eth.ExecutionInvalid, LatestValidHash: &common.Hash{}}
)

// computePayloadId computes a pseudo-random payloadid, based on the parameters.
func computePayloadId(headBlockHash common.Hash, params *eth.PayloadAttributes) beacon.PayloadID {
	// Hash
	hasher := sha256.New()
	hasher.Write(headBlockHash[:])
	_ = binary.Write(hasher, binary.BigEndian, params.Timestamp)
	hasher.Write(params.PrevRandao[:])
	hasher.Write(params.SuggestedFeeRecipient[:])
	_ = binary.Write(hasher, binary.BigEndian, params.NoTxPool)
	_ = binary.Write(hasher, binary.BigEndian, uint64(len(params.Transactions)))
	for _, tx := range params.Transactions {
		_ = binary.Write(hasher, binary.BigEndian, uint64(len(tx))) // length-prefix to avoid collisions
		hasher.Write(tx)
	}
	_ = binary.Write(hasher, binary.BigEndian, *params.GasLimit)
	var out beacon.PayloadID
	copy(out[:], hasher.Sum(nil)[:8])
	return out
}

func (ea *EngineAPI) setFinalized(id eth.BlockID) {
	ea.finalized = id
}

func (ea *EngineAPI) setSafe(h common.Hash) {
	ea.safe = h
}

func (ea *EngineAPI) startBlock(parent common.Hash, params *eth.PayloadAttributes) error {
	if ea.l2BuildingHeader != nil {
		ea.log.Warn("started building new block without ending previous block", "previous", ea.l2BuildingHeader, "prev_payload_id", ea.payloadID)
	}

	parentHeader := ea.sugar.getHeaderByHash(parent)
	if parentHeader == nil {
		return fmt.Errorf("uknown parent block: %s", parent)
	}
	statedb, err := state.New(parentHeader.Root, state.NewDatabase(ea.l2Database), nil)
	if err != nil {
		return fmt.Errorf("failed to init state db around block %s (state %s): %w", parent, parentHeader.Root, err)
	}

	header := &types.Header{
		ParentHash: parent,
		Coinbase:   params.SuggestedFeeRecipient,
		Difficulty: common.Big0,
		Number:     new(big.Int).Add(parentHeader.Number, common.Big1),
		GasLimit:   uint64(*params.GasLimit),
		Time:       uint64(params.Timestamp),
		Extra:      nil,
		MixDigest:  common.Hash(params.PrevRandao),
	}

	header.BaseFee = misc.CalcBaseFee(ea.l2Cfg.Config, parentHeader)

	ea.l2BuildingHeader = header
	ea.l2BuildingState = statedb
	ea.l2Receipts = make([]*types.Receipt, 0)
	ea.l2Transactions = make([]*types.Transaction, 0)
	ea.pendingIndices = make(map[common.Address]uint64)
	ea.l2ForceEmpty = params.NoTxPool
	ea.l2GasPool = new(core.GasPool).AddGas(header.GasLimit)
	ea.payloadID = computePayloadId(parent, params)

	// pre-process the deposits
	for i, otx := range params.Transactions {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(otx); err != nil {
			return fmt.Errorf("transaction %d is not valid: %w", i, err)
		}
		ea.l2BuildingState.Prepare(tx.Hash(), i)
		receipt, err := core.ApplyTransaction(ea.l2Cfg.Config, ea.chainCtx, &ea.l2BuildingHeader.Coinbase,
			ea.l2GasPool, ea.l2BuildingState, ea.l2BuildingHeader, &tx, &ea.l2BuildingHeader.GasUsed, ea.vmCfg)
		if err != nil {
			ea.l2TxFailed = append(ea.l2TxFailed, &tx)
			return fmt.Errorf("failed to apply deposit transaction to L2 block (tx %d): %w", i, err)
		}
		ea.l2Receipts = append(ea.l2Receipts, receipt)
		ea.l2Transactions = append(ea.l2Transactions, &tx)
	}
	return nil
}

func (ea *EngineAPI) endBlock() (*types.Block, error) {
	if ea.l2BuildingHeader == nil {
		return nil, fmt.Errorf("no block is being built currently (id %s)", ea.payloadID)
	}
	header := ea.l2BuildingHeader
	ea.l2BuildingHeader = nil

	header.GasUsed = header.GasLimit - uint64(*ea.l2GasPool)
	header.Root = ea.l2BuildingState.IntermediateRoot(ea.l2Cfg.Config.IsEIP158(header.Number))
	block := types.NewBlock(header, ea.l2Transactions, nil, ea.l2Receipts, trie.NewStackTrie(nil))

	// Write state changes to db
	root, err := ea.l2BuildingState.Commit(ea.l2Cfg.Config.IsEIP158(header.Number))
	if err != nil {
		return nil, fmt.Errorf("l2 state write error: %w", err)
	}
	if err := ea.l2BuildingState.Database().TrieDB().Commit(root, false, nil); err != nil {
		return nil, fmt.Errorf("l2 trie write error: %w", err)
	}
	return block, nil
}

func (ea *EngineAPI) GetPayload(ctx context.Context, payloadId eth.PayloadID) (*eth.ExecutionPayload, error) {
	ea.log.Trace("L2Engine API request received", "method", "GetPayload", "id", payloadId)
	if ea.payloadID != payloadId {
		ea.log.Warn("unexpected payload ID requested for block building", "expected", ea.payloadID, "got", payloadId)
		return nil, beacon.UnknownPayload
	}
	bl, err := ea.endBlock()
	if err != nil {
		ea.log.Error("failed to finish block building", "err", err)
		return nil, beacon.UnknownPayload
	}
	return eth.BlockAsPayload(bl)
}

func (ea *EngineAPI) ForkchoiceUpdate(ctx context.Context, state *eth.ForkchoiceState, attr *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	ea.log.Trace("L2Engine API request received", "method", "ForkchoiceUpdated", "head", state.HeadBlockHash, "finalized", state.FinalizedBlockHash, "safe", state.SafeBlockHash)
	if state.HeadBlockHash == (common.Hash{}) {
		ea.log.Warn("Forkchoice requested update to zero hash")
		return STATUS_INVALID, nil
	}
	// Check whether we have the block yet in our database or not. If not, we'll
	// need to either trigger a sync, or to reject this forkchoice update for a
	// reason.
	block := ea.sugar.getBlockInfoByHash(state.HeadBlockHash)
	if block == nil {
		// TODO: syncing not supported yet
		return STATUS_SYNCING, nil
	}
	valid := func(id *beacon.PayloadID) *eth.ForkchoiceUpdatedResult {
		return &eth.ForkchoiceUpdatedResult{
			PayloadStatus: eth.PayloadStatusV1{Status: eth.ExecutionValid, LatestValidHash: &state.HeadBlockHash},
			PayloadID:     id,
		}
	}
	if ea.sugar.getBlockHashByNumber(block.NumberU64()) != state.HeadBlockHash {
		return nil, fmt.Errorf("cannot reorg! from %s to %s", block.Hash(), state.HeadBlockHash)
	} else if ea.sugar.CurrentBlock().Hash() == state.HeadBlockHash {
		// If the specified head matches with our local head, do nothing and keep
		// generating the payload. It's a special corner case that a few slots are
		// missing and we are requested to generate the payload in slot.
	} else if ea.l2Cfg.Config.Optimism == nil { // minor L2Engine API divergence: allow proposers to reorg their own chain
		panic("engine not configured as optimism engine")
	}

	// If the beacon client also advertised a finalized block, mark the local
	// chain final and completely in PoS mode.
	if state.FinalizedBlockHash != (common.Hash{}) {
		// If the finalized block is not in our canonical tree, somethings wrong
		finalBlock, err := ea.sugar.L2BlockRefByHash(context.Background(), state.FinalizedBlockHash)
		if err != nil {
			ea.log.Warn("Final block not available in database", "hash", state.FinalizedBlockHash)
			return STATUS_INVALID, beacon.InvalidForkChoiceState.With(errors.New("final block not available in database"))
		} else if ea.sugar.getBlockHashByNumber(finalBlock.Number) != state.FinalizedBlockHash {
			ea.log.Warn("Final block not in canonical chain", "number", block.NumberU64(), "hash", state.HeadBlockHash)
			return STATUS_INVALID, beacon.InvalidForkChoiceState.With(errors.New("final block not in canonical chain"))
		}
		// Set the finalized block
		ea.setFinalized(finalBlock.ID())
	}
	// Check if the safe block hash is in our canonical tree, if not somethings wrong
	if state.SafeBlockHash != (common.Hash{}) {
		safeBlock := ea.sugar.getBlockInfoByHash(state.SafeBlockHash)
		if safeBlock == nil {
			ea.log.Warn("Safe block not available in database")
			return STATUS_INVALID, beacon.InvalidForkChoiceState.With(errors.New("safe block not available in database"))
		}
		if ea.sugar.getBlockHashByNumber(safeBlock.NumberU64()) != state.SafeBlockHash {
			ea.log.Warn("Safe block not in canonical chain")
			return STATUS_INVALID, beacon.InvalidForkChoiceState.With(errors.New("safe block not in canonical chain"))
		}
		// Set the safe block
		ea.setSafe(safeBlock.Hash())
	}
	// If payload generation was requested, create a new block to be potentially
	// sealed by the beacon client. The payload will be requested later, and we
	// might replace it arbitrarily many times in between.
	if attr != nil {
		err := ea.startBlock(state.HeadBlockHash, attr)
		if err != nil {
			ea.log.Error("Failed to start block building", "err", err, "noTxPool", attr.NoTxPool, "txs", len(attr.Transactions), "timestamp", attr.Timestamp)
			return STATUS_INVALID, beacon.InvalidPayloadAttributes.With(err)
		}

		return valid(&ea.payloadID), nil
	}
	return valid(nil), nil
}

func (ea *EngineAPI) NewPayload(ctx context.Context, payload *eth.ExecutionPayload) (*eth.PayloadStatusV1, error) {
	ea.log.Trace("L2Engine API request received", "method", "ExecutePayload", "number", payload.BlockNumber, "hash", payload.BlockHash)
	txs := make([][]byte, len(payload.Transactions))
	for i, tx := range payload.Transactions {
		txs[i] = tx
	}
	block, err := beacon.ExecutableDataToBlock(beacon.ExecutableDataV1{
		ParentHash:    payload.ParentHash,
		FeeRecipient:  payload.FeeRecipient,
		StateRoot:     common.Hash(payload.StateRoot),
		ReceiptsRoot:  common.Hash(payload.ReceiptsRoot),
		LogsBloom:     payload.LogsBloom[:],
		Random:        common.Hash(payload.PrevRandao),
		Number:        uint64(payload.BlockNumber),
		GasLimit:      uint64(payload.GasLimit),
		GasUsed:       uint64(payload.GasUsed),
		Timestamp:     uint64(payload.Timestamp),
		ExtraData:     payload.ExtraData,
		BaseFeePerGas: payload.BaseFeePerGas.ToBig(),
		BlockHash:     payload.BlockHash,
		Transactions:  txs,
	})
	if err != nil {
		log.Debug("Invalid NewPayload params", "params", payload, "error", err)
		return &eth.PayloadStatusV1{Status: eth.ExecutionInvalidBlockHash}, nil
	}
	// If we already have the block locally, ignore the entire execution and just
	// return a fake success.
	if block := ea.sugar.getBlockInfoByHash(payload.BlockHash); block != nil {
		ea.log.Warn("Ignoring already known beacon payload", "number", payload.BlockNumber, "hash", payload.BlockHash, "age", common.PrettyAge(time.Unix(int64(block.Time()), 0)))
		hash := block.Hash()
		return &eth.PayloadStatusV1{Status: eth.ExecutionValid, LatestValidHash: &hash}, nil
	}

	// TODO: skipping invalid ancestor check (i.e. not remembering previously failed blocks)

	parent := ea.sugar.getBlockInfoByHash(block.ParentHash())
	if parent == nil {
		// TODO: hack, saying we accepted if we don't know the parent block. Might want to return critical error if we can't actually sync.
		return &eth.PayloadStatusV1{Status: eth.ExecutionAccepted, LatestValidHash: nil}, nil
	}
	ea.sugar.SetHead(eth.HeaderBlockInfo(block.Header()))
	hash := block.Hash()
	return &eth.PayloadStatusV1{Status: eth.ExecutionValid, LatestValidHash: &hash}, nil
}

func (ea *EngineAPI) invalid(err error, latestValid *types.Header) *eth.PayloadStatusV1 {
	currentHash := ea.sugar.CurrentBlock().Hash()
	if latestValid != nil {
		// Set latest valid hash to 0x0 if parent is PoW block
		currentHash = common.Hash{}
		if latestValid.Difficulty.BitLen() == 0 {
			// Otherwise set latest valid hash to parent hash
			currentHash = latestValid.Hash()
		}
	}
	errorMsg := err.Error()
	return &eth.PayloadStatusV1{Status: eth.ExecutionInvalid, LatestValidHash: &currentHash, ValidationError: &errorMsg}
}
