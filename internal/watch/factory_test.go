// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package watch

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/cache"
)

func TestGVRWatcherSeq(t *testing.T) {
	f := NewFactory(nil)

	gvr := client.NewGVR("v1/pods")

	assert.True(t, f.HasChanged(gvr), "unknown GVR should report changed")

	w := &gvrWatcher{handles: make(map[string]handlerEntry)}
	f.watchers.Add(gvr.String(), w)

	assert.False(t, f.HasChanged(gvr), "fresh watcher with seq=0 should not be changed")

	w.seq.Add(1)
	assert.True(t, f.HasChanged(gvr), "seq>0 should report changed")

	f.ResetChanged(gvr)
	assert.False(t, f.HasChanged(gvr), "after reset should not be changed")
}

func TestGVRWatcherIsolation(t *testing.T) {
	f := NewFactory(nil)

	pods := client.NewGVR("v1/pods")
	nodes := client.NewGVR("v1/nodes")

	pw := &gvrWatcher{handles: make(map[string]handlerEntry)}
	nw := &gvrWatcher{handles: make(map[string]handlerEntry)}
	f.watchers.Add(pods.String(), pw)
	f.watchers.Add(nodes.String(), nw)

	pw.seq.Add(1)

	assert.True(t, f.HasChanged(pods))
	assert.False(t, f.HasChanged(nodes), "node watcher should be unaffected by pod events")
}

type fakeRegistration struct{}

func (fakeRegistration) HasSynced() bool { return true }

type fakeInformer struct {
	cache.SharedInformer
	removed int
}

func (f *fakeInformer) RemoveEventHandler(_ cache.ResourceEventHandlerRegistration) error {
	f.removed++
	return nil
}

func TestTerminateCleansUpHandlers(t *testing.T) {
	f := NewFactory(nil)
	f.stopChan = make(chan struct{})

	inf := &fakeInformer{}
	w := &gvrWatcher{
		handles: map[string]handlerEntry{
			"default":     {reg: fakeRegistration{}, inf: inf},
			"kube-system": {reg: fakeRegistration{}, inf: inf},
		},
	}
	f.watchers.Add("v1/pods", w)

	f.Terminate()

	assert.Equal(t, 2, inf.removed, "should have removed both handlers")
	assert.Equal(t, 0, f.watchers.Len(), "watchers should be purged")
}

func TestLRUEvictsOldWatchers(t *testing.T) {
	f := NewFactory(nil)

	inf := &fakeInformer{}
	for i := range maxWatchedGVRs + 5 {
		gvr := client.NewGVR("v1/fake" + string(rune('a'+i)))
		w := &gvrWatcher{
			handles: map[string]handlerEntry{
				"default": {reg: fakeRegistration{}, inf: inf},
			},
		}
		f.watchers.Add(gvr.String(), w)
	}

	assert.Equal(t, maxWatchedGVRs, f.watchers.Len(), "LRU should cap at maxWatchedGVRs")
	assert.Equal(t, 5, inf.removed, "evicted watchers should have handlers removed")
}
