package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog/log"
)

// LogsListener represents a log model listener.
type LogsListener interface {
	// LogChanged notifies the model changed.
	LogChanged(dao.LogItems)

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

// LogOptions returns the current log options.
func (l *Log) LogOptions() dao.LogOptions {
	return l.logOptions
}

// SinceSeconds returns since seconds option.
func (l *Log) SinceSeconds() int64 {
	l.mx.RLock()
	defer l.mx.RUnlock()
	return l.logOptions.SinceSeconds
}

// SetLogOptions updates logger options.
func (l *Log) SetLogOptions(opts dao.LogOptions) {
	l.logOptions = opts
	l.Restart()
}

// Configure sets logger configuration.
func (l *Log) Configure(opts *config.Logger) {
	l.logOptions.Lines = int64(opts.BufferSize)
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
	l.fireLogChanged(l.lines)
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
	defer l.mx.Unlock()
	l.lines = items
	l.fireLogCleared()
	l.fireLogChanged(items)
}

// ClearFilter resets the log filter if any.
func (l *Log) ClearFilter() {
	l.mx.RLock()
	defer l.mx.RUnlock()

	l.filter = ""
	l.fireLogCleared()
	l.fireLogChanged(l.lines)
}

// Filter filters the model using either fuzzy or regexp.
func (l *Log) Filter(q string) error {
	l.mx.RLock()
	defer l.mx.RUnlock()

	l.filter = q
	l.fireLogCleared()
	l.fireLogBuffChanged(l.lines)

	return nil
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
	if err := logger.TailLogs(ctx, c, l.logOptions); err != nil {
		log.Error().Err(err).Msgf("Tail logs failed")
		if l.cancelFn != nil {
			l.cancelFn()
		}
		return err
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

	l.logOptions.SinceTime = line.Timestamp
	if l.lines == nil {
		l.fireLogCleared()
	}

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
func (l *Log) Notify(timedOut bool) {
	l.mx.Lock()
	defer l.mx.Unlock()

	if timedOut && l.lastSent < len(l.lines) {
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
				l.Notify(true)
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
				l.Notify(true)
			}
		case <-time.After(l.flushTimeout):
			l.Notify(true)
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

func applyFilter(q string, lines dao.LogItems) (dao.LogItems, error) {
	if q == "" {
		return lines, nil
	}
	indexes, err := lines.Filter(q)
	if err != nil {
		return nil, err
	}
	// No filter!
	if indexes == nil {
		return lines, nil
	}
	// Blank filter
	if len(indexes) == 0 {
		return nil, nil
	}
	filtered := make(dao.LogItems, 0, len(indexes))
	for _, idx := range indexes {
		filtered = append(filtered, lines[idx])
	}

	return filtered, nil
}

func (l *Log) fireLogBuffChanged(lines dao.LogItems) {
	filtered, err := applyFilter(l.filter, lines)
	if err != nil {
		l.fireLogError(err)
		return
	}
	if len(filtered) > 0 {
		l.fireLogChanged(filtered)
	}
}

func (l *Log) fireLogError(err error) {
	for _, lis := range l.listeners {
		lis.LogFailed(err)
	}
}

func (l *Log) fireLogChanged(lines dao.LogItems) {
	for _, lis := range l.listeners {
		lis.LogChanged(lines)
	}
}

func (l *Log) fireLogCleared() {
	for _, lis := range l.listeners {
		lis.LogCleared()
	}
}
