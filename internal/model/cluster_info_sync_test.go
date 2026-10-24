// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type gateConn struct {
	client.Connection
	entered chan struct{}
	release chan struct{}
}

func (c *gateConn) ConnectionOK() bool {
	select {
	case c.entered <- struct{}{}:
	default:
	}
	<-c.release

	return false
}

type connFactory struct {
	testFactory
	conn client.Connection
}

func (f connFactory) Client() client.Connection { return f.conn }

type infoRecorder struct {
	events atomic.Int32
}

func (r *infoRecorder) ClusterInfoChanged(_, _ *ClusterMeta) { r.events.Add(1) }
func (r *infoRecorder) ClusterInfoUpdated(*ClusterMeta)      { r.events.Add(1) }

func TestClusterInfoConcurrentResetAndRefresh(_ *testing.T) {
	f := makeFactory()
	ci := NewClusterInfo(f, "v1", &config.K9s{SkipLatestRevCheck: true})

	var wg sync.WaitGroup
	wg.Add(3)
	for range 2 {
		go func() {
			defer wg.Done()
			for range 50 {
				ci.Reset(f)
			}
		}()
	}
	go func() {
		defer wg.Done()
		for range 50 {
			ci.Refresh()
		}
	}()
	wg.Wait()
}

func TestClusterInfoDropsRefreshThatPredatesReset(t *testing.T) {
	conn := &gateConn{
		Connection: client.NewTestAPIClient(),
		entered:    make(chan struct{}, 1),
		release:    make(chan struct{}),
	}
	f := connFactory{conn: conn}
	ci := NewClusterInfo(f, "v0.51.0", &config.K9s{SkipLatestRevCheck: true})
	var l infoRecorder
	ci.AddListener(&l)

	refreshed := make(chan struct{})
	go func() {
		defer close(refreshed)
		ci.Refresh()
	}()
	<-conn.entered

	reset := make(chan struct{})
	go func() {
		defer close(reset)
		ci.Reset(f)
	}()
	require.Eventually(t, func() bool {
		ci.mx.RLock()
		defer ci.mx.RUnlock()
		return ci.gen == 1
	}, time.Second, time.Millisecond)

	close(conn.release)
	<-refreshed
	<-reset

	assert.Equal(t, int32(1), l.events.Load(), "only the refresh that follows the reset notifies")
}
