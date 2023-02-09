package store_test

import (
	"math/rand"
	"op-mordor/store"
	"testing"

	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/stretchr/testify/require"
)

func TestDiskStore(t *testing.T) {
	rng := rand.New(rand.NewSource(420))

	storePath := t.TempDir()
	s, err := store.NewDiskStore(storePath)
	require.NoError(t, err)

	rndHash := testutils.RandomHash(rng)

	t.Run("ReadHeader/not-exist", func(t *testing.T) {
		h, err := s.ReadHeader(rndHash)
		requireNoDataError(t, err)
		require.Nil(t, h)
	})

	t.Run("ReadNode/not-exist", func(t *testing.T) {
		node, err := s.ReadNode(rndHash)
		requireNoDataError(t, err)
		require.Nil(t, node)
	})

	t.Run("Store+ReadHeader", func(t *testing.T) {
		rndHeader := testutils.RandomHeader(rng)
		require.NoError(t, s.StoreHeader(rndHash, rndHeader))

		header, err := s.ReadHeader(rndHash)
		require.NoError(t, err)
		require.Equal(t, header, rndHeader)
	})

	t.Run("Store+ReadNode", func(t *testing.T) {
		rndNode := testutils.RandomData(rng, 420)
		require.NoError(t, s.StoreNode(rndHash, rndNode))

		node, err := s.ReadNode(rndHash)
		require.NoError(t, err)
		require.Equal(t, node, rndNode)
	})
}

func TestBlockStoreSource(t *testing.T) {
	rng := rand.New(rand.NewSource(420))

	storePath := t.TempDir()
	s, err := store.NewDiskStore(storePath)
	require.NoError(t, err)

	bstore := store.BlockStore{s}
	bsource := store.BlockSource{s}

	rndHash := testutils.RandomHash(rng)

	t.Run("BlockSource/not-exist", func(t *testing.T) {
		b, err := bsource.ReadBlock(rndHash)
		requireNoDataError(t, err)
		require.Nil(t, b)
	})

	t.Run("Store+ReadBlock", func(t *testing.T) {
		rndBlock, _ := testutils.RandomBlock(rng, 16)
		require.NoError(t, bstore.StoreBlock(rndBlock))

		// TODO: enable after ReadTransactions is implemented
		//block, err := bsource.ReadBlock(rndBlock.Hash())
		//require.NoError(t, err)
		//require.Equal(t, block, rndBlock)
	})
}

func requireNoDataError(t *testing.T, err error) {
	var noDataErr store.NoDataError
	require.ErrorAs(t, err, &noDataErr)
}
