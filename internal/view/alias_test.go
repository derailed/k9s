package view_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestAliasNew(t *testing.T) {
	v := view.NewAlias(client.GVR("aliases"))

	assert.Nil(t, v.Init(makeContext()))
	assert.Equal(t, "Aliases", v.Name())
	assert.Equal(t, 4, len(v.Hints()))
}

func TestAliasSearch(t *testing.T) {
	v := view.NewAlias(client.GVR("aliases"))
	assert.Nil(t, v.Init(makeContext()))
	v.GetTable().SetModel(&testModel{})
	v.GetTable().SearchBuff().SetActive(true)
	v.GetTable().SearchBuff().Set("dump")

	v.GetTable().SendKey(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone))

	assert.Equal(t, 3, v.GetTable().GetColumnCount())
	assert.Equal(t, 1, v.GetTable().GetRowCount())
}

func TestAliasGoto(t *testing.T) {
	v := view.NewAlias(client.GVR("aliases"))
	assert.Nil(t, v.Init(makeContext()))
	v.GetTable().Select(0, 0)

	b := buffL{}
	v.GetTable().SearchBuff().SetActive(true)
	v.GetTable().SearchBuff().AddListener(&b)
	v.GetTable().SendKey(tcell.NewEventKey(tcell.KeyEnter, 256, tcell.ModNone))

	assert.True(t, v.GetTable().SearchBuff().IsActive())
}

// ----------------------------------------------------------------------------
// Helpers...

type buffL struct {
	active  int
	changed int
}

func (b *buffL) BufferChanged(s string) {
	b.changed++
}
func (b *buffL) BufferActive(state bool, kind ui.BufferKind) {
	b.active++
}

func makeContext() context.Context {
	a := view.NewApp(config.NewConfig(ks{}))
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

type ks struct{}

func (k ks) CurrentContextName() (string, error) {
	return "test", nil
}

func (k ks) CurrentClusterName() (string, error) {
	return "test", nil
}

func (k ks) CurrentNamespaceName() (string, error) {
	return "test", nil
}

func (k ks) ClusterNames() ([]string, error) {
	return []string{"test"}, nil
}

func (k ks) NamespaceNames(nn []v1.Namespace) []string {
	return []string{"test"}
}

type testModel struct{}

var _ ui.Tabular = &testModel{}

func (t *testModel) Empty() bool                     { return false }
func (t *testModel) Peek() render.TableData          { return makeTableData() }
func (t *testModel) ClusterWide() bool               { return false }
func (t *testModel) GetNamespace() string            { return "blee" }
func (t *testModel) SetNamespace(string)             {}
func (t *testModel) AddListener(model.TableListener) {}
func (t *testModel) Watch(context.Context)           {}
func (t *testModel) InNamespace(string) bool         { return true }
func (t *testModel) SetRefreshRate(time.Duration)    {}

func makeTableData() render.TableData {
	return render.TableData{
		Namespace: render.ClusterScope,
		Header: render.HeaderRow{
			render.Header{Name: "RESOURCE"},
			render.Header{Name: "COMMAND"},
			render.Header{Name: "APIGROUP"},
		},
		RowEvents: render.RowEvents{
			render.RowEvent{
				Row: render.Row{
					ID:     "r1",
					Fields: render.Fields{"blee", "duh", "fred"},
				},
			},
			render.RowEvent{
				Row: render.Row{
					ID:     "r2",
					Fields: render.Fields{"fred", "duh", "zorg"},
				},
			},
		},
	}
}
