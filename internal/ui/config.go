// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"errors"
	"os"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Synchronizer manages ui event queue.
type synchronizer interface {
	QueueUpdateDraw(func())
	QueueUpdate(func())
}

// Configurator represents an application configuration.
type Configurator struct {
	Config     *config.Config
	Styles     *config.Styles
	CustomView *config.CustomView
	BenchFile  string
	skinFile   string
}

// HasSkin returns true if a skin file was located.
func (c *Configurator) HasSkin() bool {
	return c.skinFile != ""
}

// CustomViewsWatcher watches for view config file changes.
func (c *Configurator) CustomViewsWatcher(ctx context.Context, s synchronizer) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				if evt.Name == config.AppViewsFile {
					s.QueueUpdateDraw(func() {
						if err := c.RefreshCustomViews(); err != nil {
							log.Warn().Err(err).Msgf("Custom views refresh failed")
						}
					})
				}
			case err := <-w.Errors:
				log.Warn().Err(err).Msg("CustomView watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msgf("CustomViewWatcher CANCELED `%s!!", config.AppViewsFile)
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing CustomView watcher")
				}
				return
			}
		}
	}()

	if err := c.RefreshCustomViews(); err != nil {
		return err
	}
	return w.Add(config.AppViewsFile)
}

// RefreshCustomViews load view configuration changes.
func (c *Configurator) RefreshCustomViews() error {
	if c.CustomView == nil {
		c.CustomView = config.NewCustomView()
	} else {
		c.CustomView.Reset()
	}

	return c.CustomView.Load(config.AppViewsFile)
}

// StylesWatcher watches for skin file changes.
func (c *Configurator) StylesWatcher(ctx context.Context, s synchronizer) error {
	if !c.HasSkin() {
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
				if evt.Name == c.skinFile && evt.Op != fsnotify.Chmod {
					log.Debug().Msgf("Skin changed: %s", c.skinFile)
					s.QueueUpdateDraw(func() {
						c.RefreshStyles(c.Config.K9s.ActiveContextName())
					})
				}
			case err := <-w.Errors:
				log.Info().Err(err).Msg("Skin watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msgf("SkinWatcher CANCELED `%s!!", c.skinFile)
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing Skin watcher")
				}
				return
			}
		}
	}()

	log.Debug().Msgf("SkinWatcher watching %q", config.K9sHome())
	if err := w.Add(config.K9sHome()); err != nil {
		return err
	}
	log.Debug().Msgf("SkinWatcher watching %q", config.AppSkinsDir)
	return w.Add(config.AppSkinsDir)
}

// RefreshStyles load for skin configuration changes.
func (c *Configurator) RefreshStyles(context string) {
	cluster := "na"
	if c.Config != nil {
		if ct, err := c.Config.K9s.ActiveContext(); err == nil {
			cluster = ct.ClusterName
		}
	}

	if bc, err := config.EnsureBenchmarksCfgFile(cluster, context); err != nil {
		log.Warn().Err(err).Msgf("No benchmark config file found for context: %s", context)
	} else {
		c.BenchFile = bc
	}

	if c.Styles == nil {
		c.Styles = config.NewStyles()
	} else {
		c.Styles.Reset()
	}

	var skin string
	if c.Config != nil {
		skin = c.Config.K9s.UI.Skin
		if ct, err := c.Config.K9s.ActiveContext(); err != nil {
			log.Warn().Msgf("No active context found. Using default skin")
		} else if ct.Skin != "" {
			skin = ct.Skin
		}
	}
	if skin == "" {
		c.updateStyles("")
		return
	}
	var skinFile = config.SkinFileFromName(skin)
	if err := c.Styles.Load(skinFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn().Msgf("Skin file %q not found in skins dir: %s", skinFile, config.AppSkinsDir)
		} else {
			log.Error().Msgf("Failed to parse skin file -- %s: %s.", skinFile, err)
		}
	} else {
		c.updateStyles(skinFile)
	}
}

func (c *Configurator) updateStyles(f string) {
	c.skinFile = f
	if !c.HasSkin() {
		c.Styles.DefaultSkin()
	}
	c.Styles.Update()

	render.ModColor = c.Styles.Frame().Status.ModifyColor.Color()
	render.AddColor = c.Styles.Frame().Status.AddColor.Color()
	render.ErrColor = c.Styles.Frame().Status.ErrorColor.Color()
	render.StdColor = c.Styles.Frame().Status.NewColor.Color()
	render.PendingColor = c.Styles.Frame().Status.PendingColor.Color()
	render.HighlightColor = c.Styles.Frame().Status.HighlightColor.Color()
	render.KillColor = c.Styles.Frame().Status.KillColor.Color()
	render.CompletedColor = c.Styles.Frame().Status.CompletedColor.Color()
}
