// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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

	c := make(dao.LogChan, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.updateLogs(ctx, c)

	for i := 0; i < 2*size; i++ {
		c <- dao.NewLogItemFromString("line" + strconv.Itoa(i))
	}

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
	item := dao.NewLogItem([]byte("\033[0;38m2018-12-14T10:36:43.326972-07:00 \033[0;32mblee line"))

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c <- item
	}
	close(c)
}

// Helpers...

func makeLogOpts(count int) *dao.LogOptions {
	return &dao.LogOptions{
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

func (t *mockLogView) LogChanged(ll [][]byte) {
	t.count += len(ll)
}
func (t *mockLogView) LogStop()            {}
func (t *mockLogView) LogCanceled()        {}
func (t *mockLogView) LogResume()          {}
func (t *mockLogView) LogCleared()         {}
func (t *mockLogView) LogFailed(err error) {}
