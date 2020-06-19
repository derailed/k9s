package model

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestUpdateLogs(t *testing.T) {
	size := 100
	m := NewLog(client.NewGVR("fred"), makeLogOpts(size), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newMockLogView()
	m.AddListener(v)

	c := make(dao.LogChan)
	go func() {
		m.updateLogs(context.Background(), c)
	}()

	for i := 0; i < 2*size; i++ {
		c <- dao.NewLogItemFromString("line" + strconv.Itoa(i))
	}
	close(c)

	time.Sleep(2 * time.Second)
	assert.Equal(t, size, v.count)
}

func BenchmarkUpdateLogs(b *testing.B) {
	size := 100
	m := NewLog(client.NewGVR("fred"), makeLogOpts(size), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newMockLogView()
	m.AddListener(v)

	c := make(dao.LogChan)
	go func() {
		m.updateLogs(context.Background(), c)
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c <- dao.NewLogItemFromString("line" + strconv.Itoa(n))
	}
	close(c)
}

// Helpers...

func makeLogOpts(count int) dao.LogOptions {
	return dao.LogOptions{
		Path:      "fred",
		Container: "blee",
		Lines:     int64(count),
	}
}

type mockLogView struct {
	count int
}

func newMockLogView() *mockLogView {
	return &mockLogView{}
}

func (t *mockLogView) LogChanged(d dao.LogItems) {
	t.count += len(d.Lines())
}
func (t *mockLogView) LogCleared()         {}
func (t *mockLogView) LogFailed(err error) {}
