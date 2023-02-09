package main_test

import (
	"math/rand"
	"op-mordor"
	"testing"

	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/stretchr/testify/require"
)

func TestDiskStore(t *testing.T) {
	rng := rand.New(rand.NewSource(420))

	storePath := t.TempDir()
	s, err := main.NewDiskStore(storePath)
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
	s, err := main.NewDiskStore(storePath)
	require.NoError(t, err)

	//bstore := main.BlockStore{s}
	bsource := main.BlockSource{s}

	rndHash := testutils.RandomHash(rng)

	t.Run("BlockSource/not-exist", func(t *testing.T) {
		b, err := bsource.ReadBlock(rndHash)
		requireNoDataError(t, err)
		require.Nil(t, b)
	})
}

func requireNoDataError(t *testing.T, err error) {
	var noDataErr main.NoDataError
	require.ErrorAs(t, err, &noDataErr)
}
