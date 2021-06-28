package view

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestLogAutoScroll(t *testing.T) {
	opts := dao.LogOptions{
		Path:            "fred/p1",
		Container:       "blee",
		SingleContainer: true,
	}
	v := NewLog(client.NewGVR("v1/pods"), &opts)
	v.Init(makeContext())
	ii := dao.NewLogItems()
	ii.Add(dao.NewLogItemFromString("blee"), dao.NewLogItemFromString("bozo"))
	v.GetModel().Set(ii)
	v.GetModel().Notify()

	assert.Equal(t, 15, len(v.Hints()))

	v.toggleAutoScrollCmd(nil)
	assert.Equal(t, "Autoscroll:Off     FullScreen:Off     Timestamps:Off     Wrap:Off", v.Indicator().GetText(true))
}

func TestLogViewNav(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := NewLog(client.NewGVR("v1/pods"), &opts)
	v.Init(makeContext())

	buff := dao.NewLogItems()
	for i := 0; i < 100; i++ {
		buff.Add(dao.NewLogItemFromString(fmt.Sprintf("line-%d\n", i)))
	}
	v.GetModel().Set(buff)
	v.toggleAutoScrollCmd(nil)

	r, _ := v.Logs().GetScrollOffset()
	assert.Equal(t, -1, r)
}

func TestLogViewClear(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := NewLog(client.NewGVR("v1/pods"), &opts)
	v.Init(makeContext())

	v.toggleAutoScrollCmd(nil)
	v.Logs().SetText("blee\nblah")
	v.Logs().Clear()

	assert.Equal(t, "", v.Logs().GetText(true))
}

func TestLogTimestamp(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/blee",
		Container: "c1",
	}
	l := NewLog(client.NewGVR("test"), &opts)
	l.Init(makeContext())
	ii := dao.NewLogItems()
	ii.Add(
		&dao.LogItem{
			Pod:       "fred/blee",
			Container: "c1",
			Timestamp: "ttt",
			Bytes:     []byte("Testing 1, 2, 3"),
		},
	)
	var list logList
	l.GetModel().AddListener(&list)
	l.GetModel().Set(ii)
	l.SendKeys(ui.KeyT)
	l.Logs().Clear()
	l.Flush(ii.Lines(true))

	assert.Equal(t, fmt.Sprintf("\n%-30s %s", "ttt", "fred/blee:c1 Testing 1, 2, 3"), l.Logs().GetText(true))
	assert.Equal(t, 2, list.change)
	assert.Equal(t, 2, list.clear)
	assert.Equal(t, 0, list.fail)
}

func TestLogFilter(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/blee",
		Container: "c1",
	}
	l := NewLog(client.NewGVR("test"), &opts)
	l.Init(makeContext())
	buff := dao.NewLogItems()
	buff.Add(
		dao.NewLogItemFromString("duh"),
		dao.NewLogItemFromString("zorg"),
	)
	var list logList
	l.GetModel().AddListener(&list)
	l.GetModel().Set(buff)
	l.SendKeys(ui.KeySlash)
	l.SendStrokes("zorg")

	assert.Equal(t, "duhzorg", list.lines)
	assert.Equal(t, 1, list.change)
	assert.Equal(t, 1, list.clear)
	assert.Equal(t, 0, list.fail)
}

// ----------------------------------------------------------------------------
// Helpers...

type logList struct {
	change, clear, fail int
	lines               string
}

func (l *logList) LogChanged(ll [][]byte) {
	l.change++
	l.lines = ""
	for _, line := range ll {
		l.lines += string(line)
	}
}
func (l *logList) LogCleared()     { l.clear++ }
func (l *logList) LogFailed(error) { l.fail++ }
