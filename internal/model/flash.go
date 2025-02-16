// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	// DefaultFlashDelay sets the flash clear delay.
	DefaultFlashDelay = 3 * time.Second

	// FlashInfo represents an info message.
	FlashInfo FlashLevel = iota
	// FlashWarn represents an warning message.
	FlashWarn
	// FlashErr represents an error message.
	FlashErr
)

// LevelMessage tracks a message and severity.
type LevelMessage struct {
	Level FlashLevel
	Text  string
}

func newClearMessage() LevelMessage {
	return LevelMessage{}
}

// IsClear returns true if message is empty.
func (l LevelMessage) IsClear() bool {
	return l.Text == ""
}

// FlashLevel represents flash message severity.
type FlashLevel int

// FlashChan represents a flash event channel.
type FlashChan chan LevelMessage

// FlashListener represents a text model listener.
type FlashListener interface {
	// FlashChanged notifies the model changed.
	FlashChanged(FlashLevel, string)

	// FlashCleared notifies when the filter changed.
	FlashCleared()
}

// Flash represents a flash message model.
type Flash struct {
	msg     LevelMessage
	cancel  context.CancelFunc
	delay   time.Duration
	msgChan chan LevelMessage
}

// NewFlash returns a new instance.
func NewFlash(dur time.Duration) *Flash {
	return &Flash{
		delay:   dur,
		msgChan: make(FlashChan, 3),
	}
}

// Channel returns the flash channel.
func (f *Flash) Channel() FlashChan {
	return f.msgChan
}

// Info displays an info flash message.
func (f *Flash) Info(msg string) {
	f.SetMessage(FlashInfo, msg)
}

// Infof displays a formatted info flash message.
func (f *Flash) Infof(fmat string, args ...interface{}) {
	f.Info(fmt.Sprintf(fmat, args...))
}

// Warn displays a warning flash message.
func (f *Flash) Warn(msg string) {
	log.Warn().Msg(msg)
	f.SetMessage(FlashWarn, msg)
}

// Warnf displays a formatted warning flash message.
func (f *Flash) Warnf(fmat string, args ...interface{}) {
	f.Warn(fmt.Sprintf(fmat, args...))
}

// Err displays an error flash message.
func (f *Flash) Err(err error) {
	log.Error().Msg(err.Error())
	f.SetMessage(FlashErr, err.Error())
}

// Errf displays a formatted error flash message.
func (f *Flash) Errf(fmat string, args ...interface{}) {
	var err error
	for _, a := range args {
		switch e := a.(type) {
		case error:
			err = e
		}
	}
	log.Error().Err(err).Msgf(fmat, args...)
	f.SetMessage(FlashErr, fmt.Sprintf(fmat, args...))
}

// Clear clears the flash message.
func (f *Flash) Clear() {
	f.fireCleared()
}

// SetMessage sets the flash level message.
func (f *Flash) SetMessage(level FlashLevel, msg string) {
	if f.cancel != nil {
		f.cancel()
		f.cancel = nil
	}

	f.setLevelMessage(LevelMessage{Level: level, Text: msg})
	f.fireFlashChanged()

	ctx := context.Background()
	ctx, f.cancel = context.WithCancel(ctx)
	go f.refresh(ctx)
}

func (f *Flash) refresh(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(f.delay):
			f.fireCleared()
			return
		}
	}
}

func (f *Flash) setLevelMessage(msg LevelMessage) {
	f.msg = msg
}

func (f *Flash) fireFlashChanged() {
	f.msgChan <- f.msg
}

func (f *Flash) fireCleared() {
	f.msgChan <- newClearMessage()
}
