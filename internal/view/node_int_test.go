// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type refreshCountingModel struct {
	mockTableModel
	refreshes atomic.Int32
}

func (m *refreshCountingModel) Refresh(context.Context) error {
	m.refreshes.Add(1)
	return nil
}

func TestNodeShouldCountPods(t *testing.T) {
	uu := map[string]struct {
		disablePodCounting bool
		wide               bool
		vs                 *config.ViewSetting
		expected           bool
	}{
		"default-narrow": {
			expected: false,
		},
		"default-wide": {
			wide:     true,
			expected: true,
		},
		"disabled-in-wide": {
			disablePodCounting: true,
			wide:               true,
			expected:           false,
		},
		"custom-view-show-in-narrow": {
			vs: &config.ViewSetting{
				Columns: []string{"NAME", "PODS|S", "AGE"},
			},
			expected: true,
		},
		"custom-view-hidden-in-wide": {
			wide: true,
			vs: &config.ViewSetting{
				Columns: []string{"NAME", "PODS|H", "AGE"},
			},
			expected: false,
		},
		"custom-view-omitted-in-narrow": {
			vs: &config.ViewSetting{
				Columns: []string{"NAME", "STATUS", "AGE"},
			},
			expected: false,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			nv, ok := NewNode(client.NodeGVR).(*Node)
			require.True(t, ok)
			require.NoError(t, nv.GetTable().Init(makeContext(t)))
			nv.App().Config.K9s.DisablePodCounting = u.disablePodCounting
			if u.vs != nil {
				nv.GetTable().SetViewSetting(u.vs)
			}
			if u.wide {
				nv.GetTable().ToggleWide()
			}

			ctx := nv.nodeContext(context.Background())
			fn, ok := ctx.Value(internal.KeyPodCounting).(func() bool)
			require.True(t, ok)
			assert.Equal(t, u.expected, fn())
		})
	}
}

func TestNodeToggleWideTriggersRefreshLazily(t *testing.T) {
	nv, ok := NewNode(client.NodeGVR).(*Node)
	require.True(t, ok)
	require.NoError(t, nv.GetTable().Init(makeContext(t)))

	m := &refreshCountingModel{}
	nv.GetTable().SetModel(m)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nv.GetTable().SetContext(nv.nodeContext(ctx))

	// Narrow -> Wide enables pod counting and triggers a background model refresh.
	nv.ToggleWide()
	assert.True(t, nv.GetTable().IsWide())
	require.Eventually(t, func() bool {
		return m.refreshes.Load() == 1
	}, time.Second, 10*time.Millisecond)

	// Wide -> Narrow disables pod counting without triggering an extra API refresh.
	nv.ToggleWide()
	assert.False(t, nv.GetTable().IsWide())
	assert.Equal(t, int32(1), m.refreshes.Load())
}
