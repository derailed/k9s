// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"

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

// SkinsDirWatcher watches for skin directory file changes.
func (c *Configurator) SkinsDirWatcher(ctx context.Context, s synchronizer) error {
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
						c.RefreshStyles()
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

	log.Debug().Msgf("SkinWatcher watching %q", config.AppSkinsDir)
	return w.Add(config.AppSkinsDir)
}

// ConfigWatcher watches for skin settings changes.
func (c *Configurator) ConfigWatcher(ctx context.Context, s synchronizer) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				if evt.Has(fsnotify.Create) || evt.Has(fsnotify.Write) {
					log.Debug().Msgf("ConfigWatcher file changed: %s -- %#v", evt.Name, evt.Op.String())
					if evt.Name == config.AppConfigFile {
						if err := c.Config.Load(evt.Name); err != nil {
							log.Error().Err(err).Msgf("Config reload failed")
						}
					} else {
						if err := c.Config.K9s.Reload(); err != nil {
							log.Error().Err(err).Msgf("Context config reload failed")
						}
					}
					s.QueueUpdateDraw(func() {
						c.RefreshStyles()
					})
				}
			case err := <-w.Errors:
				log.Info().Err(err).Msg("ConfigWatcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msg("ConfigWatcher CANCELED")
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Canceling ConfigWatcher")
				}
				return
			}
		}
	}()

	log.Debug().Msgf("ConfigWatcher watching: %q", config.AppConfigFile)
	if err := w.Add(config.AppConfigFile); err != nil {
		return err
	}

	cl, ct, ok := c.activeConfig()
	if !ok {
		return nil
	}
	ctConfigFile := filepath.Join(config.AppContextConfig(cl, ct))
	log.Debug().Msgf("ConfigWatcher watching: %q", ctConfigFile)

	return w.Add(ctConfigFile)
}

func (c *Configurator) activeSkin() (string, bool) {
	var skin string
	if c.Config == nil || c.Config.K9s == nil {
		return skin, false
	}

	if ct, err := c.Config.K9s.ActiveContext(); err == nil {
		skin = ct.Skin
	}
	if skin == "" {
		skin = c.Config.K9s.UI.Skin
	}

	return skin, skin != ""
}

func (c *Configurator) activeConfig() (cluster string, context string, ok bool) {
	if c.Config == nil || c.Config.K9s == nil {
		return
	}
	ct, err := c.Config.K9s.ActiveContext()
	if err != nil {
		return
	}
	cluster, context = ct.ClusterName, c.Config.K9s.ActiveContextName()
	if cluster != "" && context != "" {
		ok = true
	}

	return
}

// RefreshStyles load for skin configuration changes.
func (c *Configurator) RefreshStyles() {
	if c.Styles == nil {
		c.Styles = config.NewStyles()
	}

	cl, ct, ok := c.activeConfig()
	if !ok {
		log.Debug().Msgf("No custom skin found. Using stock skin")
		c.updateStyles("")
		return
	}

	if bc, err := config.EnsureBenchmarksCfgFile(cl, ct); err != nil {
		log.Warn().Err(err).Msgf("No benchmark config file found: %q@%q", cl, ct)
	} else {
		c.BenchFile = bc
	}

	skin, ok := c.activeSkin()
	if !ok {
		log.Debug().Msgf("No custom skin found. Using stock skin")
		c.updateStyles("")
		return
	}
	skinFile := config.SkinFileFromName(skin)
	if err := c.Styles.Load(skinFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn().Msgf("Skin file %q not found in skins dir: %s", skinFile, config.AppSkinsDir)
		} else {
			log.Error().Msgf("Failed to parse skin file -- %s: %s.", skinFile, err)
		}
		c.updateStyles("")
	} else {
		log.Debug().Msgf("Loading skin file: %q", skinFile)
		c.updateStyles(skinFile)
	}
}

func (c *Configurator) updateStyles(f string) {
	c.skinFile = f
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
