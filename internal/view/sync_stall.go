// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"sync"
	"time"
)

// syncStallDelay is how long an informer cache may stay unsynced before k9s
// stops reporting a plain synchronizing status and calls the wait out instead.
// Some api servers accept a watch-list request, stream the initial events and
// then never send the k8s.io/initial-events-end bookmark the reflector blocks
// on. The cache stays unsynced for good and no error is ever raised, so a wait
// that runs past this delay is reported rather than spun on silently.
const syncStallDelay = 30 * time.Second

// syncStallMsg is the status shown once a sync wait runs past syncStallDelay. It
// names the escape hatch, since the most common cause is an api server that
// accepts watch-list requests without ever ending the initial event stream.
const syncStallMsg = "Still synchronizing %s after %s. If your api server never ends the watch-list stream, restart k9s with KUBE_FEATURE_WatchListClient=false"

// syncStallTracker tracks how long a single resource has been waiting on its
// informer cache, so a wait that is never going to end can be called out.
type syncStallTracker struct {
	mx     sync.Mutex
	key    string
	since  time.Time
	logged bool
}

// stalled records a sync wait for the given key and reports whether that wait
// has now run past syncStallDelay, along with whether this is the first time it
// has. The clock is supplied by the caller to keep this testable. Switching keys
// starts a fresh wait, since the previous resource is no longer on screen.
func (s *syncStallTracker) stalled(key string, now time.Time) (stalled, first bool) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.key != key || s.since.IsZero() {
		s.key, s.since, s.logged = key, now, false
		return false, false
	}
	if now.Sub(s.since) < syncStallDelay {
		return false, false
	}
	first, s.logged = !s.logged, true

	return true, first
}

// reset drops any pending sync wait.
func (s *syncStallTracker) reset() {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.key, s.since, s.logged = "", time.Time{}, false
}
