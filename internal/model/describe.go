// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"reflect"
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

// Describe tracks describable resources.
type Describe struct {
	gvr         client.GVR
	inUpdate    int32
	path        string
	query       string
	lines       []string
	refreshRate time.Duration
	listeners   []ResourceViewerListener
	decode      bool
}

// NewDescribe returns a new describe resource model.
func NewDescribe(gvr client.GVR, path string) *Describe {
	return &Describe{
		gvr:         gvr,
		path:        path,
		refreshRate: defaultReaderRefreshRate,
	}
}

// GVR returns the resource gvr.
func (d *Describe) GVR() client.GVR {
	return d.gvr
}

// GetPath returns the active resource path.
func (d *Describe) GetPath() string {
	return d.path
}

// SetOptions toggle model options.
func (d *Describe) SetOptions(context.Context, ViewerToggleOpts) {}

// Filter filters the model.
func (d *Describe) Filter(q string) {
	d.query = q
	d.filterChanged(d.lines)
}

func (d *Describe) filterChanged(lines []string) {
	d.fireResourceChanged(lines, d.filter(d.query, lines))
}

func (d *Describe) filter(q string, lines []string) fuzzy.Matches {
	if q == "" {
		return nil
	}
	if f, ok := internal.IsFuzzySelector(q); ok {
		return d.fuzzyFilter(strings.TrimSpace(f), lines)
	}
	return rxFilter(q, lines)
}

func (*Describe) fuzzyFilter(q string, lines []string) fuzzy.Matches {
	return fuzzy.Find(q, lines)
}

func (d *Describe) fireResourceChanged(lines []string, matches fuzzy.Matches) {
	for _, l := range d.listeners {
		l.ResourceChanged(lines, matches)
	}
}

func (d *Describe) fireResourceFailed(err error) {
	for _, l := range d.listeners {
		l.ResourceFailed(err)
	}
}

// ClearFilter clear out the filter.
func (d *Describe) ClearFilter() {
}

// Peek returns current model state.
func (d *Describe) Peek() []string {
	return d.lines
}

// Refresh updates model data.
func (d *Describe) Refresh(ctx context.Context) error {
	return d.refresh(ctx)
}

// Watch watches for describe data changes.
func (d *Describe) Watch(ctx context.Context) error {
	if err := d.refresh(ctx); err != nil {
		return err
	}
	go d.updater(ctx)
	return nil
}

func (d *Describe) updater(ctx context.Context) {
	defer log.Debug().Msgf("Describe canceled -- %q", d.gvr)

	backOff := NewExpBackOff(ctx, defaultReaderRefreshRate, maxReaderRetryInterval)
	delay := defaultReaderRefreshRate
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			if err := d.refresh(ctx); err != nil {
				d.fireResourceFailed(err)
				if delay = backOff.NextBackOff(); delay == backoff.Stop {
					log.Error().Err(err).Msgf("Describe gave up!")
					return
				}
			} else {
				backOff.Reset()
				delay = defaultReaderRefreshRate
			}
		}
	}
}

func (d *Describe) refresh(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&d.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return nil
	}
	defer atomic.StoreInt32(&d.inUpdate, 0)

	if err := d.reconcile(ctx); err != nil {
		log.Error().Err(err).Msgf("reconcile failed %q", d.gvr)
		d.fireResourceFailed(err)
		return err
	}

	return nil
}

func (d *Describe) reconcile(ctx context.Context) error {
	s, err := d.describe(ctx, d.gvr, d.path)
	if err != nil {
		return err
	}
	lines := strings.Split(s, "\n")
	if reflect.DeepEqual(lines, d.lines) {
		return nil
	}
	d.lines = lines
	d.fireResourceChanged(d.lines, d.filter(d.query, d.lines))

	return nil
}

// Describe describes a given resource.
func (d *Describe) describe(ctx context.Context, gvr client.GVR, path string) (string, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("Describe model elapsed: %v", time.Since(t))
	}(time.Now())

	meta, err := getMeta(ctx, gvr)
	if err != nil {
		return "", err
	}
	desc, ok := meta.DAO.(dao.Describer)
	if !ok {
		return "", fmt.Errorf("no describer for %q", meta.DAO.GVR())
	}

	if desc, ok := meta.DAO.(*dao.Secret); ok {
		desc.SetDecode(d.decode)
	}

	return desc.Describe(path)
}

// AddListener adds a new model listener.
func (d *Describe) AddListener(l ResourceViewerListener) {
	d.listeners = append(d.listeners, l)
}

// RemoveListener delete a listener from the list.
func (d *Describe) RemoveListener(l ResourceViewerListener) {
	victim := -1
	for i, lis := range d.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		d.listeners = append(d.listeners[:victim], d.listeners[victim+1:]...)
	}
}

// Toggle toggles the decode flag.
func (d *Describe) Toggle() {
	d.decode = !d.decode
}
