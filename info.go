package main

import (
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

type PreimageBlockInfo struct {
	// Prefixed all fields with "Info" to avoid collisions with the interface method names.

	InfoHash        common.Hash
	InfoParentHash  common.Hash
	InfoCoinbase    common.Address
	InfoRoot        common.Hash
	InfoNum         uint64
	InfoTime        uint64
	InfoMixDigest   [32]byte
	InfoBaseFee     *big.Int
	InfoReceiptRoot common.Hash
}

func (l *PreimageBlockInfo) Hash() common.Hash {
	return l.InfoHash
}

func (l *PreimageBlockInfo) ParentHash() common.Hash {
	return l.InfoParentHash
}

func (l *PreimageBlockInfo) Coinbase() common.Address {
	return l.InfoCoinbase
}

func (l *PreimageBlockInfo) Root() common.Hash {
	return l.InfoRoot
}

func (l *PreimageBlockInfo) NumberU64() uint64 {
	return l.InfoNum
}

func (l *PreimageBlockInfo) Time() uint64 {
	return l.InfoTime
}

func (l *PreimageBlockInfo) MixDigest() common.Hash {
	return l.InfoMixDigest
}

func (l *PreimageBlockInfo) BaseFee() *big.Int {
	return l.InfoBaseFee
}

func (l *PreimageBlockInfo) ReceiptHash() common.Hash {
	return l.InfoReceiptRoot
}

func (l *PreimageBlockInfo) ID() eth.BlockID {
	return eth.BlockID{Hash: l.InfoHash, Number: l.InfoNum}
}

func (l *PreimageBlockInfo) BlockRef() eth.L1BlockRef {
	return eth.L1BlockRef{
		Hash:       l.InfoHash,
		Number:     l.InfoNum,
		ParentHash: l.InfoParentHash,
		Time:       l.InfoTime,
	}
}
