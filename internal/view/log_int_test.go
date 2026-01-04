// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogAutoScroll(t *testing.T) {
	opts := dao.LogOptions{
		Path:            "fred/p1",
		Container:       "blee",
		SingleContainer: true,
	}
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))
	ii := dao.NewLogItems()
	ii.Add(dao.NewLogItemFromString("blee"), dao.NewLogItemFromString("bozo"))
	v.GetModel().Set(ii)
	v.GetModel().Notify()

	assert.Len(t, v.Hints(), 20)

	v.toggleAutoScrollCmd(nil)
	assert.Equal(t, "Autoscroll:Off     ColumnLock:Off     FullScreen:Off     Timestamps:Off     Wrap:Off", v.Indicator().GetText(true))
}

func TestLogColumnLock(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))

	buff := dao.NewLogItems()
	for i := range 100 {
		buff.Add(dao.NewLogItemFromString(fmt.Sprintf("line-%d\n", i)))
	}
	v.GetModel().Set(buff)

	v.toggleColumnLockCmd(nil)
	const column = 2
	v.Logs().ScrollTo(-1, column)
	v.toggleAutoScrollCmd(nil)

	r, c := v.Logs().GetScrollOffset()
	assert.Equal(t, -1, r)
	assert.Equal(t, column, c)
}

func TestLogHorizontalScroll(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))
	if v.indicator.TextWrap() {
		v.toggleTextWrapCmd(nil)
	}

	buff := dao.NewLogItems()
	buff.Add(dao.NewLogItemFromString("this is a very long log line that needs horizontal scrolling to view completely\n"))
	v.GetModel().Set(buff)
	v.toggleAutoScrollCmd(nil)

	_, c := v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Initial column offset should be 0")

	v.scrollRightCmd(nil)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 16, c, "Column offset should be 16 after scroll right")

	v.scrollRightCmd(nil)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 32, c, "Column offset should be 32 after second scroll right")

	v.scrollLeftCmd(nil)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 16, c, "Column offset should be 16 after scroll left")

	v.scrollLeftCmd(nil)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Column offset should be 0 after second scroll left")

	v.scrollLeftCmd(nil)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Column offset should stay at 0 when scrolling left at beginning")

	v.Logs().ScrollTo(0, 0)
	if !v.indicator.TextWrap() {
		v.toggleTextWrapCmd(nil)
	}
	v.scrollRightCmd(nil)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Column offset should stay at 0 when wrap is enabled")
}

func TestLogMouseHorizontalScroll(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))
	if v.indicator.TextWrap() {
		v.toggleTextWrapCmd(nil)
	}

	buff := dao.NewLogItems()
	buff.Add(dao.NewLogItemFromString("this is a very long log line that needs horizontal scrolling to view completely\n"))
	v.GetModel().Set(buff)
	v.toggleAutoScrollCmd(nil)

	_, c := v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Initial column offset should be 0")

	// Test mouse scroll right
	event := tcell.NewEventMouse(0, 0, tcell.WheelRight, tcell.ModNone)
	v.mouseHandler(tview.MouseScrollRight, event)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 16, c, "Column offset should be 16 after mouse scroll right")

	// Test mouse scroll right again
	v.mouseHandler(tview.MouseScrollRight, event)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 32, c, "Column offset should be 32 after second mouse scroll right")

	// Test mouse scroll left
	event = tcell.NewEventMouse(0, 0, tcell.WheelLeft, tcell.ModNone)
	v.mouseHandler(tview.MouseScrollLeft, event)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 16, c, "Column offset should be 16 after mouse scroll left")

	// Test mouse scroll left again
	v.mouseHandler(tview.MouseScrollLeft, event)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Column offset should be 0 after second mouse scroll left")

	// Test boundary: scroll left at position 0
	v.mouseHandler(tview.MouseScrollLeft, event)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Column offset should stay at 0 when mouse scrolling left at beginning")

	// Test wrap mode disables mouse horizontal scroll
	v.Logs().ScrollTo(0, 0)
	if !v.indicator.TextWrap() {
		v.toggleTextWrapCmd(nil)
	}
	event = tcell.NewEventMouse(0, 0, tcell.WheelRight, tcell.ModNone)
	v.mouseHandler(tview.MouseScrollRight, event)
	_, c = v.Logs().GetScrollOffset()
	assert.Equal(t, 0, c, "Column offset should stay at 0 when wrap is enabled with mouse scroll")
}

func TestLogViewNav(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/p1",
		Container: "blee",
	}
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))

	buff := dao.NewLogItems()
	for i := range 100 {
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
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))

	v.toggleAutoScrollCmd(nil)
	v.Logs().SetText("blee\nblah")
	v.Logs().Clear()

	assert.Empty(t, v.Logs().GetText(true))
}

func TestLogTimestamp(t *testing.T) {
	opts := dao.LogOptions{
		Path:      "fred/blee",
		Container: "c1",
	}
	l := NewLog(client.NewGVR("test"), &opts)
	require.NoError(t, l.Init(makeContext(t)))
	ii := dao.NewLogItems()
	ii.Add(
		&dao.LogItem{
			Pod:       "fred/blee",
			Container: "c1",
			Bytes:     []byte("ttt Testing 1, 2, 3\n"),
		},
	)
	var list logList
	l.GetModel().AddListener(&list)
	l.GetModel().Set(ii)
	l.SendKeys(ui.KeyT)
	l.Logs().Clear()
	ll := make([][]byte, ii.Len())
	ii.Lines(0, true, ll)
	l.Flush(ll)

	assert.Equal(t, fmt.Sprintf("%-30s %s", "ttt", "fred/blee c1 Testing 1, 2, 3\n"), l.Logs().GetText(true))
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
	require.NoError(t, l.Init(makeContext(t)))
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
func (*logList) LogCanceled()      {}
func (*logList) LogStop()          {}
func (*logList) LogResume()        {}
func (l *logList) LogCleared()     { l.clear++ }
func (l *logList) LogFailed(error) { l.fail++ }
