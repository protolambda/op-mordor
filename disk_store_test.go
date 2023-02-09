package main_test

import (
	"math/rand"
	"op-mordor"
	"testing"

	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/stretchr/testify/require"
)

func TestDiskStore(t *testing.T) {
	require := require.New(t)
	rng := rand.New(rand.NewSource(420))

	storePath := t.TempDir()
	s, err := main.NewDiskStore(storePath)
	require.NoError(err)

	rndHash := testutils.RandomHash(rng)
	var noDataErr main.NoDataError

	t.Run("ReadHeader/not-exist", func(t *testing.T) {
		h, err := s.ReadHeader(rndHash)
		require.ErrorAs(err, &noDataErr)
		require.Nil(h)
	})

	t.Run("ReadNode/not-exist", func(t *testing.T) {
		node, err := s.ReadNode(rndHash)
		require.ErrorAs(err, &noDataErr)
		require.Nil(node)
	})

	t.Run("Store+ReadHeader", func(t *testing.T) {
		rndHeader := testutils.RandomHeader(rng)
		require.NoError(s.StoreHeader(rndHash, rndHeader))

		header, err := s.ReadHeader(rndHash)
		require.NoError(err)
		require.Equal(header, rndHeader)
	})

	t.Run("Store+ReadNode", func(t *testing.T) {
		rndNode := testutils.RandomData(rng, 420)
		require.NoError(s.StoreNode(rndHash, rndNode))

		node, err := s.ReadNode(rndHash)
		require.NoError(err)
		require.Equal(node, rndNode)
	})
}
