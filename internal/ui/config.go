package ui

import (
	"context"
	"path/filepath"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type synchronizer interface {
	QueueUpdateDraw(func()) *tview.Application
	QueueUpdate(func()) *tview.Application
}

// StylesUpdater watches for skin file changes.
func (c *Configurator) StylesUpdater(ctx context.Context, s synchronizer) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				_ = evt
				s.QueueUpdateDraw(func() {
					c.RefreshStyles()
				})
			case err := <-w.Errors:
				log.Info().Err(err).Msg("Skin watcher failed")
				return
			case <-ctx.Done():
				w.Close()
				return
			}
		}
	}()

	return w.Add(config.K9sStylesFile)
}

// InitBench load benchmark configuration if any.
func (c *Configurator) InitBench(cluster string) {
	var err error
	if c.Bench, err = config.NewBench(BenchConfig(cluster)); err != nil {
		log.Info().Err(err).Msg("No benchmark config file found, using defaults.")
	}
}

// RefreshStyles load for skin configuration changes.
func (c *Configurator) RefreshStyles() {
	var err error
	if c.Styles, err = config.NewStyles(config.K9sStylesFile); err != nil {
		log.Info().Msg("No skin file found. Loading stock skins.")
	}
	if err == nil {
		c.HasSkins = true
	}
	c.Styles.Update()

	StdColor = config.AsColor(c.Styles.Frame().Status.NewColor)
	AddColor = config.AsColor(c.Styles.Frame().Status.AddColor)
	ModColor = config.AsColor(c.Styles.Frame().Status.ModifyColor)
	ErrColor = config.AsColor(c.Styles.Frame().Status.ErrorColor)
	HighlightColor = config.AsColor(c.Styles.Frame().Status.HighlightColor)
	CompletedColor = config.AsColor(c.Styles.Frame().Status.CompletedColor)
}

// BenchConfig location of the benchmarks configuration file.
func BenchConfig(cluster string) string {
	return filepath.Join(config.K9sHome, config.K9sBench+"-"+cluster+".yml")
}
