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
	v.GetTable().SetModel(&mockModel{})
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

type mockModel struct{}

var _ ui.Tabular = (*mockModel)(nil)
var _ ui.Suggester = (*mockModel)(nil)

func (t *mockModel) CurrentSuggestion() (string, bool)  { return "", false }
func (t *mockModel) NextSuggestion() (string, bool)     { return "", false }
func (t *mockModel) PrevSuggestion() (string, bool)     { return "", false }
func (t *mockModel) ClearSuggestions()                  {}
func (t *mockModel) SetInstance(string)                 {}
func (t *mockModel) SetLabelFilter(string)              {}
func (t *mockModel) Empty() bool                        { return false }
func (t *mockModel) HasMetrics() bool                   { return true }
func (t *mockModel) Peek() render.TableData             { return makeTableData() }
func (t *mockModel) ClusterWide() bool                  { return false }
func (t *mockModel) GetNamespace() string               { return "blee" }
func (t *mockModel) SetNamespace(string)                {}
func (t *mockModel) ToggleToast()                       {}
func (t *mockModel) AddListener(model.TableListener)    {}
func (t *mockModel) RemoveListener(model.TableListener) {}
func (t *mockModel) Watch(context.Context)              {}
func (t *mockModel) Refresh(context.Context)            {}
func (t *mockModel) Get(context.Context, string) (runtime.Object, error) {

	return nil, nil
}
func (t *mockModel) Delete(context.Context, string, bool, bool) error {
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
