package view_test

import (
	"bytes"
	"fmt"
	"os"
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
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := view.NewLog(client.NewGVR("v1/pods"), &opts)
	v.Init(makeContext())

	ii := dao.NewLogItems()
	ii.Add(dao.NewLogItemFromString("blee"), dao.NewLogItemFromString("bozo"))
	v.Flush(ii.Lines(false))

	assert.Equal(t, 29, len(v.Logs().GetText(true)))
}

func BenchmarkLogFlush(b *testing.B) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := view.NewLog(client.NewGVR("v1/pods"), &opts)
	v.Init(makeContext())

	items := dao.NewLogItems()
	items.Add(
		dao.NewLogItemFromString("blee"),
		dao.NewLogItemFromString("bozo"),
	)
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
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := view.NewLog(client.NewGVR("v1/pods"), &opts)
	v.Init(makeContext())

	app := makeApp()
	ii := dao.NewLogItems()
	ii.Add(dao.NewLogItemFromString("blee"), dao.NewLogItemFromString("bozo"))
	v.Flush(ii.Lines(false))
	config.K9sDumpDir = "/tmp"
	dir := filepath.Join(config.K9sDumpDir, app.Config.K9s.CurrentCluster)
	c1, _ := os.ReadDir(dir)
	v.SaveCmd(nil)
	c2, _ := os.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestAllContainerKeyBinding(t *testing.T) {
	uu := map[string]struct {
		opts *dao.LogOptions
		e    bool
	}{
		"action-present": {
			opts: &dao.LogOptions{Path: "", DefaultContainer: "container"},
			e:    true,
		},
		"action-missing": {
			opts: &dao.LogOptions{},
		},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			v := view.NewLog(client.NewGVR("v1/pods"), u.opts)
			v.Init(makeContext())
			_, got := v.Logs().Actions()[ui.KeyA]
			assert.Equal(t, u.e, got)
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeApp() *view.App {
	return view.NewApp(config.NewConfig(ks{}))
}
