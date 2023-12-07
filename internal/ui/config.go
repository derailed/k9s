// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"errors"
	"fmt"
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
				if evt.Name == config.K9sViewConfigFile {
					s.QueueUpdateDraw(func() {
						c.RefreshCustomViews()
					})
				}
			case err := <-w.Errors:
				log.Warn().Err(err).Msg("CustomView watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msgf("CustomViewWatcher CANCELED `%s!!", config.K9sViewConfigFile)
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing CustomView watcher")
				}
				return
			}
		}
	}()

	log.Debug().Msgf("CustomView watching `%s", config.K9sViewConfigFile)
	c.RefreshCustomViews()
	return w.Add(config.K9sHome())
}

// RefreshCustomViews load view configuration changes.
func (c *Configurator) RefreshCustomViews() {
	if c.CustomView == nil {
		c.CustomView = config.NewCustomView()
	} else {
		c.CustomView.Reset()
	}

	if err := c.CustomView.Load(config.K9sViewConfigFile); err != nil {
		log.Warn().Err(err).Msgf("Custom view load failed %s", config.K9sViewConfigFile)
		return
	}
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
						c.RefreshStyles(c.Config.K9s.CurrentCluster)
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
	log.Debug().Msgf("SkinWatcher watching %q", config.K9sSkinDir)
	return w.Add(config.K9sSkinDir)
}

// BenchConfig location of the benchmarks configuration file.
func BenchConfig(context string) string {
	return filepath.Join(config.K9sHome(), config.K9sBench+"-"+context+".yml")
}

func (c *Configurator) clusterFromContext(name string) (*config.Cluster, error) {
	if c.Config == nil || c.Config.GetConnection() == nil {
		return nil, fmt.Errorf("No config set in configurator")
	}

	cc, err := c.Config.GetConnection().Config().Contexts()
	if err != nil {
		return nil, errors.New("unable to retrieve contexts map")
	}

	context, ok := cc[name]
	if !ok {
		return nil, fmt.Errorf("no context named %s found", name)
	}

	cl, ok := c.Config.K9s.Clusters[context.Cluster]
	if !ok {
		return nil, fmt.Errorf("no cluster named %s found", context.Cluster)
	}

	return cl, nil
}

// RefreshStyles load for skin configuration changes.
func (c *Configurator) RefreshStyles(context string) {
	c.BenchFile = BenchConfig(context)

	if c.Styles == nil {
		c.Styles = config.NewStyles()
	} else {
		c.Styles.Reset()
	}

	var skin string
	cl, err := c.clusterFromContext(context)
	if err != nil {
		log.Warn().Err(err).Msgf("No cluster found. Using default skin")
	} else {
		skin = cl.Skin
	}

	var (
		skinFile = filepath.Join(config.K9sSkinDir, skin+".yml")
	)
	if skin != "" {
		if err := c.Styles.Load(skinFile); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Warn().Msgf("Skin file %q not found in skins dir: %s", skinFile, config.K9sSkinDir)
			} else {
				log.Error().Msgf("Failed to parse skin file -- %s: %s.", skinFile, err)
			}
		} else {
			c.updateStyles(skinFile)
			return
		}
	}

	if err := c.Styles.Load(config.K9sStylesFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Warn().Msgf("No skin file found -- %s. Loading stock skins.", config.K9sStylesFile)
		} else {
			log.Error().Msgf("Failed to parse skin file -- %s. %s. Loading stock skins.", config.K9sStylesFile, err)
		}
		c.updateStyles("")
		return
	}
	c.updateStyles(config.K9sStylesFile)
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
