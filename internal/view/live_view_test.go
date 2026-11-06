// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLiveViewSetText(t *testing.T) {
	s := `
apiVersion: v1
  data:
    the secret name you want to quote to use tls.","title":"secretName","type":"string"}},"required":["http","class","classInSpec"],"type":"object"}
`

	v := NewLiveView(NewApp(mock.NewMockConfig(t)), "fred", nil)
	require.NoError(t, v.Init(context.Background()))
	v.text.SetText(colorizeYAML(config.Yaml{}, s))

	assert.Equal(t, s, sanitizeEsc(v.text.GetText(true)))
}

type pathModel struct {
	gvr  *client.GVR
	path string
}

func (m pathModel) GetPath() string  { return m.path }
func (pathModel) Filter(string)      {}
func (m pathModel) GVR() *client.GVR { return m.gvr }
func (pathModel) ClearFilter()       {}
func (pathModel) Peek() []string     { return nil }
func (pathModel) SetOptions(context.Context, model.ViewerToggleOpts) {
}
func (pathModel) Watch(context.Context) error                 { return nil }
func (pathModel) Refresh(context.Context) error               { return nil }
func (pathModel) AddListener(model.ResourceViewerListener)    {}
func (pathModel) RemoveListener(model.ResourceViewerListener) {}

func TestLiveViewSecretEditKeys(t *testing.T) {
	v := NewLiveView(NewApp(mock.NewMockConfig(t)), "YAML", pathModel{gvr: client.SecGVR, path: "default/s"})
	require.NoError(t, v.Init(context.Background()))

	decoded, ok := v.Actions().Get(ui.KeyE)
	require.True(t, ok)
	assert.Equal(t, "Edit Decoded", decoded.Description)
	raw, ok := v.Actions().Get(ui.KeyShiftE)
	require.True(t, ok)
	assert.Equal(t, "Edit Raw", raw.Description)
}

func TestLiveViewSecretEditKeysReadOnly(t *testing.T) {
	cfg := mock.NewMockConfig(t)
	cfg.K9s.ReadOnly = true
	v := NewLiveView(NewApp(cfg), "YAML", pathModel{gvr: client.SecGVR, path: "default/s"})
	require.NoError(t, v.Init(context.Background()))

	_, ok := v.Actions().Get(ui.KeyE)
	assert.False(t, ok)
	_, ok = v.Actions().Get(ui.KeyShiftE)
	assert.False(t, ok)
}

func TestLiveViewPodEditKey(t *testing.T) {
	v := NewLiveView(NewApp(mock.NewMockConfig(t)), "YAML", pathModel{gvr: client.PodGVR, path: "default/p"})
	require.NoError(t, v.Init(context.Background()))

	edit, ok := v.Actions().Get(ui.KeyE)
	require.True(t, ok)
	assert.Equal(t, "Edit", edit.Description)
	_, ok = v.Actions().Get(ui.KeyShiftE)
	assert.False(t, ok)
}
