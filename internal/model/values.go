package model

import (
	"context"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

// YAML tracks yaml resource representations.
type Values struct {
	gvr       client.GVR
	inUpdate  int32
	path      string
	query     string
	lines     []string
	listeners []ResourceViewerListener
	options   ViewerToggleOpts
}

// NewYAML return a new yaml resource model.
func NewValues(gvr client.GVR, path string, vals string) *Values {
	return &Values{
		gvr:   gvr,
		path:  path,
		lines: strings.Split(vals, "\n"),
	}
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
	if dao.IsFuzzySelector(q) {
		return v.fuzzyFilter(strings.TrimSpace(q[2:]), lines)
	}
	return v.rxFilter(q, lines)
}

func (*Values) fuzzyFilter(q string, lines []string) fuzzy.Matches {
	return fuzzy.Find(q, lines)
}

func (*Values) rxFilter(q string, lines []string) fuzzy.Matches {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil
	}
	matches := make(fuzzy.Matches, 0, len(lines))
	for i, l := range lines {
		if loc := rx.FindStringIndex(l); len(loc) == 2 {
			matches = append(matches, fuzzy.Match{Str: q, Index: i, MatchedIndexes: loc})
		}
	}

	return matches
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

// Watch watches for YAML changes.
func (v *Values) Watch(ctx context.Context) error {
	if err := v.refresh(ctx); err != nil {
		return err
	}
	go v.updater(ctx)

	return nil
}

func (v *Values) updater(ctx context.Context) {
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

func (v *Values) refresh(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&v.inUpdate, 0, 1) {
		log.Debug().Msgf("Dropping update...")
		return nil
	}
	defer atomic.StoreInt32(&v.inUpdate, 0)

	return nil
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
