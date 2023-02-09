package main

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Store interface {
	StoreHeader(hash common.Hash, header *types.Header) error

	StoreTransactions(txRoot common.Hash, transactions types.Transactions) error

	StoreReceipts(receipts types.Receipts) error

	StoreNode(nodeHash common.Hash, node []byte) error
}

type Source interface {
	ReadHeader(hash common.Hash) (*types.Header, error)

	ReadTransactions(txRoot common.Hash) (types.Transactions, error)

	ReadReceipts(hash common.Hash) (types.Receipts, error)

	ReadNode(nodeHash common.Hash) (node []byte, err error)
}
