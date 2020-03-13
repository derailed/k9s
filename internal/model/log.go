package model

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

const logMaxBufferSize = 100

// LogsListener represents a log model listener.
type LogsListener interface {
	// LogChanged notifies the model changed.
	LogChanged([]string)

	// LogCleanred indicates logs are cleared.
	LogCleared()

	// LogFailed indicates a log failure.
	LogFailed(error)
}

// Log represents a resource logger.
type Log struct {
	factory       dao.Factory
	lines         []string
	listeners     []LogsListener
	gvr           client.GVR
	logOptions    dao.LogOptions
	cancelFn      context.CancelFunc
	mx            sync.RWMutex
	filter        string
	lastSent      int
	showTimestamp bool
	timeOut       time.Duration
}

// NewLog returns a new model.
func NewLog(gvr client.GVR, opts dao.LogOptions, timeOut time.Duration) *Log {
	return &Log{
		gvr:        gvr,
		logOptions: opts,
		lines:      nil,
		timeOut:    timeOut,
	}
}

// GetPath returns resource path.
func (l *Log) GetPath() string { return l.logOptions.Path }

// GetContainer returns the resource container if any or "" otherwise.
func (l *Log) GetContainer() string { return l.logOptions.Container }

// Init initializes the model.
func (l *Log) Init(f dao.Factory) {
	l.factory = f
}

// Clear the logs.
func (l *Log) Clear() {
	l.mx.Lock()
	{
		l.lines, l.lastSent = []string{}, 0
	}
	l.mx.Unlock()
	l.fireLogCleared()
}

// ShowTimestamp toggles timestamp on logs.
func (l *Log) ShowTimestamp(b bool) {
	l.mx.RLock()
	defer l.mx.RUnlock()

	l.showTimestamp = b
	l.fireLogCleared()
	l.fireLogChanged(l.lines)
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
func (l *Log) Set(lines []string) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.lines = lines
	l.fireLogChanged(lines)
}

// ClearFilter resets the log filter if any.
func (l *Log) ClearFilter() {
	log.Debug().Msgf("CLEARED!!")
	l.mx.RLock()
	defer l.mx.RUnlock()

	l.filter = ""
	l.fireLogChanged(l.lines)
}

// Filter filters the model using either fuzzy or regexp.
func (l *Log) Filter(q string) error {
	l.mx.RLock()
	defer l.mx.RUnlock()

	log.Debug().Msgf("FILTER!")
	l.filter = q
	filtered, err := applyFilter(l.filter, l.lines)
	if err != nil {
		return err
	}
	l.fireLogCleared()
	l.fireLogChanged(filtered)

	return nil
}

func (l *Log) load() error {
	var ctx context.Context
	ctx = context.WithValue(context.Background(), internal.KeyFactory, l.factory)
	ctx, l.cancelFn = context.WithCancel(ctx)

	c := make(chan []byte, 10)
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
		if l.cancelFn != nil {
			l.cancelFn()
		}
		close(c)
		return err
	}

	return nil
}

// Append adds a log line.
func (l *Log) Append(line string) {
	if line == "" {
		return
	}

	l.mx.Lock()
	defer l.mx.Unlock()

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

func (l *Log) updateLogs(ctx context.Context, c <-chan []byte) {
	defer func() {
		log.Debug().Msgf("updateLogs view bailing out!")
	}()
	for {
		select {
		case bytes, ok := <-c:
			if !ok {
				log.Debug().Msgf("Closed channel detected. Bailing out...")
				l.Append(string(bytes))
				l.Notify(false)
				return
			}
			l.Append(string(bytes))
			var overflow bool
			l.mx.RLock()
			{
				overflow = len(l.lines)-l.lastSent > logMaxBufferSize
			}
			l.mx.RUnlock()
			if overflow {
				l.Notify(true)
			}
		case <-time.After(l.timeOut):
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

func applyFilter(q string, lines []string) ([]string, error) {
	if q == "" {
		return lines, nil
	}
	indexes, err := filter(q, lines)
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
	filtered := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		filtered = append(filtered, lines[idx])
	}

	return filtered, nil
}

func (l *Log) fireLogBuffChanged(lines []string) {
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

func (l *Log) fireLogChanged(lines []string) {
	for _, lis := range l.listeners {
		lis.LogChanged(lines)
	}
}

func (l *Log) fireLogCleared() {
	for _, lis := range l.listeners {
		lis.LogCleared()
	}
}

// ----------------------------------------------------------------------------
// Helpers...

var fuzzyRx = regexp.MustCompile(`\A\-f`)

func isFuzzySelector(s string) bool {
	if s == "" {
		return false
	}
	return fuzzyRx.MatchString(s)
}

func filter(q string, lines []string) ([]int, error) {
	if q == "" {
		return nil, nil
	}
	if isFuzzySelector(q) {
		return fuzzyFilter(strings.TrimSpace(q[2:]), lines), nil
	}
	indexes, err := filterLogs(q, lines)
	if err != nil {
		log.Error().Err(err).Msgf("Logs filter failed")
		return nil, err
	}
	return indexes, nil
}

func fuzzyFilter(q string, lines []string) []int {
	matches := make([]int, 0, len(lines))
	mm := fuzzy.Find(q, lines)
	for _, m := range mm {
		matches = append(matches, m.Index)
	}

	return matches
}

func filterLogs(q string, lines []string) ([]int, error) {
	rx, err := regexp.Compile(`(?i)` + q)
	if err != nil {
		return nil, err
	}
	matches := make([]int, 0, len(lines))
	for i, l := range lines {
		if rx.MatchString(l) {
			matches = append(matches, i)
		}
	}

	return matches, nil
}
