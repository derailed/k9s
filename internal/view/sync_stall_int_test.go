// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncStallTrackerBelowDelay(t *testing.T) {
	var s syncStallTracker
	now := time.Now()

	stalled, first := s.stalled("v1/pods@fred", now)
	assert.False(t, stalled)
	assert.False(t, first)

	stalled, first = s.stalled("v1/pods@fred", now.Add(syncStallDelay-time.Second))
	assert.False(t, stalled)
	assert.False(t, first)
}

func TestSyncStallTrackerPastDelay(t *testing.T) {
	var s syncStallTracker
	now := time.Now()

	s.stalled("v1/pods@fred", now)

	stalled, first := s.stalled("v1/pods@fred", now.Add(syncStallDelay))
	assert.True(t, stalled)
	assert.True(t, first, "the first stalled wait must be reported once")

	stalled, first = s.stalled("v1/pods@fred", now.Add(2*syncStallDelay))
	assert.True(t, stalled)
	assert.False(t, first, "a wait already reported must not be reported again")
}

func TestSyncStallTrackerNewKeyRestartsWait(t *testing.T) {
	var s syncStallTracker
	now := time.Now()

	s.stalled("v1/pods@fred", now)

	stalled, first := s.stalled("v1/svc@fred", now.Add(2*syncStallDelay))
	assert.False(t, stalled, "switching resource must start a fresh wait")
	assert.False(t, first)

	stalled, first = s.stalled("v1/svc@fred", now.Add(3*syncStallDelay))
	assert.True(t, stalled)
	assert.True(t, first)
}

func TestSyncStallTrackerNewNamespaceRestartsWait(t *testing.T) {
	var s syncStallTracker
	now := time.Now()

	s.stalled("v1/pods@fred", now)

	stalled, _ := s.stalled("v1/pods@blee", now.Add(2*syncStallDelay))
	assert.False(t, stalled, "switching namespace must start a fresh wait")
}

func TestSyncStallTrackerReset(t *testing.T) {
	var s syncStallTracker
	now := time.Now()

	s.stalled("v1/pods@fred", now)
	stalled, _ := s.stalled("v1/pods@fred", now.Add(syncStallDelay))
	assert.True(t, stalled)

	s.reset()

	stalled, first := s.stalled("v1/pods@fred", now.Add(syncStallDelay))
	assert.False(t, stalled, "a reset tracker must start a fresh wait")
	assert.False(t, first)

	stalled, first = s.stalled("v1/pods@fred", now.Add(2*syncStallDelay))
	assert.True(t, stalled)
	assert.True(t, first, "a reset tracker must report its next stalled wait")
}
