// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/derailed/tview"
)

// Spinner represents a loading spinner with a message.
type Spinner struct {
	*tview.TextView

	message   string
	frames    []string
	frameIdx  int
	active    bool
	cancel    chan struct{}
	mx        sync.Mutex
	fgColor   string
}

// NewSpinner returns a new spinner widget.
func NewSpinner() *Spinner {
	s := &Spinner{
		TextView: tview.NewTextView(),
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		cancel:   make(chan struct{}),
		fgColor:  "green",
	}
	s.SetTextAlign(tview.AlignCenter)
	s.SetDynamicColors(true)
	return s
}

// SetMessage sets the spinner message.
func (s *Spinner) SetMessage(msg string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.message = msg
}

// SetColor sets the spinner foreground color.
func (s *Spinner) SetColor(color string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.fgColor = color
}

// Start begins the spinner animation.
func (s *Spinner) Start(message string) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.active {
		return
	}
	s.active = true
	s.message = message
	s.cancel = make(chan struct{})
	go s.animate()
}

// Stop stops the spinner animation.
func (s *Spinner) Stop() {
	s.mx.Lock()
	defer s.mx.Unlock()

	if !s.active {
		return
	}
	s.active = false
	close(s.cancel)
}

// IsActive returns whether the spinner is currently animating.
func (s *Spinner) IsActive() bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.active
}

func (s *Spinner) animate() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.cancel:
			return
		case <-ticker.C:
			s.mx.Lock()
			frame := s.frames[s.frameIdx]
			msg := s.message
			color := s.fgColor
			s.frameIdx = (s.frameIdx + 1) % len(s.frames)
			s.mx.Unlock()

			display := fmt.Sprintf("[%s::b]%s [-::]%s", color, frame, msg)
			s.SetText(display)
		}
	}
}
