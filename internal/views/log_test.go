package views

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestAnsi(t *testing.T) {
	buff := bytes.NewBufferString("")
	w := tview.ANSIWriter(buff)
	fmt.Fprintf(w, "[YELLOW] ok")
	assert.Equal(t, "[YELLOW] ok", buff.String())

	v := tview.NewTextView()
	v.SetDynamicColors(true)
	aw := tview.ANSIWriter(v)
	s := "[2019-03-27T15:05:15,246][INFO ][o.e.c.r.a.AllocationService] [es-0] Cluster health status changed from [YELLOW] to [GREEN] (reason: [shards started [[.monitoring-es-6-2019.03.27][0]]"
	fmt.Fprintf(aw, s)
	assert.Equal(t, s+"\n", v.GetText(false))
}

func TestLogViewFlush(t *testing.T) {
	v := newLogView("Logs", NewApp(config.NewConfig(ks{})), nil)
	v.flush(2, []string{"blee", "bozo"})

	v.toggleScrollCmd(nil)
	assert.Equal(t, "blee\nbozo\n", v.logs.GetText(true))
	assert.Equal(t, " Autoscroll: Off ", v.status.GetText(true))
	v.toggleScrollCmd(nil)
	assert.Equal(t, " Autoscroll: On  ", v.status.GetText(true))
}

func TestLogViewSave(t *testing.T) {
	v := newLogView("Logs", NewApp(config.NewConfig(ks{})), nil)
	v.flush(2, []string{"blee", "bozo"})
	v.path = "k9s-test"
	dir := filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster)
	c1, _ := ioutil.ReadDir(dir)
	v.saveCmd(nil)
	c2, _ := ioutil.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestLogViewNav(t *testing.T) {
	v := newLogView("Logs", NewApp(config.NewConfig(ks{})), nil)
	var buff []string
	v.autoScroll = 1
	for i := 0; i < 100; i++ {
		buff = append(buff, fmt.Sprintf("line-%d\n", i))
	}
	v.flush(100, buff)

	v.topCmd(nil)
	r, _ := v.logs.GetScrollOffset()
	assert.Equal(t, 0, r)
	v.pageDownCmd(nil)
	r, _ = v.logs.GetScrollOffset()
	assert.Equal(t, 0, r)
	v.pageUpCmd(nil)
	r, _ = v.logs.GetScrollOffset()
	assert.Equal(t, 0, r)
	v.bottomCmd(nil)
	r, _ = v.logs.GetScrollOffset()
	assert.Equal(t, 0, r)
}

func TestLogViewClear(t *testing.T) {
	v := newLogView("Logs", NewApp(config.NewConfig(ks{})), nil)
	v.flush(2, []string{"blee", "bozo"})

	v.toggleScrollCmd(nil)
	assert.Equal(t, "blee\nbozo\n", v.logs.GetText(true))
	v.clearCmd(nil)
	assert.Equal(t, "", v.logs.GetText(true))
}
