package views

import (
	"context"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type synchronizer interface {
	QueueUpdateDraw(func()) *tview.Application
	QueueUpdate(func()) *tview.Application
}

type configurator struct {
	hasSkins bool
	config   *config.Config
	styles   *config.Styles
	bench    *config.Bench
}

func (c *configurator) stylesUpdater(ctx context.Context, s synchronizer) error {
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
					c.refreshStyles()
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

func (c *configurator) initBench(cluster string) {
	var err error
	if c.bench, err = config.NewBench(benchConfig(cluster)); err != nil {
		log.Info().Err(err).Msg("No benchmark config file found, using defaults.")
	}
}

func (c *configurator) refreshStyles() {
	var err error
	if c.styles, err = config.NewStyles(config.K9sStylesFile); err != nil {
		log.Info().Msg("No skin file found. Loading stock skins.")
	}
	if err == nil {
		c.hasSkins = true
	}
	c.styles.Update()

	stdColor = config.AsColor(c.styles.Frame().Status.NewColor)
	addColor = config.AsColor(c.styles.Frame().Status.AddColor)
	modColor = config.AsColor(c.styles.Frame().Status.ModifyColor)
	errColor = config.AsColor(c.styles.Frame().Status.ErrorColor)
	highlightColor = config.AsColor(c.styles.Frame().Status.HighlightColor)
	completedColor = config.AsColor(c.styles.Frame().Status.CompletedColor)
}
