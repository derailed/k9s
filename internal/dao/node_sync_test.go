// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type syncedInformer struct {
	cache.SharedIndexInformer
	synced atomic.Bool
}

func (i *syncedInformer) HasSynced() bool { return i.synced.Load() }

type syncedGenericInformer struct{ inf *syncedInformer }

func (i syncedGenericInformer) Informer() cache.SharedIndexInformer { return i.inf }
func (syncedGenericInformer) Lister() cache.GenericLister           { return nil }

type authConn struct {
	client.Connection
	allow bool
}

func (c authConn) CanI(string, *client.GVR, string, []string) (bool, error) { return c.allow, nil }

type nodesFactory struct {
	Factory
	conn    client.Connection
	inf     *syncedInformer
	nodes   []runtime.Object
	lookups atomic.Int32
}

func (f *nodesFactory) Client() client.Connection { return f.conn }

func (f *nodesFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	f.lookups.Add(1)
	return syncedGenericInformer{inf: f.inf}, nil
}

func (f *nodesFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	if !f.inf.synced.Load() {
		return nil, nil
	}

	return f.nodes, nil
}

func newNodesFactory(allow bool) *nodesFactory {
	node := func(n string) runtime.Object {
		return &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Node",
			"metadata":   map[string]any{"name": n},
		}}
	}

	return &nodesFactory{
		conn:  authConn{Connection: client.NewTestAPIClient(), allow: allow},
		inf:   &syncedInformer{},
		nodes: []runtime.Object{node("n1"), node("n2")},
	}
}

func TestFetchNodesWaitsForInformerSync(t *testing.T) {
	f := newNodesFactory(true)
	time.AfterFunc(50*time.Millisecond, func() { f.inf.synced.Store(true) })

	start := time.Now()
	nn, err := FetchNodes(context.Background(), f, "")
	require.NoError(t, err)
	assert.Len(t, nn.Items, 2, "the first fetch sees the synced nodes")
	assert.Less(t, time.Since(start), time.Second)
}

func TestFetchNodesStopsWaitingOnCancel(t *testing.T) {
	f := newNodesFactory(true)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	nn, err := FetchNodes(ctx, f, "")
	require.NoError(t, err)
	assert.Empty(t, nn.Items)
	assert.Less(t, time.Since(start), 500*time.Millisecond)
}

func TestFetchNodesDenied(t *testing.T) {
	f := newNodesFactory(false)

	_, err := FetchNodes(context.Background(), f, "")
	require.Error(t, err)
	assert.Zero(t, f.lookups.Load(), "no informer when nodes cannot be listed")
}
