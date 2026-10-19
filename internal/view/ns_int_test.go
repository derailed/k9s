// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamespaceDecorateKeepsAge(t *testing.T) {
	cfg := mock.NewMockConfig(t)
	cfg.SetConnection(mock.NewMockConnection())
	app := NewApp(cfg)
	ctx := context.WithValue(context.Background(), internal.KeyApp, app)
	ctx = context.WithValue(ctx, internal.KeyStyles, app.Styles)

	rv := NewNamespace(client.NsGVR)
	require.NoError(t, rv.Init(ctx))
	ns, ok := rv.(*Namespace)
	require.True(t, ok)
	require.NotNil(t, ns.App().Conn())

	header := model1.Header{
		{Name: "NAME"},
		{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
	re := model1.NewRowEvents(1)
	re.Add(&model1.RowEvent{Row: model1.Row{
		ID:     "default",
		Fields: model1.Fields{"default", "1m"},
		Age:    time.Now().Add(-time.Minute),
	}})
	td := model1.NewTableDataWithRows(client.NsGVR, header, re)

	ns.decorate(td)

	assert.Equal(t, 2, td.RowCount())
	td.RowsRange(func(_ int, e model1.RowEvent) bool {
		assert.False(t, e.Row.Age.IsZero(), "row %q lost its Age", e.Row.ID)
		return true
	})
}
