// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

// RevValues tracks Helm values representations.
type RevValues struct {
	gvr       client.GVR
	inUpdate  int32
	path      string
	rev       string
	query     string
	lines     []string
	allValues bool
	listeners []ResourceViewerListener
	options   ViewerToggleOpts
}

// NewRevValues return a new Helm values resource model.
func NewRevValues(gvr client.GVR, path, rev string) *RevValues {
	return &RevValues{
		gvr:       gvr,
		path:      path,
		rev:       rev,
		allValues: false,
		lines:     getRevValues(path, rev),
	}
}

func getHelmHistDao() *dao.HelmHistory {
	return Registry["helm-history"].DAO.(*dao.HelmHistory)
}

func getRevValues(path, rev string) []string {
	vals, err := getHelmHistDao().GetValues(path, true)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get Helm values")
	}
	return strings.Split(string(vals), "\n")
}

// GVR returns the resource gvr.
func (v *RevValues) GVR() client.GVR {
	return v.gvr
}

// GetPath returns the active resource path.
func (v *RevValues) GetPath() string {
	return v.path
}

// SetOptions toggle model options.
func (v *RevValues) SetOptions(ctx context.Context, opts ViewerToggleOpts) {
	v.options = opts
	if err := v.refresh(ctx); err != nil {
		v.fireResourceFailed(err)
	}
}

// Filter filters the model.
func (v *RevValues) Filter(q string) {
	v.query = q
	v.filterChanged(v.lines)
}

func (v *RevValues) filterChanged(lines []string) {
	v.fireResourceChanged(lines, v.filter(v.query, lines))
}

func (v *RevValues) filter(q string, lines []string) fuzzy.Matches {
	if q == "" {
		return nil
	}
	if f, ok := internal.IsFuzzySelector(q); ok {
		return v.fuzzyFilter(strings.TrimSpace(f), lines)
	}
	return rxFilter(q, lines)
}

func (*RevValues) fuzzyFilter(q string, lines []string) fuzzy.Matches {
	return fuzzy.Find(q, lines)
}

func (v *RevValues) fireResourceChanged(lines []string, matches fuzzy.Matches) {
	for _, l := range v.listeners {
		l.ResourceChanged(lines, matches)
	}
}

func (v *RevValues) fireResourceFailed(err error) {
	for _, l := range v.listeners {
		l.ResourceFailed(err)
	}
}

// ClearFilter clear out the filter.
func (v *RevValues) ClearFilter() {
	v.query = ""
}

// Peek returns the current model data.
func (v *RevValues) Peek() []string {
	return v.lines
}

// Refresh updates model data.
func (v *RevValues) Refresh(ctx context.Context) error {
	return v.refresh(ctx)
}

// Watch watches for Values changes.
func (v *RevValues) Watch(ctx context.Context) error {
	if err := v.refresh(ctx); err != nil {
		return err
	}
	go v.updater(ctx)

	return nil
}

func (v *RevValues) updater(ctx context.Context) {
	defer log.Debug().Msgf("YAML canceled -- %q", v.gvr)

	backOff := NewExpBackOff(ctx, defaultReaderRefreshRate, maxReaderRetryInterval)
	delay := defaultReaderRefreshRate
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			if err := v.refresh(ctx); err != nil {
				v.fireResourceFailed(err)
				if delay = backOff.NextBackOff(); delay == backoff.Stop {
					log.Error().Err(err).Msgf("giving up retrieving chart values")
					return
				}
			} else {
				backOff.Reset()
				delay = defaultReaderRefreshRate
			}
		}
	}
}

func (v *RevValues) refresh(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&v.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return nil
	}
	defer atomic.StoreInt32(&v.inUpdate, 0)

	if err := v.reconcile(ctx); err != nil {
		return err
	}

	return nil
}

func (v *RevValues) reconcile(_ context.Context) error {
	v.fireResourceChanged(v.lines, v.filter(v.query, v.lines))

	return nil
}

// AddListener adds a new model listener.
func (v *RevValues) AddListener(l ResourceViewerListener) {
	v.listeners = append(v.listeners, l)
}

// RemoveListener delete a listener from the list.
func (v *RevValues) RemoveListener(l ResourceViewerListener) {
	victim := -1
	for i, lis := range v.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		v.listeners = append(v.listeners[:victim], v.listeners[victim+1:]...)
	}
}
