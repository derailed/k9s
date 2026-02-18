// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
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

func TestDetailsEditAction(t *testing.T) {
	var called int

	d := NewDetails(NewApp(mock.NewMockConfig(t)), "Secret Decoder", "default/secret", contentYAML, true).
		SetEditFn(func() error {
			called++
			return nil
		})
	require.NoError(t, d.Init(context.Background()))

	a, ok := d.Actions().Get(ui.KeyE)
	require.True(t, ok)
	assert.Equal(t, "Edit", a.Description)
	assert.True(t, a.Opts.Dangerous)
	assert.Nil(t, d.keyboard(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)))
	assert.Equal(t, 1, called)
}

func TestDetailsEditActionCanRefreshContent(t *testing.T) {
	d := NewDetails(NewApp(mock.NewMockConfig(t)), "Secret Decoder", "default/secret", contentYAML, true).
		Update("old: value\n")
	d.SetEditFn(func() error {
		d.Update("new: value\n")
		return nil
	})
	require.NoError(t, d.Init(context.Background()))

	assert.Contains(t, sanitizeEsc(d.text.GetText(true)), "old: value")
	assert.Nil(t, d.keyboard(tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)))
	assert.Contains(t, sanitizeEsc(d.text.GetText(true)), "new: value")
}

func TestDetailsWithoutEditAction(t *testing.T) {
	d := NewDetails(NewApp(mock.NewMockConfig(t)), "Details", "subject", contentYAML, true)
	require.NoError(t, d.Init(context.Background()))

	_, ok := d.Actions().Get(ui.KeyE)
	assert.False(t, ok)
}
