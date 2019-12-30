package ui

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Synchronizer manages ui event queue.
type synchronizer interface {
	QueueUpdateDraw(func()) *tview.Application
	QueueUpdate(func()) *tview.Application
}

// Configurator represents an application configurationa.
type Configurator struct {
	skinFile string
	Config   *config.Config
	Styles   *config.Styles
	Bench    *config.Bench
}

// HasSkins returns true if a skin file was located.
func (c *Configurator) HasSkins() bool {
	return c.skinFile != ""
}

// StylesUpdater watches for skin file changes.
func (c *Configurator) StylesUpdater(ctx context.Context, s synchronizer) error {
	if !c.HasSkins() {
		return nil
	}

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
					c.RefreshStyles(c.Config.K9s.CurrentCluster)
				})
			case err := <-w.Errors:
				log.Info().Err(err).Msg("Skin watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msgf("SkinWatcher Done `%s!!", c.skinFile)
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing watcher")
				}
				return
			}
		}
	}()

	log.Debug().Msgf("SkinWatcher watching `%s", c.skinFile)
	return w.Add(c.skinFile)
}

// InitBench load benchmark configuration if any.
func (c *Configurator) InitBench(cluster string) {
	var err error
	if c.Bench, err = config.NewBench(BenchConfig(cluster)); err != nil {
		log.Info().Err(err).Msg("No benchmark config file found, using defaults.")
	}
}

// BenchConfig location of the benchmarks configuration file.
func BenchConfig(cluster string) string {
	return filepath.Join(config.K9sHome, config.K9sBench+"-"+cluster+".yml")
}

// RefreshStyles load for skin configuration changes.
func (c *Configurator) RefreshStyles(cluster string) {
	clusterSkins := filepath.Join(config.K9sHome, fmt.Sprintf("%s_skin.yml", cluster))
	if c.Styles == nil {
		c.Styles = config.NewStyles()
	}
	if err := c.Styles.Load(clusterSkins); err != nil {
		log.Info().Msgf("No cluster specific skin file found -- %s", clusterSkins)
	} else {
		log.Debug().Msgf("Found cluster skins %s", clusterSkins)
		c.updateStyles(clusterSkins)
		return
	}

	if err := c.Styles.Load(config.K9sStylesFile); err != nil {
		log.Info().Msgf("No skin file found -- %s. Loading stock skins.", config.K9sStylesFile)
		return
	}
	c.updateStyles(config.K9sStylesFile)
}

func (c *Configurator) updateStyles(f string) {
	c.skinFile = f
	c.Styles.Update()

	render.StdColor = config.AsColor(c.Styles.Frame().Status.NewColor)
	render.AddColor = config.AsColor(c.Styles.Frame().Status.AddColor)
	render.ModColor = config.AsColor(c.Styles.Frame().Status.ModifyColor)
	render.ErrColor = config.AsColor(c.Styles.Frame().Status.ErrorColor)
	render.HighlightColor = config.AsColor(c.Styles.Frame().Status.HighlightColor)
	render.CompletedColor = config.AsColor(c.Styles.Frame().Status.CompletedColor)
}
