// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"

	"github.com/derailed/k9s/internal/config"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

// Synchronizer manages ui event queue.
type synchronizer interface {
	Flash() *model.Flash
	Logo() *Logo
	UpdateClusterInfo()
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
				if evt.Name == config.AppViewsFile && evt.Op != fsnotify.Chmod {
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

	if err := w.Add(config.AppViewsFile); err != nil {
		return err
	}

	return c.RefreshCustomViews()
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
	if _, err := os.Stat(config.AppSkinsDir); errors.Is(err, fs.ErrNotExist) {
		return err
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case evt := <-w.Events:
				if evt.Op != fsnotify.Chmod && filepath.Base(evt.Name) == filepath.Base(c.skinFile) {
					log.Debug().Msgf("Skin changed: %s", c.skinFile)
					s.QueueUpdateDraw(func() {
						c.RefreshStyles(s)
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
					log.Debug().Msgf("ConfigWatcher file changed: %s", evt.Name)
					if evt.Name == config.AppConfigFile {
						if err := c.Config.Load(evt.Name, false); err != nil {
							log.Error().Err(err).Msgf("k9s config reload failed")
							s.Flash().Warn("k9s config reload failed. Check k9s logs!")
							s.Logo().Warn("K9s config reload failed!")
						}
					} else {
						if err := c.Config.K9s.Reload(); err != nil {
							log.Error().Err(err).Msgf("k9s context config reload failed")
							s.Flash().Warn("Context config reload failed. Check k9s logs!")
							s.Logo().Warn("Context config reload failed!")
						}
					}
					s.QueueUpdateDraw(func() {
						c.RefreshStyles(s)
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

	if ct, err := c.Config.K9s.ActiveContext(); err == nil && ct.Skin != "" {
		if _, err := os.Stat(config.SkinFileFromName(ct.Skin)); err == nil {
			skin = ct.Skin
			log.Debug().Msgf("[Skin] Loading context skin (%q) from %q", skin, c.Config.K9s.ActiveContextName())
		}
	}

	if sk := c.Config.K9s.UI.Skin; skin == "" && sk != "" {
		if _, err := os.Stat(config.SkinFileFromName(sk)); err == nil {
			skin = sk
			log.Debug().Msgf("[Skin] Loading global skin (%q)", skin)
		}
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
	cluster, context = ct.GetClusterName(), c.Config.K9s.ActiveContextName()
	if cluster != "" && context != "" {
		ok = true
	}

	return
}

// RefreshStyles load for skin configuration changes.
func (c *Configurator) RefreshStyles(s synchronizer) {
	s.UpdateClusterInfo()
	if c.Styles == nil {
		c.Styles = config.NewStyles()
	}
	defer c.loadSkinFile(s)

	cl, ct, ok := c.activeConfig()
	if !ok {
		return
	}
	// !!BOZO!! Lame move out!
	if bc, err := config.EnsureBenchmarksCfgFile(cl, ct); err != nil {
		log.Warn().Err(err).Msgf("No benchmark config file found: %q@%q", cl, ct)
	} else {
		c.BenchFile = bc
	}
}

func (c *Configurator) loadSkinFile(s synchronizer) {
	skin, ok := c.activeSkin()
	if !ok {
		log.Debug().Msgf("No custom skin found. Using stock skin")
		c.updateStyles("")
		return
	}

	skinFile := config.SkinFileFromName(skin)
	log.Debug().Msgf("Loading skin file: %q", skinFile)
	if err := c.Styles.Load(skinFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn().Msgf("Skin file %q not found in skins dir: %s", filepath.Base(skinFile), config.AppSkinsDir)
			c.updateStyles("")
		} else {
			log.Error().Msgf("Failed to parse skin file -- %s: %s.", filepath.Base(skinFile), err)
			c.updateStyles(skinFile)
		}
	} else {
		c.updateStyles(skinFile)
	}
}

func (c *Configurator) updateStyles(f string) {
	c.skinFile = f
	if f == "" {
		c.Styles.Reset()
	}
	c.Styles.Update()

	model1.ModColor = c.Styles.Frame().Status.ModifyColor.Color()
	model1.AddColor = c.Styles.Frame().Status.AddColor.Color()
	model1.ErrColor = c.Styles.Frame().Status.ErrorColor.Color()
	model1.StdColor = c.Styles.Frame().Status.NewColor.Color()
	model1.PendingColor = c.Styles.Frame().Status.PendingColor.Color()
	model1.HighlightColor = c.Styles.Frame().Status.HighlightColor.Color()
	model1.KillColor = c.Styles.Frame().Status.KillColor.Color()
	model1.CompletedColor = c.Styles.Frame().Status.CompletedColor.Color()
}
