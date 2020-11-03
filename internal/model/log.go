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

	// LogCleanred indicates logs are cleared.
	LogCleared()

	// LogFailed indicates a log failure.
	LogFailed(error)
}

// Log represents a resource logger.
type Log struct {
	factory      dao.Factory
	lines        dao.LogItems
	listeners    []LogsListener
	gvr          client.GVR
	logOptions   dao.LogOptions
	cancelFn     context.CancelFunc
	mx           sync.RWMutex
	filter       string
	lastSent     int
	flushTimeout time.Duration
}

// NewLog returns a new model.
func NewLog(gvr client.GVR, opts dao.LogOptions, flushTimeout time.Duration) *Log {
	return &Log{
		gvr:          gvr,
		logOptions:   opts,
		lines:        nil,
		flushTimeout: flushTimeout,
	}
}

// SinceSeconds returns since seconds option.
func (l *Log) SinceSeconds() int64 {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.logOptions.SinceSeconds
}

// ToggleShowTimestamp toggles to logs timestamps.
func (l *Log) ToggleShowTimestamp(b bool) {
	l.logOptions.ShowTimestamp = b
	l.Refresh()
}

// SetSinceSeconds sets the logs retrieval time.
func (l *Log) SetSinceSeconds(i int64) {
	l.logOptions.SinceSeconds = i
	l.Restart()
}

// Configure sets logger configuration.
func (l *Log) Configure(opts *config.Logger) {
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

// Init initializes the model.
func (l *Log) Init(f dao.Factory) {
	l.factory = f
}

// Clear the logs.
func (l *Log) Clear() {
	l.mx.Lock()
	{
		l.lines, l.lastSent = dao.LogItems{}, 0
	}
	l.mx.Unlock()

	l.fireLogCleared()
}

// Refresh refreshes the logs.
func (l *Log) Refresh() {
	l.fireLogCleared()
	ll := make([][]byte, len(l.lines))
	l.lines.Render(l.logOptions.ShowTimestamp, ll)
	l.fireLogChanged(ll)
}

// Restart restarts the logger.
func (l *Log) Restart() {
	l.Clear()
	l.Stop()
	l.Start()
}

// Start initialize log tailer.
func (l *Log) Start() {
	if err := l.load(); err != nil {
		log.Error().Err(err).Msgf("Tail logs failed!")
		l.fireLogError(err)
	}
}

// Stop terminates log tailing.
func (l *Log) Stop() {
	defer log.Debug().Msgf("<<<< Logger STOPPED!")
	if l.cancelFn != nil {
		l.cancelFn()
		l.cancelFn = nil
	}
}

// Set sets the log lines (for testing only!)
func (l *Log) Set(items dao.LogItems) {
	l.mx.Lock()
	{
		l.lines = items
	}
	l.mx.Unlock()

	l.fireLogCleared()
	ll := make([][]byte, len(l.lines))
	l.lines.Render(l.logOptions.ShowTimestamp, ll)
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
	ll := make([][]byte, len(l.lines))
	l.lines.Render(l.logOptions.ShowTimestamp, ll)
	l.fireLogChanged(ll)
}

// Filter filters the model using either fuzzy or regexp.
func (l *Log) Filter(q string) {
	l.mx.Lock()
	defer l.mx.Unlock()

	if len(q) == 0 {
		l.filter = ""
		l.fireLogCleared()
		l.fireLogBuffChanged(l.lines)
		return
	}

	l.filter = q
	l.fireLogCleared()
	l.fireLogBuffChanged(l.lines)
}

func (l *Log) load() error {
	var ctx context.Context
	ctx = context.WithValue(context.Background(), internal.KeyFactory, l.factory)
	ctx, l.cancelFn = context.WithCancel(ctx)

	c := make(dao.LogChan, 10)
	go l.updateLogs(ctx, c)

	accessor, err := dao.AccessorFor(l.factory, l.gvr)
	if err != nil {
		return err
	}
	logger, ok := accessor.(dao.Loggable)
	if !ok {
		return fmt.Errorf("Resource %s is not Loggable", l.gvr)
	}

	go func() {
		if err = logger.TailLogs(ctx, c, l.logOptions); err != nil {
			log.Error().Err(err).Msgf("Tail logs failed")
			if l.cancelFn != nil {
				l.cancelFn()
			}
		}
	}()

	return nil
}

// Append adds a log line.
func (l *Log) Append(line *dao.LogItem) {
	if line == nil || line.IsEmpty() {
		return
	}

	var lines dao.LogItems
	l.mx.Lock()
	{
		l.logOptions.SinceTime = line.Timestamp
		lines = l.lines
	}
	l.mx.Unlock()

	if lines == nil {
		l.fireLogCleared()
	}

	l.mx.Lock()
	defer l.mx.Unlock()
	if len(l.lines) < int(l.logOptions.Lines) {
		l.lines = append(l.lines, line)
		return
	}
	l.lines = append(l.lines[1:], line)
	l.lastSent--
	if l.lastSent < 0 {
		l.lastSent = 0
	}
}

// Notify fires of notifications to the listeners.
func (l *Log) Notify() {
	l.mx.Lock()
	defer l.mx.Unlock()

	if l.lastSent < len(l.lines) {
		l.fireLogBuffChanged(l.lines[l.lastSent:])
		l.lastSent = len(l.lines)
	}
}

func (l *Log) updateLogs(ctx context.Context, c dao.LogChan) {
	defer func() {
		log.Debug().Msgf("updateLogs view bailing out!")
	}()
	for {
		select {
		case item, ok := <-c:
			if !ok {
				log.Debug().Msgf("Closed channel detected. Bailing out...")
				l.Append(item)
				l.Notify()
				return
			}
			l.Append(item)
			var overflow bool
			l.mx.RLock()
			{
				overflow = int64(len(l.lines)-l.lastSent) > l.logOptions.Lines
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
	l.listeners = append(l.listeners, listener)
}

// RemoveListener delete a listener from the lisl.
func (l *Log) RemoveListener(listener LogsListener) {
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

func (l *Log) applyFilter(q string) ([][]byte, error) {
	if q == "" {
		return nil, nil
	}
	matches, indices, err := l.lines.Filter(q, l.logOptions.ShowTimestamp)
	if err != nil {
		return nil, err
	}

	// No filter!
	if matches == nil {
		ll := make([][]byte, len(l.lines))
		l.lines.Render(l.logOptions.ShowTimestamp, ll)
		return ll, nil
	}
	// Blank filter
	if len(matches) == 0 {
		return nil, nil
	}
	filtered := make([][]byte, 0, len(matches))
	lines := l.lines.Lines(l.logOptions.ShowTimestamp)
	for i, idx := range matches {
		filtered = append(filtered, color.Highlight(lines[idx], indices[i], 209))
	}

	return filtered, nil
}

func (l *Log) fireLogBuffChanged(lines dao.LogItems) {
	ll := make([][]byte, len(lines))
	if l.filter == "" {
		lines.Render(l.logOptions.ShowTimestamp, ll)
	} else {
		ff, err := l.applyFilter(l.filter)
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
	for _, lis := range l.listeners {
		lis.LogCleared()
	}
}
