// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestScreenGuard_PollEvent_Normal(t *testing.T) {
	sim := tcell.NewSimulationScreen("")
	assert.NoError(t, sim.Init())
	defer sim.Fini()

	var deadCalled atomic.Bool
	guard := NewScreenGuard(sim, func(err error) {
		deadCalled.Store(true)
	})

	// Post normal key event
	keyEv := tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone)
	assert.NoError(t, sim.PostEvent(keyEv))

	ev := guard.PollEvent()
	assert.Equal(t, keyEv, ev)
	assert.False(t, deadCalled.Load())
}

func TestScreenGuard_PollEvent_EventError(t *testing.T) {
	sim := tcell.NewSimulationScreen("")
	assert.NoError(t, sim.Init())
	defer sim.Fini()

	var deadCalled atomic.Bool
	deadCh := make(chan struct{})

	testErr := errors.New("input/output error")
	guard := NewScreenGuard(sim, func(err error) {
		deadCalled.Store(true)
		assert.Equal(t, testErr.Error(), err.Error())
		close(deadCh)
	})

	errEv := tcell.NewEventError(testErr)
	assert.NoError(t, sim.PostEvent(errEv))

	ev := guard.PollEvent()
	assert.Equal(t, errEv, ev)

	select {
	case <-deadCh:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for onTerminalDead callback")
	}

	assert.True(t, deadCalled.Load())
}
