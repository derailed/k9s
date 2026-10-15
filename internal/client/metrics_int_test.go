// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchListFreshCacheHit(t *testing.T) {
	m := NewMetricsServer(nil)
	m.cache.Add("nodes", "cached", mxCacheExpiry)

	v, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
		t.Fatal("list should not be called on a cache hit")
		return nil, nil
	})

	require.NoError(t, err)
	assert.Equal(t, "cached", v)
}

func TestFetchListColdWaitsForFastFetch(t *testing.T) {
	m := NewMetricsServer(nil)

	v, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
		return "fresh", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "fresh", v)
}

func TestFetchListSlowFetchDoesNotBlock(t *testing.T) {
	m := NewMetricsServer(nil)
	release := make(chan struct{})

	start := time.Now()
	v, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
		<-release
		return "slow", nil
	})
	elapsed := time.Since(start)

	assert.Nil(t, v)
	assert.ErrorIs(t, err, ErrMetricsNotReady)
	assert.Less(t, elapsed, 5*time.Second)

	// Once the background fetch lands, callers get the data.
	close(release)
	assert.Eventually(t, func() bool {
		v, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
			return "slow", nil
		})
		return err == nil && v == "slow"
	}, 2*time.Second, 10*time.Millisecond)
}

func TestFetchListServesStaleWhileRefreshing(t *testing.T) {
	m := NewMetricsServer(nil)
	m.stale["nodes"] = "stale"
	release := make(chan struct{})
	defer close(release)

	v, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
		<-release
		return "fresh", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "stale", v)
}

func TestFetchListSingleFlight(t *testing.T) {
	m := NewMetricsServer(nil)
	release := make(chan struct{})
	defer close(release)
	var calls atomic.Int32

	for range 3 {
		_, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
			calls.Add(1)
			<-release
			return "fresh", nil
		})
		assert.ErrorIs(t, err, ErrMetricsNotReady)
	}

	assert.Equal(t, int32(1), calls.Load())
}

func TestFetchListErrorRetriesNextCall(t *testing.T) {
	m := NewMetricsServer(nil)

	_, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
		return nil, errors.New("boom")
	})
	assert.ErrorIs(t, err, ErrMetricsNotReady)

	v, err := m.fetchList(context.Background(), "nodes", func(context.Context) (any, error) {
		return "recovered", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "recovered", v)
}
