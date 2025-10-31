// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAliasNew(t *testing.T) {
	v := view.NewAlias(client.AliGVR)

	require.NoError(t, v.Init(makeContext(t)))
	assert.Equal(t, "Aliases", v.Name())
	assert.Len(t, v.Hints(), 7)
}

func TestAliasSearch(t *testing.T) {
	v := view.NewAlias(client.AliGVR)
	require.NoError(t, v.Init(makeContext(t)))
	v.GetTable().SetModel(new(mockModel))
	v.GetTable().Refresh()
	v.App().Prompt().SetModel(v.GetTable().CmdBuff())
	v.App().Prompt().SendStrokes("blee")

	assert.Equal(t, 3, v.GetTable().GetColumnCount())
	assert.Equal(t, 3, v.GetTable().GetRowCount())
}

func TestAliasGoto(t *testing.T) {
	v := view.NewAlias(client.AliGVR)
	require.NoError(t, v.Init(makeContext(t)))
	v.GetTable().Select(0, 0)

	b := buffL{}
	v.GetTable().CmdBuff().SetActive(true)
	v.GetTable().CmdBuff().AddListener(&b)
	v.GetTable().SendKey(tcell.NewEventKey(tcell.KeyEnter, 256, tcell.ModNone))

	assert.True(t, v.GetTable().CmdBuff().IsActive())
}

// ----------------------------------------------------------------------------
// Helpers...

type buffL struct {
	active  int
	changed int
}

func (b *buffL) BufferChanged(string, string) {
	b.changed++
}
func (*buffL) BufferCompleted(string, string) {}

func (b *buffL) BufferActive(bool, model.BufferKind) {
	b.active++
}

func makeContext(t testing.TB) context.Context {
	a := view.NewApp(mock.NewMockConfig(t))
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

type mockModel struct{}

var (
	_ ui.Tabular   = (*mockModel)(nil)
	_ ui.Suggester = (*mockModel)(nil)
)

func (*mockModel) SetViewSetting(context.Context, *config.ViewSetting) {}
func (*mockModel) CurrentSuggestion() (string, bool)                   { return "", false }
func (*mockModel) NextSuggestion() (string, bool)                      { return "", false }
func (*mockModel) PrevSuggestion() (string, bool)                      { return "", false }
func (*mockModel) ClearSuggestions()                                   {}
func (*mockModel) SetInstance(string)                                  {}
func (*mockModel) SetLabelSelector(labels.Selector)                    {}
func (*mockModel) GetLabelSelector() labels.Selector                   { return nil }
func (*mockModel) Empty() bool                                         { return false }
func (*mockModel) RowCount() int                                       { return 1 }
func (*mockModel) HasMetrics() bool                                    { return true }
func (*mockModel) Peek() *model1.TableData                             { return makeTableData() }
func (*mockModel) ClusterWide() bool                                   { return false }
func (*mockModel) GetNamespace() string                                { return "blee" }
func (*mockModel) SetNamespace(string)                                 {}
func (*mockModel) ToggleToast()                                        {}
func (*mockModel) AddListener(model.TableListener)                     {}
func (*mockModel) RemoveListener(model.TableListener)                  {}
func (*mockModel) Watch(context.Context) error                         { return nil }
func (*mockModel) Refresh(context.Context) error                       { return nil }
func (*mockModel) Get(context.Context, string) (runtime.Object, error) {
	return nil, nil
}

func (*mockModel) Delete(context.Context, string, *metav1.DeletionPropagation, dao.Grace) error {
	return nil
}

func (*mockModel) Describe(context.Context, string) (string, error) {
	return "", nil
}

func (*mockModel) ToYAML(context.Context, string) (string, error) {
	return "", nil
}

func (*mockModel) InNamespace(string) bool      { return true }
func (*mockModel) SetRefreshRate(time.Duration) {}

func makeTableData() *model1.TableData {
	return model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "RESOURCE"},
			model1.HeaderColumn{Name: "COMMAND"},
			model1.HeaderColumn{Name: "APIGROUP"},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					ID:     "r1",
					Fields: model1.Fields{"blee", "duh", "fred"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					ID:     "r2",
					Fields: model1.Fields{"fred", "duh", "zorg"},
				},
			},
		),
	)
}
