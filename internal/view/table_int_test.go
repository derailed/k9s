package view

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestTableSave(t *testing.T) {
	v := NewTable("test")
	v.Init(makeContext())
	v.SetTitle("k9s-test")

	dir := filepath.Join(config.K9sDumpDir, v.app.Config.K9s.CurrentCluster)
	c1, _ := ioutil.ReadDir(dir)
	v.saveCmd(nil)

	c2, _ := ioutil.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestTableNew(t *testing.T) {
	v := NewTable("test")
	v.Init(makeContext())

	data := render.TableData{
		Header: render.HeaderRow{
			render.Header{Name: "NAMESPACE"},
			render.Header{Name: "NAME", Align: tview.AlignRight},
			render.Header{Name: "FRED"},
			render.Header{Name: "AGE", Decorator: render.AgeDecorator},
		},
		RowEvents: render.RowEvents{
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"ns1", "a", "10", "3m"},
				},
			},
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"ns1", "b", "15", "1m"},
				},
			},
		},
		Namespace: "",
	}
	v.Update(data)
	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewFilter(t *testing.T) {
	v := NewTable("test")
	v.Init(makeContext())

	data := render.TableData{
		Header: render.HeaderRow{
			render.Header{Name: "NAMESPACE"},
			render.Header{Name: "NAME", Align: tview.AlignRight},
			render.Header{Name: "FRED"},
			render.Header{Name: "AGE", Decorator: render.AgeDecorator},
		},
		RowEvents: render.RowEvents{
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"ns1", "blee", "10", "3m"},
				},
			},
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"ns1", "fred", "15", "1m"},
				},
			},
		},
		Namespace: "",
	}
	v.Update(data)
	v.SearchBuff().SetActive(true)
	v.SearchBuff().Set("blee")
	v.filterCmd(nil)
	assert.Equal(t, 2, v.GetRowCount())
	v.resetCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewSort(t *testing.T) {
	v := NewTable("test")
	v.Init(makeContext())

	data := render.TableData{
		Header: render.HeaderRow{
			render.Header{Name: "NAMESPACE"},
			render.Header{Name: "NAME", Align: tview.AlignRight},
			render.Header{Name: "FRED"},
			render.Header{Name: "AGE", Decorator: render.AgeDecorator},
		},
		RowEvents: render.RowEvents{
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"ns1", "blee", "10", "3m"},
				},
			},
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"ns1", "fred", "15", "1m"},
				},
				Deltas: render.DeltaRow{"", "", "20", ""},
			},
		},
		Namespace: "",
	}
	v.Update(data)
	v.SortColCmd(1, true)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee", v.GetCell(1, 1).Text)

	v.SortInvertCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "fred", v.GetCell(1, 1).Text)
}

// Helpers...

func makeContext() context.Context {
	a := NewApp(config.NewConfig(ks{}))
	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	return context.WithValue(ctx, ui.KeyStyles, a.Styles)
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
