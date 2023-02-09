package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
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

func (s DiskStore) StoreNode(nodeHash common.Hash, node []byte) error {
	return s.store(nodeHash, func(w io.Writer) error {
		_, err := io.Copy(w, bytes.NewReader(node))
		return err
	})
}

type dataSource func(w io.Writer) error

func (s DiskStore) store(hash common.Hash, source dataSource) error {
	f, err := os.Create(s.fileName(hash))
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	err = source(f)
	if err != nil {
		return fmt.Errorf("writing data: %w", err)
	}
	return nil
}

func (s DiskStore) fileName(hash common.Hash) string {
	return fmt.Sprintf("%s/%s", s.dir, hash.Hex())
}

type NoDataError struct {
	Key common.Hash
}

func (nde NoDataError) Error() string {
	return fmt.Sprintf("no data for key %x", nde.Key)
}

func IsNoDataError(err error) bool {
	var _noDataError NoDataError
	return errors.As(err, &_noDataError)
}

func (s DiskStore) ReadHeader(hash common.Hash) (*types.Header, error) {
	var header types.Header
	if err := s.read(hash, func(f io.Reader) error {
		return rlp.Decode(f, &header)
	}); err != nil {
		return nil, err
	}
	return &header, nil
}

func (s DiskStore) ReadTransactions(txRoot common.Hash) (types.Transactions, error) {
	panic("implement")
}

func (s DiskStore) ReadReceipts(hash common.Hash) (types.Receipts, error) {
	panic("implement")
}

func (s DiskStore) ReadNode(nodeHash common.Hash) (node []byte, err error) {
	err = s.read(nodeHash, func(r io.Reader) error {
		node, err = ioutil.ReadAll(r)
		return err
	})
	return
}

func (s DiskStore) read(hash common.Hash, restore func(io.Reader) error) error {
	f, err := os.Open(s.fileName(hash))
	if errors.Is(err, fs.ErrNotExist) {
		return NoDataError{hash}
	} else if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	return restore(f)
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
