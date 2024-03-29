package l2

import (
	"context"
	"fmt"
	"op-mordor/oracle"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rpc"
)

func externalDBGet(cl *rpc.Client) func(key []byte) ([]byte, error) {
	return func(key []byte) ([]byte, error) {
		var resp hexutil.Bytes
		err := cl.Call(&resp, "debug_dbGet", hexutil.Bytes(key))
		return resp, err
	}
}

type OracleBackedDB struct {
	// attempt first: load from in-memory db of previously accessed values
	db *memorydb.Database

	oracle oracle.L2StateOracle
}

func NewOracleBackedDB(oracle oracle.L2StateOracle) *OracleBackedDB {
	return &OracleBackedDB{
		db:     memorydb.New(),
		oracle: oracle,
	}
}

func (p *OracleBackedDB) Has(key []byte) (bool, error) {
	panic("not supported")
}

func (p *OracleBackedDB) Get(key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("can only read 32-byte key values, pre-images must be identified by hash")
	}
	v, err := p.db.Get(key)
	if err == nil {
		return v, nil
	}
	if err.Error() == "not found" {
		v, err := p.oracle.FetchL2MPTNode(context.TODO(), *(*[32]byte)(key))
		if err != nil {
			return nil, err
		}
		if err := p.db.Put(key, v); err != nil {
			panic(fmt.Errorf("failed to put value into mem db: %w", err))
		}
		return v, nil
	}
	return nil, err
}

func (p *OracleBackedDB) Put(key []byte, value []byte) error {
	return p.db.Put(key, value)
}

func (p OracleBackedDB) Delete(key []byte) error {
	// we never delete pre-images
	return nil
}

func (p OracleBackedDB) Stat(property string) (string, error) {
	panic("not supported")
}

func (p OracleBackedDB) NewBatch() ethdb.Batch {
	return p.db.NewBatch()
}

func (p OracleBackedDB) NewBatchWithSize(size int) ethdb.Batch {
	return p.db.NewBatchWithSize(size)
}

func (p OracleBackedDB) NewIterator(prefix []byte, start []byte) ethdb.Iterator {
	panic("not supported")
}

func (p OracleBackedDB) Compact(start []byte, limit []byte) error {
	return nil // no-op
}

func (p OracleBackedDB) NewSnapshot() (ethdb.Snapshot, error) {
	panic("not supported")
}

func (p OracleBackedDB) Close() error {
	return nil
}

// We implement the full ethdb.Database bloat because the StateDB takes this full interface,
// even though it only uses the KeyValue subset.

func (p *OracleBackedDB) HasAncient(kind string, number uint64) (bool, error) {
	panic("not supported")
}

func (p *OracleBackedDB) Ancient(kind string, number uint64) ([]byte, error) {
	panic("not supported")
}

func (p *OracleBackedDB) AncientRange(kind string, start, count, maxBytes uint64) ([][]byte, error) {
	panic("not supported")
}

func (p *OracleBackedDB) Ancients() (uint64, error) {
	panic("not supported")
}

func (p *OracleBackedDB) Tail() (uint64, error) {
	panic("not supported")
}

func (p *OracleBackedDB) AncientSize(kind string) (uint64, error) {
	panic("not supported")
}

func (p *OracleBackedDB) ReadAncients(fn func(ethdb.AncientReaderOp) error) (err error) {
	panic("not supported")
}

func (p *OracleBackedDB) ModifyAncients(f func(ethdb.AncientWriteOp) error) (int64, error) {
	panic("not supported")
}

func (p *OracleBackedDB) TruncateHead(n uint64) error {
	panic("not supported")
}

func (p *OracleBackedDB) TruncateTail(n uint64) error {
	panic("not supported")
}

func (p *OracleBackedDB) Sync() error {
	panic("not supported")
}

func (p *OracleBackedDB) MigrateTable(s string, f func([]byte) ([]byte, error)) error {
	panic("not supported")
}

func (p *OracleBackedDB) AncientDatadir() (string, error) {
	panic("not supported")
}

var _ ethdb.KeyValueStore = (*OracleBackedDB)(nil)
