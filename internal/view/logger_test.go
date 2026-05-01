// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestLoggerWriteClampsCount(t *testing.T) {
	a := newTestApp(t)
	l := NewLogger(a)

	p := []byte("hello world\n")
	n, err := l.Write(p)

	assert.NoError(t, err)
	assert.LessOrEqual(t, n, len(p), "Write must not return n > len(p)")
}

func TestLoggerWriteRoundTrip(t *testing.T) {
	a := newTestApp(t)
	l := NewLogger(a)
	l.SetDynamicColors(true)

	lines := []string{
		"plain text line\n",
		"[gray::b]timestamp[-::-] message\n",
		"line with [[escaped[[ brackets]]\n",
	}
	for _, line := range lines {
		n, err := l.Write([]byte(line))
		assert.NoError(t, err)
		assert.LessOrEqual(t, n, len(line))
	}

	text := l.GetText(false)
	assert.NotEmpty(t, text)
}

func TestLoggerWriteEmpty(t *testing.T) {
	a := newTestApp(t)
	l := NewLogger(a)

	n, err := l.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestLoggerEmbeddedTextViewWrite(t *testing.T) {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)

	p := []byte("[red]hello[-]\n")
	n, _ := tv.Write(p)

	// This documents the tview bug: n can exceed len(p) because
	// tview counts bytes written to its internal buffer after tag processing.
	// Our Logger.Write wrapper fixes this.
	_ = n
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	return NewApp(mock.NewMockConfig(t))
}
