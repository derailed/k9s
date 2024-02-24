// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAliasNew(t *testing.T) {
	v := view.NewAlias(client.NewGVR("aliases"))

	assert.Nil(t, v.Init(makeContext()))
	assert.Equal(t, "Aliases", v.Name())
	assert.Equal(t, 6, len(v.Hints()))
}

func TestAliasSearch(t *testing.T) {
	v := view.NewAlias(client.NewGVR("aliases"))
	assert.Nil(t, v.Init(makeContext()))
	v.GetTable().SetModel(&mockModel{})
	v.GetTable().Refresh()
	v.App().Prompt().SetModel(v.GetTable().CmdBuff())
	v.App().Prompt().SendStrokes("blee")

	assert.Equal(t, 3, v.GetTable().GetColumnCount())
	assert.Equal(t, 3, v.GetTable().GetRowCount())
}

func TestAliasGoto(t *testing.T) {
	v := view.NewAlias(client.NewGVR("aliases"))
	assert.Nil(t, v.Init(makeContext()))
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

func (b *buffL) BufferChanged(_, _ string) {
	b.changed++
}
func (b *buffL) BufferCompleted(_, _ string) {}

func (b *buffL) BufferActive(state bool, kind model.BufferKind) {
	b.active++
}

func makeContext() context.Context {
	a := view.NewApp(mock.NewMockConfig())
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

type mockModel struct{}

var (
	_ ui.Tabular   = (*mockModel)(nil)
	_ ui.Suggester = (*mockModel)(nil)
)

func (t *mockModel) CurrentSuggestion() (string, bool)  { return "", false }
func (t *mockModel) NextSuggestion() (string, bool)     { return "", false }
func (t *mockModel) PrevSuggestion() (string, bool)     { return "", false }
func (t *mockModel) ClearSuggestions()                  {}
func (t *mockModel) SetInstance(string)                 {}
func (t *mockModel) SetLabelFilter(string)              {}
func (t *mockModel) GetLabelFilter() string             { return "" }
func (t *mockModel) Empty() bool                        { return false }
func (t *mockModel) RowCount() int                      { return 1 }
func (t *mockModel) HasMetrics() bool                   { return true }
func (t *mockModel) Peek() *model1.TableData            { return makeTableData() }
func (t *mockModel) ClusterWide() bool                  { return false }
func (t *mockModel) GetNamespace() string               { return "blee" }
func (t *mockModel) SetNamespace(string)                {}
func (t *mockModel) ToggleToast()                       {}
func (t *mockModel) AddListener(model.TableListener)    {}
func (t *mockModel) RemoveListener(model.TableListener) {}
func (t *mockModel) Watch(context.Context) error        { return nil }
func (t *mockModel) Refresh(context.Context) error      { return nil }
func (t *mockModel) Get(context.Context, string) (runtime.Object, error) {
	return nil, nil
}

func (t *mockModel) Delete(context.Context, string, *metav1.DeletionPropagation, dao.Grace) error {
	return nil
}

func (t *mockModel) Describe(context.Context, string) (string, error) {
	return "", nil
}

func (t *mockModel) ToYAML(ctx context.Context, path string) (string, error) {
	return "", nil
}

func (t *mockModel) InNamespace(string) bool      { return true }
func (t *mockModel) SetRefreshRate(time.Duration) {}

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
