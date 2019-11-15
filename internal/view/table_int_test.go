package view

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
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

	data := resource.TableData{
		Header: resource.Row{"NAMESPACE", "NAME", "FRED", "AGE"},
		Rows: resource.RowEvents{
			"ns1/a": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"ns1", "a", "10", "3m"},
				Deltas: resource.Row{"", "", "", ""},
			},
			"ns1/b": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"ns1", "b", "15", "1m"},
				Deltas: resource.Row{"", "", "20", ""},
			},
		},
		NumCols: map[string]bool{
			"FRED": true,
		},
		Namespace: "",
	}
	v.Update(data)
	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewFilter(t *testing.T) {
	v := NewTable("test")
	v.Init(makeContext())

	data := resource.TableData{
		Header: resource.Row{"NAMESPACE", "NAME", "FRED", "AGE"},
		Rows: resource.RowEvents{
			"ns1/blee": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"ns1", "blee", "10", "3m"},
				Deltas: resource.Row{"", "", "", ""},
			},
			"ns1/fred": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"ns1", "fred", "15", "1m"},
				Deltas: resource.Row{"", "", "20", ""},
			},
		},
		NumCols: map[string]bool{
			"FRED": true,
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

	data := resource.TableData{
		Header: resource.Row{"NAMESPACE", "NAME", "FRED", "AGE"},
		Rows: resource.RowEvents{
			"ns1/blee": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"ns1", "blee", "10", "3m"},
				Deltas: resource.Row{"", "", "", ""},
			},
			"ns1/fred": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"ns1", "fred", "15", "1m"},
				Deltas: resource.Row{"", "", "20", ""},
			},
		},
		NumCols: map[string]bool{
			"FRED": true,
		},
		Namespace: "",
	}
	v.Update(data)
	v.SortColCmd(1)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee ", v.GetCell(1, 1).Text)

	v.SortInvertCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "fred ", v.GetCell(1, 1).Text)
}

// Helpers...

func makeContext() context.Context {
	a := NewApp(config.NewConfig(ks{}))
	ctx := context.WithValue(context.Background(), ui.KeyApp, a)
	return context.WithValue(ctx, ui.KeyStyles, a.Styles)
}
