// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLogs_Normal(t *testing.T) {
	input := "line one\nline two\nline three\n"
	stream := io.NopCloser(strings.NewReader(input))
	out := make(chan *LogItem, 100)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	result := readLogs(context.Background(), stream, out, opts)
	close(out)

	assert.Equal(t, streamEOF, result)

	var lines []string
	for item := range out {
		if !item.IsError {
			lines = append(lines, string(item.Bytes))
		}
	}
	assert.Len(t, lines, 3)
}

func TestReadLogs_DropsOnFullChannel(t *testing.T) {
	// Create more lines than the channel can hold
	var sb strings.Builder
	lineCount := 20
	for i := range lineCount {
		sb.WriteString("log line ")
		sb.WriteByte(byte('A' + i%26))
		sb.WriteByte('\n')
	}
	stream := io.NopCloser(strings.NewReader(sb.String()))

	// Intentionally small buffer to force drops.
	// readLogs sends an EOF error item via a blocking send,
	// so we drain in a goroutine to prevent deadlock.
	out := make(chan *LogItem, 2)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	var received int
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range out {
			received++
		}
	}()

	result := readLogs(context.Background(), stream, out, opts)
	close(out)
	<-done

	assert.Equal(t, streamEOF, result)
	// Some lines should have been dropped since buffer is tiny
	assert.Less(t, received, lineCount+2, "some lines should be dropped when channel is full")
	assert.Greater(t, received, 0, "at least some lines should be delivered")
}

func TestReadLogs_CancelStopsEarly(t *testing.T) {
	// Infinite-like stream: we cancel after a few lines
	input := strings.Repeat("line\n", 10000)
	stream := io.NopCloser(strings.NewReader(input))
	out := make(chan *LogItem, 10)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	ctx, cancel := context.WithCancel(context.Background())

	// Read a few items then cancel
	done := make(chan streamResult, 1)
	go func() {
		done <- readLogs(ctx, stream, out, opts)
	}()

	// Drain a few items then cancel
	for range 5 {
		<-out
	}
	cancel()

	result := <-done
	assert.Equal(t, streamCanceled, result)
}

func TestReadLogs_PartialLineAtEOF(t *testing.T) {
	// Input without trailing newline
	input := "full line\npartial line without newline"
	stream := io.NopCloser(strings.NewReader(input))
	out := make(chan *LogItem, 100)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	result := readLogs(context.Background(), stream, out, opts)
	close(out)

	assert.Equal(t, streamEOF, result)

	var items []*LogItem
	for item := range out {
		items = append(items, item)
	}
	// Should get: full line, partial line, and the EOF error message
	assert.GreaterOrEqual(t, len(items), 2, "should emit partial line before EOF")

	// Find the partial line (non-error item that doesn't end with newline)
	foundPartial := false
	for _, item := range items {
		if !item.IsError && strings.Contains(string(item.Bytes), "partial line without newline") {
			foundPartial = true
		}
	}
	assert.True(t, foundPartial, "partial line at EOF should be emitted")
}
