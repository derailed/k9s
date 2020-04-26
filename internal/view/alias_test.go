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
	"k8s.io/apimachinery/pkg/runtime"
)

func TestAliasNew(t *testing.T) {
	v := view.NewAlias(client.NewGVR("aliases"))

	assert.Nil(t, v.Init(makeContext()))
	assert.Equal(t, "Aliases", v.Name())
	assert.Equal(t, 5, len(v.Hints()))
}

func TestAliasSearch(t *testing.T) {
	v := view.NewAlias(client.NewGVR("aliases"))
	assert.Nil(t, v.Init(makeContext()))
	v.GetTable().SetModel(&testModel{})
	v.GetTable().Refresh()
	v.App().Prompt().SetModel(v.GetTable().CmdBuff())
	v.App().Prompt().SendStrokes("blee")

	assert.Equal(t, 3, v.GetTable().GetColumnCount())
	assert.Equal(t, 2, v.GetTable().GetRowCount())
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

func (b *buffL) BufferChanged(s string) {
	b.changed++
}
func (b *buffL) BufferActive(state bool, kind model.BufferKind) {
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

var _ ui.Tabular = (*testModel)(nil)
var _ ui.Suggester = (*testModel)(nil)

func (t *testModel) CurrentSuggestion() (string, bool) { return "", false }
func (t *testModel) NextSuggestion() (string, bool)    { return "", false }
func (t *testModel) PrevSuggestion() (string, bool)    { return "", false }
func (t *testModel) ClearSuggestions()                 {}

func (t *testModel) SetInstance(string)              {}
func (t *testModel) Empty() bool                     { return false }
func (t *testModel) HasMetrics() bool                { return true }
func (t *testModel) Peek() render.TableData          { return makeTableData() }
func (t *testModel) ClusterWide() bool               { return false }
func (t *testModel) GetNamespace() string            { return "blee" }
func (t *testModel) SetNamespace(string)             {}
func (t *testModel) ToggleToast()                    {}
func (t *testModel) AddListener(model.TableListener) {}
func (t *testModel) Watch(context.Context)           {}
func (t *testModel) Get(context.Context, string) (runtime.Object, error) {
	return nil, nil
}
func (t *testModel) Delete(context.Context, string, bool, bool) error {
	return nil
}
func (t *testModel) Describe(context.Context, string) (string, error) {
	return "", nil
}
func (t *testModel) ToYAML(ctx context.Context, path string) (string, error) {
	return "", nil
}

func (t *testModel) InNamespace(string) bool      { return true }
func (t *testModel) SetRefreshRate(time.Duration) {}

func makeTableData() render.TableData {
	return render.TableData{
		Namespace: client.ClusterScope,
		Header: render.Header{
			render.HeaderColumn{Name: "RESOURCE"},
			render.HeaderColumn{Name: "COMMAND"},
			render.HeaderColumn{Name: "APIGROUP"},
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
