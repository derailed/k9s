// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
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

	assert.Len(t, v.Hints(), 18)

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

func TestLogClearThenChanged(t *testing.T) {
	v, _ := runLogView(t)

	for i := range 50 {
		line := "blee-" + strconv.Itoa(i) + "\n"
		v.LogCleared()
		v.LogChanged([][]byte{[]byte(line)})
		require.Eventually(t, func() bool {
			return v.logText() == line
		}, 2*time.Second, time.Millisecond, "round %d", i)
	}
}

func TestLogChangedKeepsOrder(t *testing.T) {
	v, _ := runLogView(t)

	const size = 500
	want := make([]string, 0, size)
	for i := range size {
		want = append(want, strconv.Itoa(i))
		v.LogChanged([][]byte{[]byte(want[i] + "\n")})
	}
	require.Eventually(t, func() bool {
		return len(strings.Fields(v.logText())) == size
	}, 5*time.Second, time.Millisecond)
	assert.Equal(t, want, strings.Fields(v.logText()))
}

func TestLogChangedCoalescesWhileBusy(t *testing.T) {
	v, draws := runLogView(t)
	a := v.app.Application

	busy, release := make(chan struct{}), make(chan struct{})
	go a.QueueUpdate(func() {
		close(busy)
		<-release
	})
	<-busy
	draws.Store(0)

	const size = 100
	for i := range size {
		v.LogChanged([][]byte{[]byte(strconv.Itoa(i) + "\n")})
	}
	close(release)

	require.Eventually(t, func() bool {
		return len(strings.Fields(v.logText())) == size
	}, 5*time.Second, time.Millisecond)
	assert.Equal(t, int64(1), draws.Load())
}

// ----------------------------------------------------------------------------
// Helpers...

// runLogView runs a log view on a simulated screen and counts the redraws.
func runLogView(t *testing.T) (*Log, *atomic.Int64) {
	t.Helper()

	opts := dao.LogOptions{Path: "fred/p1", Container: "blee"}
	v := NewLog(client.PodGVR, &opts)
	require.NoError(t, v.Init(makeContext(t)))

	s := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, s.Init())

	var draws atomic.Int64
	a := v.app.Application
	a.SetScreen(s).SetRoot(v, true)
	a.SetAfterDrawFunc(func(tcell.Screen) { draws.Add(1) })

	done := make(chan error, 1)
	go func() { done <- a.Run() }()
	t.Cleanup(func() {
		a.Stop()
		require.NoError(t, <-done)
	})

	return v, &draws
}

// logText reads the text on the UI goroutine.
func (l *Log) logText() string {
	var text string
	l.app.Application.QueueUpdate(func() {
		text = l.logs.GetText(true)
	})

	return text
}
