// Copyright (C) 2019-2025, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/logging"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ava-labs/avalanchego/vms/txs/mempool"

	pmempool "github.com/ava-labs/avalanchego/vms/platformvm/txs/mempool"
)

var errFoo = errors.New("foo")

// Add should error if verification errors
func TestGossipMempoolAddVerificationError(t *testing.T) {
	require := require.New(t)

	txID := ids.GenerateTestID()
	tx := &txs.Tx{
		TxID: txID,
	}

	mempool, err := pmempool.New("", prometheus.NewRegistry())
	require.NoError(err)
	txVerifier := testTxVerifier{err: errFoo}

	gossipMempool, err := newGossipMempool(
		mempool,
		prometheus.NewRegistry(),
		logging.NoLog{},
		txVerifier,
		testConfig.ExpectedBloomFilterElements,
		testConfig.ExpectedBloomFilterFalsePositiveProbability,
		testConfig.MaxBloomFilterFalsePositiveProbability,
	)
	require.NoError(err)

	err = gossipMempool.Add(tx)
	require.ErrorIs(err, errFoo)
	require.False(gossipMempool.bloom.Has(tx))
}

// Adding a duplicate to the mempool should return an error
func TestMempoolDuplicate(t *testing.T) {
	require := require.New(t)

	testMempool, err := pmempool.New("", prometheus.NewRegistry())
	require.NoError(err)
	txVerifier := testTxVerifier{}

	txID := ids.GenerateTestID()
	tx := &txs.Tx{
		Unsigned: &txs.BaseTx{},
		TxID:     txID,
	}

	require.NoError(testMempool.Add(tx))
	gossipMempool, err := newGossipMempool(
		testMempool,
		prometheus.NewRegistry(),
		logging.NoLog{},
		txVerifier,
		testConfig.ExpectedBloomFilterElements,
		testConfig.ExpectedBloomFilterFalsePositiveProbability,
		testConfig.MaxBloomFilterFalsePositiveProbability,
	)
	require.NoError(err)

	err = gossipMempool.Add(tx)
	require.ErrorIs(err, mempool.ErrDuplicateTx)
	require.False(gossipMempool.bloom.Has(tx))
}

// Adding a tx to the mempool should add it to the bloom filter
func TestGossipAddBloomFilter(t *testing.T) {
	require := require.New(t)

	txID := ids.GenerateTestID()
	tx := &txs.Tx{
		Unsigned: &txs.BaseTx{},
		TxID:     txID,
	}

	txVerifier := testTxVerifier{}
	mempool, err := pmempool.New("", prometheus.NewRegistry())
	require.NoError(err)

	gossipMempool, err := newGossipMempool(
		mempool,
		prometheus.NewRegistry(),
		logging.NoLog{},
		txVerifier,
		testConfig.ExpectedBloomFilterElements,
		testConfig.ExpectedBloomFilterFalsePositiveProbability,
		testConfig.MaxBloomFilterFalsePositiveProbability,
	)
	require.NoError(err)

	require.NoError(gossipMempool.Add(tx))
	require.True(gossipMempool.bloom.Has(tx))
}
