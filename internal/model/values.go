// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/sahilm/fuzzy"
)

// Values tracks Helm values representations.
type Values struct {
	factory   dao.Factory
	gvr       *client.GVR
	inUpdate  int32
	path      string
	query     string
	lines     []string
	allValues bool
	listeners []ResourceViewerListener
	options   ViewerToggleOpts
}

// NewValues return a new Helm values resource model.
func NewValues(gvr *client.GVR, path string) *Values {
	return &Values{
		gvr:       gvr,
		path:      path,
		allValues: false,
	}
}

// Init initializes the model.
func (v *Values) Init(f dao.Factory) error {
	v.factory = f

	var err error
	v.lines, err = v.getValues()

	return err
}

func (v *Values) getValues() ([]string, error) {
	accessor, err := dao.AccessorFor(v.factory, v.gvr)
	if err != nil {
		return nil, err
	}

	valuer, ok := accessor.(dao.Valuer)
	if !ok {
		return nil, fmt.Errorf("resource %s is not Valuer", v.gvr)
	}

	values, err := valuer.GetValues(v.path, v.allValues)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(values), "\n"), nil
}

// GVR returns the resource gvr.
func (v *Values) GVR() *client.GVR {
	return v.gvr
}

// ToggleValues toggles between user supplied values and computed values.
func (v *Values) ToggleValues() error {
	v.allValues = !v.allValues

	lines, err := v.getValues()
	if err != nil {
		return err
	}

	v.lines = lines
	return nil
}

// GetPath returns the active resource path.
func (v *Values) GetPath() string {
	return v.path
}

// SetOptions toggle model options.
func (v *Values) SetOptions(ctx context.Context, opts ViewerToggleOpts) {
	v.options = opts
	if err := v.refresh(ctx); err != nil {
		v.fireResourceFailed(err)
	}
}

// Filter filters the model.
func (v *Values) Filter(q string) {
	v.query = q
	v.filterChanged(v.lines)
}

func (v *Values) filterChanged(lines []string) {
	v.fireResourceChanged(lines, v.filter(v.query, lines))
}

func (v *Values) filter(q string, lines []string) fuzzy.Matches {
	if q == "" {
		return nil
	}
	if f, ok := internal.IsFuzzySelector(q); ok {
		return v.fuzzyFilter(strings.TrimSpace(f), lines)
	}
	return rxFilter(q, lines)
}

func (*Values) fuzzyFilter(q string, lines []string) fuzzy.Matches {
	return fuzzy.Find(q, lines)
}

func (v *Values) fireResourceChanged(lines []string, matches fuzzy.Matches) {
	for _, l := range v.listeners {
		l.ResourceChanged(lines, matches)
	}
}

func (v *Values) fireResourceFailed(err error) {
	for _, l := range v.listeners {
		l.ResourceFailed(err)
	}
}

// ClearFilter clear out the filter.
func (v *Values) ClearFilter() {
	v.query = ""
}

// Peek returns the current model data.
func (v *Values) Peek() []string {
	return v.lines
}

// Refresh updates model data.
func (v *Values) Refresh(ctx context.Context) error {
	return v.refresh(ctx)
}

// Watch watches for Values changes.
func (v *Values) Watch(ctx context.Context) error {
	if err := v.refresh(ctx); err != nil {
		return err
	}
	go v.updater(ctx)

	return nil
}

func (v *Values) updater(ctx context.Context) {
	defer slog.Debug("YAML canceled", slogs.GVR, v.gvr)

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
					slog.Error("Giving up retrieving chart values", slogs.Error, err)
					return
				}
			} else {
				backOff.Reset()
				delay = defaultReaderRefreshRate
			}
		}
	}
}

func (v *Values) refresh(context.Context) error {
	if !atomic.CompareAndSwapInt32(&v.inUpdate, 0, 1) {
		slog.Debug("Dropping update...")
		return fmt.Errorf("reconcile in progress. Dropping update")
	}
	defer atomic.StoreInt32(&v.inUpdate, 0)

	v.reconcile()

	return nil
}

func (v *Values) reconcile() {
	v.fireResourceChanged(v.lines, v.filter(v.query, v.lines))
}

// AddListener adds a new model listener.
func (v *Values) AddListener(l ResourceViewerListener) {
	v.listeners = append(v.listeners, l)
}

// RemoveListener delete a listener from the list.
func (v *Values) RemoveListener(l ResourceViewerListener) {
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
