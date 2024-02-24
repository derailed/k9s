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
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

// ManagedFieldsOpts tracks managed fields.
const ManagedFieldsOpts = "ManagedFields"

// YAML tracks yaml resource representations.
type YAML struct {
	gvr       client.GVR
	inUpdate  int32
	path      string
	query     string
	lines     []string
	listeners []ResourceViewerListener
	options   ViewerToggleOpts
}

// NewYAML return a new yaml resource model.
func NewYAML(gvr client.GVR, path string) *YAML {
	return &YAML{
		gvr:  gvr,
		path: path,
	}
}

// GVR returns the resource gvr.
func (y *YAML) GVR() client.GVR {
	return y.gvr
}

// GetPath returns the active resource path.
func (y *YAML) GetPath() string {
	return y.path
}

// SetOptions toggle model options.
func (y *YAML) SetOptions(ctx context.Context, opts ViewerToggleOpts) {
	y.options = opts
	if err := y.refresh(ctx); err != nil {
		y.fireResourceFailed(err)
	}
}

// Filter filters the model.
func (y *YAML) Filter(q string) {
	y.query = q
	y.filterChanged(y.lines)
}

func (y *YAML) filterChanged(lines []string) {
	y.fireResourceChanged(lines, y.filter(y.query, lines))
}

func (y *YAML) filter(q string, lines []string) fuzzy.Matches {
	if q == "" {
		return nil
	}
	if f, ok := internal.IsFuzzySelector(q); ok {
		return y.fuzzyFilter(strings.TrimSpace(f), lines)
	}
	return rxFilter(q, lines)
}

func (*YAML) fuzzyFilter(q string, lines []string) fuzzy.Matches {
	return fuzzy.Find(q, lines)
}

func (y *YAML) fireResourceChanged(lines []string, matches fuzzy.Matches) {
	for _, l := range y.listeners {
		l.ResourceChanged(lines, matches)
	}
}

func (y *YAML) fireResourceFailed(err error) {
	for _, l := range y.listeners {
		l.ResourceFailed(err)
	}
}

// ClearFilter clear out the filter.
func (y *YAML) ClearFilter() {
	y.query = ""
}

// Peek returns the current model data.
func (y *YAML) Peek() []string {
	return y.lines
}

// Refresh updates model data.
func (y *YAML) Refresh(ctx context.Context) error {
	return y.refresh(ctx)
}

// Watch watches for YAML changes.
func (y *YAML) Watch(ctx context.Context) error {
	if err := y.refresh(ctx); err != nil {
		return err
	}
	go y.updater(ctx)

	return nil
}

func (y *YAML) updater(ctx context.Context) {
	defer log.Debug().Msgf("YAML canceled -- %q", y.gvr)

	backOff := NewExpBackOff(ctx, defaultReaderRefreshRate, maxReaderRetryInterval)
	delay := defaultReaderRefreshRate
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
			if err := y.refresh(ctx); err != nil {
				y.fireResourceFailed(err)
				if delay = backOff.NextBackOff(); delay == backoff.Stop {
					log.Error().Err(err).Msgf("YAML gave up!")
					return
				}
			} else {
				backOff.Reset()
				delay = defaultReaderRefreshRate
			}
		}
	}
}

func (y *YAML) refresh(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&y.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return nil
	}
	defer atomic.StoreInt32(&y.inUpdate, 0)

	if err := y.reconcile(ctx); err != nil {
		return err
	}

	return nil
}

func (y *YAML) reconcile(ctx context.Context) error {
	s, err := y.ToYAML(ctx, y.gvr, y.path, y.options[ManagedFieldsOpts])
	if err != nil {
		return err
	}
	lines := strings.Split(s, "\n")
	if reflect.DeepEqual(lines, y.lines) {
		return nil
	}
	y.lines = lines
	y.fireResourceChanged(y.lines, y.filter(y.query, y.lines))

	return nil
}

// AddListener adds a new model listener.
func (y *YAML) AddListener(l ResourceViewerListener) {
	y.listeners = append(y.listeners, l)
}

// RemoveListener delete a listener from the list.
func (y *YAML) RemoveListener(l ResourceViewerListener) {
	victim := -1
	for i, lis := range y.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		y.listeners = append(y.listeners[:victim], y.listeners[victim+1:]...)
	}
}

// ToYAML returns a resource yaml.
func (y *YAML) ToYAML(ctx context.Context, gvr client.GVR, path string, showManaged bool) (string, error) {
	meta, err := getMeta(ctx, gvr)
	if err != nil {
		return "", err
	}

	desc, ok := meta.DAO.(dao.Describer)
	if !ok {
		return "", fmt.Errorf("no describer for %q", meta.DAO.GVR())
	}

	return desc.ToYAML(path, showManaged)
}

func getMeta(ctx context.Context, gvr client.GVR) (ResourceMeta, error) {
	meta := resourceMeta(gvr)
	factory, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return ResourceMeta{}, fmt.Errorf("expected Factory in context but got %T", ctx.Value(internal.KeyFactory))
	}
	meta.DAO.Init(factory, gvr)

	return meta, nil
}

func resourceMeta(gvr client.GVR) ResourceMeta {
	meta, ok := Registry[gvr.String()]
	if !ok {
		meta = ResourceMeta{
			DAO:      &dao.Table{},
			Renderer: &render.Generic{},
		}
	}
	if meta.DAO == nil {
		meta.DAO = &dao.Resource{}
	}

	return meta
}
