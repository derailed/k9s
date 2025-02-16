// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog/log"
)

// LogsListener represents a log model listener.
type LogsListener interface {
	// LogChanged notifies the model changed.
	LogChanged([][]byte)

	// LogCleared indicates logs are cleared.
	LogCleared()

	// LogFailed indicates a log failure.
	LogFailed(error)

	// LogStop indicates logging was canceled.
	LogStop()

	// LogResume indicates logging has resumed.
	LogResume()

	// LogCanceled indicates no more logs will come.
	LogCanceled()
}

// Log represents a resource logger.
type Log struct {
	factory      dao.Factory
	lines        *dao.LogItems
	listeners    []LogsListener
	gvr          client.GVR
	logOptions   *dao.LogOptions
	cancelFn     context.CancelFunc
	mx           sync.RWMutex
	filter       string
	lastSent     int
	flushTimeout time.Duration
}

// NewLog returns a new model.
func NewLog(gvr client.GVR, opts *dao.LogOptions, flushTimeout time.Duration) *Log {
	return &Log{
		gvr:          gvr,
		logOptions:   opts,
		lines:        dao.NewLogItems(),
		flushTimeout: flushTimeout,
	}
}

func (l *Log) GVR() client.GVR {
	return l.gvr
}

func (l *Log) LogOptions() *dao.LogOptions {
	return l.logOptions
}

// SinceSeconds returns since seconds option.
func (l *Log) SinceSeconds() int64 {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.logOptions.SinceSeconds
}

// IsHead returns log head option.
func (l *Log) IsHead() bool {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.logOptions.Head
}

// ToggleShowTimestamp toggles to logs timestamps.
func (l *Log) ToggleShowTimestamp(b bool) {
	l.logOptions.ShowTimestamp = b
	l.Refresh()
}

func (l *Log) Head(ctx context.Context) {
	l.mx.Lock()
	{
		l.logOptions.Head = true
	}
	l.mx.Unlock()
	l.Restart(ctx)
}

// SetSinceSeconds sets the logs retrieval time.
func (l *Log) SetSinceSeconds(ctx context.Context, i int64) {
	l.logOptions.SinceSeconds, l.logOptions.Head = i, false
	l.Restart(ctx)
}

// Configure sets logger configuration.
func (l *Log) Configure(opts config.Logger) {
	l.logOptions.Lines = int64(opts.TailCount)
	l.logOptions.SinceSeconds = opts.SinceSeconds
}

// GetPath returns resource path.
func (l *Log) GetPath() string {
	return l.logOptions.Path
}

// GetContainer returns the resource container if any or "" otherwise.
func (l *Log) GetContainer() string {
	return l.logOptions.Container
}

// HasDefaultContainer returns true if the pod has a default container, false otherwise.
func (l *Log) HasDefaultContainer() bool {
	return l.logOptions.DefaultContainer != ""
}

// Init initializes the model.
func (l *Log) Init(f dao.Factory) {
	l.factory = f
}

// Clear the logs.
func (l *Log) Clear() {
	l.mx.Lock()
	{
		l.lines.Clear()
		l.lastSent = 0
	}
	l.mx.Unlock()

	l.fireLogCleared()
}

// Refresh refreshes the logs.
func (l *Log) Refresh() {
	l.fireLogCleared()
	ll := make([][]byte, l.lines.Len())
	l.lines.Render(0, l.logOptions.ShowTimestamp, ll)
	l.fireLogChanged(ll)
}

// Restart restarts the logger.
func (l *Log) Restart(ctx context.Context) {
	l.Stop()
	l.Clear()
	l.fireLogResume()
	l.Start(ctx)
}

// Start starts logging.
func (l *Log) Start(ctx context.Context) {
	if err := l.load(ctx); err != nil {
		log.Error().Err(err).Msgf("Tail logs failed!")
		l.fireLogError(err)
	}
}

// Stop terminates logging.
func (l *Log) Stop() {
	l.cancel()
}

// Set sets the log lines (for testing only!)
func (l *Log) Set(lines *dao.LogItems) {
	l.mx.Lock()
	{
		l.lines.Merge(lines)
	}
	l.mx.Unlock()

	l.fireLogCleared()
	ll := make([][]byte, l.lines.Len())
	l.lines.Render(0, l.logOptions.ShowTimestamp, ll)
	l.fireLogChanged(ll)
}

// ClearFilter resets the log filter if any.
func (l *Log) ClearFilter() {
	l.mx.Lock()
	{
		l.filter = ""
	}
	l.mx.Unlock()

	l.fireLogCleared()
	ll := make([][]byte, l.lines.Len())
	l.lines.Render(0, l.logOptions.ShowTimestamp, ll)
	l.fireLogChanged(ll)
}

// Filter filters the model using either fuzzy or regexp.
func (l *Log) Filter(q string) {
	l.mx.Lock()
	{
		l.filter = q
	}
	l.mx.Unlock()

	l.fireLogCleared()
	l.fireLogBuffChanged(0)
}

func (l *Log) cancel() {
	l.mx.Lock()
	defer l.mx.Unlock()
	if l.cancelFn != nil {
		l.cancelFn()
		log.Debug().Msgf("!!! LOG-MODEL CANCELED !!!")
		l.cancelFn = nil
	}
}

