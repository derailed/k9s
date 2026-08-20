// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadLogs_Normal(t *testing.T) {
	input := "line one\nline two\nline three\n"
	stream := io.NopCloser(strings.NewReader(input))
	out := make(chan *LogItem, 100)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	done := make(chan streamResult, 1)
	go func() {
		done <- readLogs(context.Background(), stream, out, opts)
	}()

	result := <-done
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

func TestReadLogs_CancelStopsEarly(t *testing.T) {
	pr, pw := io.Pipe()
	out := make(chan *LogItem, 100)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan streamResult, 1)
	go func() {
		done <- readLogs(ctx, pr, out, opts)
	}()

	// Write some lines
	for range 5 {
		_, _ = pw.Write([]byte("line\n"))
	}
	// Drain
	for range 5 {
		<-out
	}

	cancel()
	result := <-done
	pw.Close()
	assert.Equal(t, streamCanceled, result)
}

func TestReadLogs_PartialLineAtEOF(t *testing.T) {
	input := "full line\npartial line without newline"
	stream := io.NopCloser(strings.NewReader(input))
	out := make(chan *LogItem, 100)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	done := make(chan streamResult, 1)
	go func() {
		done <- readLogs(context.Background(), stream, out, opts)
	}()

	result := <-done
	close(out)

	assert.Equal(t, streamEOF, result)

	var items []*LogItem
	for item := range out {
		items = append(items, item)
	}
	assert.GreaterOrEqual(t, len(items), 2, "should emit partial line before EOF")

	foundPartial := false
	for _, item := range items {
		if !item.IsError && strings.Contains(string(item.Bytes), "partial line without newline") {
			foundPartial = true
		}
	}
	assert.True(t, foundPartial, "partial line at EOF should be emitted")
}

func TestReadLogs_PartialLineTimeout(t *testing.T) {
	pr, pw := io.Pipe()

	out := make(chan *LogItem, 100)
	opts := &LogOptions{Path: "ns/pod", Container: "c1"}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan streamResult, 1)
	go func() {
		done <- readLogs(ctx, pr, out, opts)
	}()

	// Write partial data (no newline)
	_, _ = pw.Write([]byte("progress: 50%"))

	// Wait for the partial line timer to fire (partialLineTimeout = 3s)
	var item *LogItem
	select {
	case item = <-out:
	case <-time.After(5 * time.Second):
		t.Fatal("expected partial line to be flushed after timeout")
	}

	assert.Contains(t, string(item.Bytes), "progress: 50%")

	cancel()
	pw.Close()
	<-done
}
