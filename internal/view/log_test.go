package view_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/view"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestAnsi(t *testing.T) {
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

func TestLogFlush(t *testing.T) {
	v := view.NewLog("Logs", makeApp(), nil)
	v.Flush(2, []string{"blee", "bozo"})

	v.ToggleAutoScrollCmd(nil)
	assert.Equal(t, "blee\nbozo\n", v.Logs().GetText(true))
	assert.Equal(t, " Autoscroll: Off ", v.ScrollIndicator().GetText(true))
	v.ToggleAutoScrollCmd(nil)
	assert.Equal(t, " Autoscroll: On  ", v.ScrollIndicator().GetText(true))
	assert.Equal(t, 8, len(v.Hints()))
}

func TestLogViewSave(t *testing.T) {
	app := makeApp()
	v := view.NewLog("Logs", app, nil)
	v.Flush(2, []string{"blee", "bozo"})
	config.K9sDumpDir = "/tmp"
	dir := filepath.Join(config.K9sDumpDir, app.Config.K9s.CurrentCluster)
	c1, _ := ioutil.ReadDir(dir)
	v.SaveCmd(nil)
	c2, _ := ioutil.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestLogViewNav(t *testing.T) {
	v := view.NewLog("Logs", makeApp(), nil)
	var buff []string
	for i := 0; i < 100; i++ {
		buff = append(buff, fmt.Sprintf("line-%d\n", i))
	}
	v.Flush(100, buff)
	v.ToggleAutoScrollCmd(nil)

	r, _ := v.Logs().GetScrollOffset()
	assert.Equal(t, -1, r)
}

func TestLogViewClear(t *testing.T) {
	v := view.NewLog("Logs", makeApp(), nil)
	v.Flush(2, []string{"blee", "bozo"})

	v.ToggleAutoScrollCmd(nil)
	assert.Equal(t, "blee\nbozo\n", v.Logs().GetText(true))
	v.Logs().Clear()
	assert.Equal(t, "", v.Logs().GetText(true))
}

// Helpers...

func makeApp() *view.App {
	return view.NewApp(config.NewConfig(ks{}))
}