func (l *Log) load(ctx context.Context) error {
	accessor, err := dao.AccessorFor(l.factory, l.gvr)
	if err != nil {
		return err
	}
	loggable, ok := accessor.(dao.Loggable)
	if !ok {
		return fmt.Errorf("Resource %s is not Loggable", l.gvr)
	}

	l.cancel()
	ctx = context.WithValue(ctx, internal.KeyFactory, l.factory)
	ctx, l.cancelFn = context.WithCancel(ctx)

	cc, err := loggable.TailLogs(ctx, l.logOptions)
	if err != nil {
		log.Error().Err(err).Msgf("Tail logs failed")
		l.cancel()
		l.fireLogError(err)
	}
	for _, c := range cc {
		go l.updateLogs(ctx, c)
	}

	return nil
}

// Append adds a log line.
func (l *Log) Append(line *dao.LogItem) {
	if line == nil || line.IsEmpty() {
		return
	}
	l.mx.Lock()
	defer l.mx.Unlock()
	l.logOptions.SinceTime = line.GetTimestamp()
	if l.lines.Len() < int(l.logOptions.Lines) {
		l.lines.Add(line)
		return
	}
	l.lines.Shift(line)
	l.lastSent--
	if l.lastSent < 0 {
		l.lastSent = 0
	}
}

// Notify fires of notifications to the listeners.
func (l *Log) Notify() {
	l.mx.Lock()
	defer l.mx.Unlock()

	if l.lastSent < l.lines.Len() {
		l.fireLogBuffChanged(l.lastSent)
		l.lastSent = l.lines.Len()
	}
}

// ToggleAllContainers toggles to show all containers logs.
func (l *Log) ToggleAllContainers(ctx context.Context) {
	l.logOptions.ToggleAllContainers()
	l.Restart(ctx)
}

func (l *Log) updateLogs(ctx context.Context, c dao.LogChan) {
	defer log.Debug().Msgf("<<< LOG-MODEL UPDATER DONE %s!!!!", l.logOptions.Info())
	log.Debug().Msgf(">>> START LOG-MODEL UPDATER %s", l.logOptions.Info())
	for {
		select {
		case item, ok := <-c:
			if !ok {
				l.Append(item)
				l.Notify()
				return
			}
			if item == dao.ItemEOF {
				l.fireCanceled()
				return
			}
			l.Append(item)
			var overflow bool
			l.mx.RLock()
			{
				overflow = int64(l.lines.Len()-l.lastSent) > l.logOptions.Lines
			}
			l.mx.RUnlock()
			if overflow {
				l.Notify()
			}
		case <-time.After(l.flushTimeout):
			l.Notify()
		case <-ctx.Done():
			return
		}
	}
}

// AddListener adds a new model listener.
func (l *Log) AddListener(listener LogsListener) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.listeners = append(l.listeners, listener)
}

// RemoveListener delete a listener from the list.
func (l *Log) RemoveListener(listener LogsListener) {
	l.mx.Lock()
	defer l.mx.Unlock()

	victim := -1
	for i, lis := range l.listeners {
		if lis == listener {
			victim = i
			break
		}
	}

	if victim >= 0 {
		l.listeners = append(l.listeners[:victim], l.listeners[victim+1:]...)
	}
}

func (l *Log) applyFilter(index int, q string) ([][]byte, error) {
	if q == "" {
		return nil, nil
	}
	matches, indices, err := l.lines.Filter(index, q, l.logOptions.ShowTimestamp)
	if err != nil {
		return nil, err
	}

	// No filter!
	if matches == nil {
		ll := make([][]byte, l.lines.Len())
		l.lines.Render(index, l.logOptions.ShowTimestamp, ll)
		return ll, nil
	}
	// Blank filter
	if len(matches) == 0 {
		return nil, nil
	}
	filtered := make([][]byte, 0, len(matches))
	ll := make([][]byte, l.lines.Len())
	l.lines.Lines(index, l.logOptions.ShowTimestamp, ll)
	for i, idx := range matches {
		filtered = append(filtered, color.Highlight(ll[idx], indices[i], 209))
	}

	return filtered, nil
}

func (l *Log) fireLogBuffChanged(index int) {
	ll := make([][]byte, l.lines.Len()-index)
	if l.filter == "" {
		l.lines.Render(index, l.logOptions.ShowTimestamp, ll)
	} else {
		ff, err := l.applyFilter(index, l.filter)
		if err != nil {
			l.fireLogError(err)
			return
		}
		ll = ff
	}

	if len(ll) > 0 {
		l.fireLogChanged(ll)
	}
}

func (l *Log) fireLogResume() {
	for _, lis := range l.listeners {
		lis.LogResume()
	}
}

func (l *Log) fireCanceled() {
	for _, lis := range l.listeners {
		lis.LogCanceled()
	}
}

func (l *Log) fireLogError(err error) {
	for _, lis := range l.listeners {
		lis.LogFailed(err)
	}
}

func (l *Log) fireLogChanged(lines [][]byte) {
	for _, lis := range l.listeners {
		lis.LogChanged(lines)
	}
}

func (l *Log) fireLogCleared() {
	var ll []LogsListener
	l.mx.RLock()
	{
		ll = l.listeners
	}
	l.mx.RUnlock()
	for _, lis := range ll {
		lis.LogCleared()
	}
}
