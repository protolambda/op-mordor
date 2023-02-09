package store

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BlockStore augments a Store with a method StoreBlock to store blocks.
// It reuses header and transactions storing of the Store.
type BlockStore struct {
	Store
}

func (s BlockStore) StoreBlock(block *types.Block) error {
	if err := s.StoreHeader(block.Hash(), block.Header()); err != nil {
		return err
	}
	return s.StoreTransactions(block.TxHash(), block.Transactions())
}

// BlockStore augments a Store with a method StoreBlock to store blocks.
// It reuses header and transactions storing of the Store.
type BlockSource struct {
	Source
}

func (s BlockSource) ReadBlock(hash common.Hash) (*types.Block, error) {
	h, err := s.ReadHeader(hash)
	if err != nil {
		return nil, fmt.Errorf("reading header: %w", err)
	}

	txs, err := s.ReadTransactions(h.TxHash)
	if err != nil {
		return nil, fmt.Errorf("reading txs: %w", err)
	}

	return types.NewBlockWithHeader(h).WithBody(txs, nil), nil
}
