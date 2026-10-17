// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"
	"time"
)

func TestSpinnerCreation(t *testing.T) {
	s := NewSpinner()
	if s == nil {
		t.Fatal("NewSpinner returned nil")
	}
	if s.IsActive() {
		t.Error("Spinner should not be active initially")
	}
}

func TestSpinnerStartStop(t *testing.T) {
	s := NewSpinner()

	// Start spinner
	s.Start("Loading...")
	if !s.IsActive() {
		t.Error("Spinner should be active after Start")
	}

	// Wait a bit for animation
	time.Sleep(200 * time.Millisecond)

	// Stop spinner
	s.Stop()
	if s.IsActive() {
		t.Error("Spinner should not be active after Stop")
	}
}

func TestSpinnerMessage(t *testing.T) {
	s := NewSpinner()

	if s.IsActive() {
		t.Error("Spinner should not be active initially")
	}

	s.SetMessage("Test message")
	s.Start("Loading...")
	if !s.IsActive() {
		t.Error("Spinner should be active after Start")
	}

	time.Sleep(100 * time.Millisecond)

	s.Stop()
	if s.IsActive() {
		t.Error("Spinner should not be active after Stop")
	}
}

func TestSpinnerMultipleStart(t *testing.T) {
	s := NewSpinner()

	s.Start("Loading...")
	s.Start("Loading again...") // Should not start again

	if !s.IsActive() {
		t.Error("Spinner should still be active")
	}

	s.Stop()
}
