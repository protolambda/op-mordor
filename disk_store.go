package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

type DiskStore struct {
	dir string
}

func NewDiskStore(dir string) (*DiskStore, error) {
	err := os.Mkdir(dir, 0777)
	if err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &DiskStore{dir: dir}, nil
}

func (s DiskStore) StoreHeader(hash common.Hash, header *types.Header) error {
	return s.store(hash, func(w io.Writer) error {
		return header.EncodeRLP(w)
	})
}

func (s DiskStore) StoreTransactions(txRoot common.Hash, txs types.Transactions) error {
	pkw := keyValueWriter{s: s}
	hasher := &noResetTrie{*trie.NewStackTrie(pkw)}

	testTxHash := types.DeriveSha(txs, hasher)
	if testTxHash != txRoot {
		return fmt.Errorf("expected txRoot %s does not match actual root %s", txRoot, testTxHash)
	}
	_, err := hasher.Commit()
	if err != nil {
		return fmt.Errorf("store tx: %w", err)
	}
	return nil
}

func (s DiskStore) StoreReceipts(receipts types.Receipts) error {
	pkw := keyValueWriter{s: s}
	hasher := &noResetTrie{*trie.NewStackTrie(pkw)}

	types.DeriveSha(receipts, hasher)
	_, err := hasher.Commit()
	if err != nil {
		return fmt.Errorf("store receipts: %w", err)
	}
	return nil
}

type dataSource func(w io.Writer) error

func (s DiskStore) store(hash common.Hash, source dataSource) error {
	f, err := os.Create(s.fileName(hash))
	if err != nil {
		return fmt.Errorf("store data: %w", err)
	}
	defer f.Close()
	err = source(f)
	if err != nil {
		return fmt.Errorf("store data: %w", err)
	}
	return nil
}

func (s DiskStore) fileName(hash common.Hash) string {
	return fmt.Sprintf("%s/%s", s.dir, hash.Hex())
}

type keyValueWriter struct {
	s DiskStore
}

func (k keyValueWriter) Put(key []byte, value []byte) error {
	return k.s.store(common.BytesToHash(key), func(w io.Writer) error {
		_, err := w.Write(value)
		return err
	})
}

func (k keyValueWriter) Delete(key []byte) error {
	return errors.New("delete not supported")
}

type noResetTrie struct {
	trie.StackTrie
}

func (t *noResetTrie) Reset() {
}
