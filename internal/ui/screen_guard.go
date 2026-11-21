// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"log/slog"
	"sync"

	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tcell/v2"
)

// ScreenGuard wraps a tcell.Screen and intercepts terminal error events
// (such as EOF or EIO caused by terminal closure/disconnection) to ensure
// K9s exits gracefully instead of running indefinitely as an orphaned process.
type ScreenGuard struct {
	tcell.Screen
	onTerminalDead func(err error)
	once           sync.Once
}

// NewScreenGuard returns a new ScreenGuard wrapping the provided screen.
func NewScreenGuard(screen tcell.Screen, onTerminalDead func(err error)) *ScreenGuard {
	return &ScreenGuard{
		Screen:         screen,
		onTerminalDead: onTerminalDead,
	}
}

// PollEvent waits for events to arrive. If a *tcell.EventError is encountered,
// it indicates that reading from the terminal failed (e.g. terminal closed or SSH
// connection dropped), so onTerminalDead is invoked to initiate graceful termination.
func (s *ScreenGuard) PollEvent() tcell.Event {
	ev := s.Screen.PollEvent()
	if errEv, ok := ev.(*tcell.EventError); ok {
		s.once.Do(func() {
			slog.Warn("Terminal connection lost, exiting...", slogs.Error, errEv)
			if s.onTerminalDead != nil {
				go s.onTerminalDead(errEv)
			}
		})
	}
	return ev
}
