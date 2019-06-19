package views

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
)

func TestTableViewSave(t *testing.T) {
	v := newTableView(NewApp(config.NewConfig(ks{})), "test")
	v.baseTitle = "k9s-test"
	dir := filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster)
	c1, _ := ioutil.ReadDir(dir)
	v.saveCmd(nil)
	c2, _ := ioutil.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestTableViewNew(t *testing.T) {
	v := newTableView(NewApp(config.NewConfig(ks{})), "test")

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
	v.update(data)
	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewFilter(t *testing.T) {
	v := newTableView(NewApp(config.NewConfig(ks{})), "test")

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
	v.update(data)
	v.cmdBuff.setActive(true)
	v.cmdBuff.buff = []rune("blee")
	v.filterCmd(nil)
	assert.Equal(t, 2, v.GetRowCount())
	v.resetCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewSort(t *testing.T) {
	v := newTableView(NewApp(config.NewConfig(ks{})), "test")

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
	v.update(data)
	v.sortColCmd(1)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee ", v.GetCell(1, 1).Text)

	v.sortInvertCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "fred ", v.GetCell(1, 1).Text)
}

func TestIsSelector(t *testing.T) {
	uu := map[string]struct {
		sel string
		e   bool
	}{
		"cool":       {"-l app=fred,env=blee", true},
		"noMode":     {"app=fred,env=blee", false},
		"noSpace":    {"-lapp=fred,env=blee", true},
		"wrongLabel": {"-f app=fred,env=blee", false},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isLabelSelector(u.sel))
		})
	}
}

func TestTrimLabelSelector(t *testing.T) {
	uu := map[string]struct {
		sel, e string
	}{
		"cool":    {"-l app=fred,env=blee", "app=fred,env=blee"},
		"noSpace": {"-lapp=fred,env=blee", "app=fred,env=blee"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, trimLabelSelector(u.sel))
		})
	}
}

func TestTVSortRows(t *testing.T) {
	uu := []struct {
		rows  resource.RowEvents
		col   int
		asc   bool
		first resource.Row
		e     []string
	}{
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"x", "y"}},
				"row2": {Fields: resource.Row{"a", "b"}},
			},
			0,
			true,
			resource.Row{"a", "b"},
			[]string{"row2", "row1"},
		},
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"x", "y"}},
				"row2": {Fields: resource.Row{"a", "b"}},
			},
			1,
			true,
			resource.Row{"a", "b"},
			[]string{"row2", "row1"},
		},
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"x", "y"}},
				"row2": {Fields: resource.Row{"a", "b"}},
			},
			1,
			false,
			resource.Row{"x", "y"},
			[]string{"row1", "row2"},
		},
		{
			resource.RowEvents{
				"row1": {Fields: resource.Row{"2175h48m0.06015s", "y"}},
				"row2": {Fields: resource.Row{"403h42m34.060166s", "b"}},
			},
			0,
			true,
			resource.Row{"403h42m34.060166s", "b"},
			[]string{"row2", "row1"},
		},
	}

	for _, u := range uu {
		keys := make([]string, len(u.rows))
		sortRows(u.rows, defaultSort, sortColumn{u.col, len(u.rows), u.asc}, keys)
		assert.Equal(t, u.e, keys)
		assert.Equal(t, u.first, u.rows[u.e[0]].Fields)
	}
}

func BenchmarkTVSortRows(b *testing.B) {
	evts := resource.RowEvents{
		"row1": {Fields: resource.Row{"x", "y"}},
		"row2": {Fields: resource.Row{"a", "b"}},
	}
	sc := sortColumn{0, 2, true}
	keys := make([]string, len(evts))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sortRows(evts, defaultSort, sc, keys)
	}
}

func BenchmarkTitleReplace(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fmat := strings.Replace(nsTitleFmt, "[fg", "["+"red", -1)
		fmat = strings.Replace(fmat, ":bg:", ":"+"blue"+":", -1)
		fmat = strings.Replace(fmat, "[hilite", "["+"green", 1)
		fmat = strings.Replace(fmat, "[count", "["+"yellow", 1)
		_ = fmt.Sprintf(fmat, "Pods", "default", 10)
	}
}

func BenchmarkTitleReplace1(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fmat := strings.Replace(nsTitleFmt, "fg:bg", "red"+":"+"blue", -1)
		fmat = strings.Replace(fmat, "[hilite", "["+"green", 1)
		fmat = strings.Replace(fmat, "[count", "["+"yellow", 1)
		_ = fmt.Sprintf(fmat, "Pods", "default", 10)
	}
}
