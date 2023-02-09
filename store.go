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
