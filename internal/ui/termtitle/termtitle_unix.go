// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

//go:build !windows

package termtitle

import (
	"errors"
	"io"
	"os"
)

const (
	// Push the window and icon titles on the terminal's title stack (XTWINOPS 22).
	pushTitle = "\x1b[22;0t"
	// Pop the window and icon titles from the terminal's title stack (XTWINOPS 23).
	popTitle = "\x1b[23;0t"
)

// ttyBackend talks to the controlling terminal with xterm escape sequences.
// Terminals that do not support the title stack ignore the push and pop.
type ttyBackend struct {
	out io.WriteCloser
}

func newBackend() backend {
	return &ttyBackend{}
}

func (b *ttyBackend) open() error {
	if b.out != nil {
		return nil
	}
	f, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	b.out = f

	return nil
}

func (b *ttyBackend) write(s string) error {
	if b.out == nil {
		return errors.New("terminal not open")
	}
	_, err := io.WriteString(b.out, s)

	return err
}

func (b *ttyBackend) save() error {
	if err := b.open(); err != nil {
		return err
	}

	return b.write(pushTitle)
}

func (b *ttyBackend) set(title string) error {
	// OSC 0 sets both the window and the icon (tab) title.
	return b.write("\x1b]0;" + title + "\x07")
}

func (b *ttyBackend) restore() error {
	if b.out == nil {
		return nil
	}
	err := b.write(popTitle)
	_ = b.out.Close()
	b.out = nil

	return err
}
