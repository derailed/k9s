package model_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestLogFullBuffer(t *testing.T) {
	size := 4
	m := model.NewLog(client.NewGVR("fred"), "Blee", makeLogOpts(size), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	data := make([]string, 0, 2*size)
	for i := 0; i < 2*size; i++ {
		data = append(data, "line"+strconv.Itoa(i))
		m.Append(data[i])
	}
	m.Notify(false)

	assert.Equal(t, 2, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Equal(t, data[4:], v.data)
}

func TestLogFilter(t *testing.T) {
	uu := map[string]struct {
		q string
		e int
	}{
		"plain": {
			q: "line-1",
			e: 2,
		},
		"regexp": {
			q: `\Apod-line-[1-3]{1}\z`,
			e: 3,
		},
		"fuzzy": {
			q: `-f po-l1`,
			e: 2,
		},
	}

	size := 10
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			m := model.NewLog(client.NewGVR("fred"), "Blee", makeLogOpts(size), 10*time.Millisecond)
			m.Init(makeFactory())

			v := newTestView()
			m.AddListener(v)

			m.Filter(u.q)
			var data []string
			for i := 0; i < size; i++ {
				data = append(data, fmt.Sprintf("pod-line-%d", i+1))
				m.Append(data[i])
			}

			m.Notify(true)
			assert.Equal(t, 3, v.dataCalled)
			assert.Equal(t, 2, v.clearCalled)
			assert.Equal(t, 0, v.errCalled)
			assert.Equal(t, u.e, len(v.data))

			m.ClearFilter()
			assert.Equal(t, 4, v.dataCalled)
			assert.Equal(t, 2, v.clearCalled)
			assert.Equal(t, 0, v.errCalled)
			assert.Equal(t, size, len(v.data))
		})
	}
}

func TestLogStartStop(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), "Blee", makeLogOpts(4), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	m.Start()
	data := []string{"line1", "line2"}
	for _, d := range data {
		m.Append(d)
	}
	m.Notify(true)
	m.Stop()

	assert.Equal(t, 2, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Equal(t, 2, len(v.data))
}

func TestLogClear(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), "Blee", makeLogOpts(4), 10*time.Millisecond)
	m.Init(makeFactory())
	assert.Equal(t, "fred", m.GetPath())
	assert.Equal(t, "blee", m.GetContainer())

	v := newTestView()
	m.AddListener(v)

	data := []string{"line1", "line2"}
	for _, d := range data {
		m.Append(d)
	}
	m.Notify(true)
	m.Clear()

	assert.Equal(t, 2, v.dataCalled)
	assert.Equal(t, 2, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Equal(t, 0, len(v.data))
}

func TestLogBasic(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), "Blee", makeLogOpts(2), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	data := []string{"line1", "line2"}
	m.Set(data)

	assert.Equal(t, 2, v.dataCalled)
	assert.Equal(t, 0, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Equal(t, data, v.data)
}

func TestLogAppend(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), "blah blah", makeLogOpts(4), 5*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)
	assert.Equal(t, []string{"blah blah"}, v.data)

	data := []string{"line1", "line2"}
	for _, d := range data {
		m.Append(d)
	}
	assert.Equal(t, 1, v.dataCalled)
	assert.Equal(t, []string{}, v.data)

	m.Notify(true)
	assert.Equal(t, 2, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Equal(t, data, v.data)
}

func TestLogTimedout(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), "Blee", makeLogOpts(4), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	m.Filter("line1")
	data := []string{"line1", "line2", "line3", "line4"}
	for _, d := range data {
		m.Append(d)
	}
	m.Notify(true)
	assert.Equal(t, 3, v.dataCalled)
	assert.Equal(t, 2, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Equal(t, []string{"line1"}, v.data)
}

// ----------------------------------------------------------------------------
// Helpers...

func makeLogOpts(count int) dao.LogOptions {
	return dao.LogOptions{
		Path:      "fred",
		Container: "blee",
		Lines:     int64(count),
	}
}

// ----------------------------------------------------------------------------

type testView struct {
	data        []string
	dataCalled  int
	clearCalled int
	errCalled   int
}

func newTestView() *testView {
	return &testView{}
}

func (t *testView) LogChanged(d []string) {
	t.data = d
	t.dataCalled++
}
func (t *testView) LogCleared() {
	t.clearCalled++
	t.data = []string{}
}
func (t *testView) LogFailed(err error) {
	fmt.Println("LogErr", err)
	t.errCalled++
}

// ----------------------------------------------------------------------------

type testFactory struct{}

var _ dao.Factory = testFactory{}

func (f testFactory) Client() client.Connection {
	return nil
}
func (f testFactory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	return nil, nil
}
func (f testFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (f testFactory) ForResource(ns, gvr string) informers.GenericInformer {
	return nil
}
func (f testFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f testFactory) WaitForCacheSync() {}
func (f testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (f testFactory) DeleteForwarder(string) {}

func makeFactory() dao.Factory {
	return testFactory{}
}
