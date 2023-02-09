package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const perms = 0644

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
