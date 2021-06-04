package view_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	v := view.NewLog(client.NewGVR("v1/pods"), "fred/p1", "blee", false)
	v.Init(makeContext())

	v.Flush(dao.LogItems{
		dao.NewLogItemFromString("blee"),
		dao.NewLogItemFromString("bozo"),
	}.Lines(false))

	assert.Equal(t, 29, len(v.Logs().GetText(true)))
}

func BenchmarkLogFlush(b *testing.B) {
	v := view.NewLog(client.NewGVR("v1/pods"), "fred/p1", "blee", false)
	v.Init(makeContext())

	items := dao.LogItems{
		dao.NewLogItemFromString("blee"),
		dao.NewLogItemFromString("bozo"),
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		v.Flush(items.Lines(false))
	}
}

func TestLogAnsi(t *testing.T) {
	buff := bytes.NewBufferString("")
	w := tview.ANSIWriter(buff, "white", "black")
	fmt.Fprintf(w, "[YELLOW] ok")
	assert.Equal(t, "[YELLOW] ok", buff.String())

	v := tview.NewTextView()
	v.SetDynamicColors(true)
	aw := tview.ANSIWriter(v, "white", "black")
	s := "[2019-03-27T15:05:15,246][INFO ][o.e.c.r.a.AllocationService] [es-0] Cluster health status changed from [YELLOW] to [GREEN] (reason: [shards started [[.monitoring-es-6-2019.03.27][0]]"
	fmt.Fprintf(aw, "%s", s)
	assert.Equal(t, s+"\n", v.GetText(false))
}

func TestLogViewSave(t *testing.T) {
	v := view.NewLog(client.NewGVR("v1/pods"), "fred/p1", "blee", false)
	v.Init(makeContext())

	app := makeApp()
	v.Flush(dao.LogItems{
		dao.NewLogItemFromString("blee"),
		dao.NewLogItemFromString("bozo"),
	}.Lines(false))
	config.K9sDumpDir = "/tmp"
	dir := filepath.Join(config.K9sDumpDir, app.Config.K9s.CurrentCluster)
	c1, _ := ioutil.ReadDir(dir)
	v.SaveCmd(nil)
	c2, _ := ioutil.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestAllContainerKeyBinding(t *testing.T) {
	uu := map[string]struct {
		l *view.Log
		e bool
	}{
		"all containers":    {view.NewLog(client.NewGVR("v1/pods"), "", "container", false), true},
		"no all containers": {view.NewLog(client.NewGVR("v1/pods"), "", "", false), false},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.l.Init(makeContext())
			_, got := u.l.Logs().Actions()[ui.KeyA]
			assert.Equal(t, u.e, got)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeApp() *view.App {
	return view.NewApp(config.NewConfig(ks{}))
}
