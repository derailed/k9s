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
	v.SetTitle("k9s-test")
	dir := filepath.Join(config.K9sDumpDir, v.app.Config.K9s.CurrentCluster)
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
	v.Update(data)
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
	v.Update(data)
	v.Cmd().SetActive(true)
	v.Cmd().Set([]rune("blee"))
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
	v.Update(data)
	v.SortColCmd(1)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee ", v.GetCell(1, 1).Text)

	v.SortInvertCmd(nil)
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
