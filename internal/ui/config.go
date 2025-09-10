// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/fsnotify/fsnotify"
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
	customView *config.CustomView
	BenchFile  string
	skinFile   string
}

func (c *Configurator) CustomView() *config.CustomView {
	if c.customView == nil {
		c.customView = config.NewCustomView()
	}

	return c.customView
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
							slog.Warn("Custom views refresh failed", slogs.Error, err)
						}
					})
				}
			case err := <-w.Errors:
				slog.Warn("CustomView watcher failed", slogs.Error, err)
				return
			case <-ctx.Done():
				slog.Debug("CustomViewWatcher canceled", slogs.FileName, config.AppViewsFile)
				if err := w.Close(); err != nil {
					slog.Error("Closing CustomView watcher", slogs.Error, err)
				}
				return
			}
		}
	}()

	if err := w.Add(config.AppViewsFile); err != nil {
		return err
	}
	slog.Debug("Loading custom views", slogs.FileName, config.AppViewsFile)

	return c.RefreshCustomViews()
}

// RefreshCustomViews load view configuration changes.
func (c *Configurator) RefreshCustomViews() error {
	c.CustomView().Reset()

	return c.CustomView().Load(config.AppViewsFile)
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
					slog.Debug("Skin file changed detected", slogs.FileName, c.skinFile)
					s.QueueUpdateDraw(func() {
						c.RefreshStyles(s)
					})
				}
			case err := <-w.Errors:
				slog.Warn("Skin watcher failed", slogs.Error, err)
				return
			case <-ctx.Done():
				slog.Debug("SkinWatcher canceled", slogs.FileName, c.skinFile)
				if err := w.Close(); err != nil {
					slog.Error("Closing Skin watcher", slogs.Error, err)
				}
				return
			}
		}
	}()

	slog.Debug("SkinWatcher initialized", slogs.Dir, config.AppSkinsDir)
	return w.Add(config.AppSkinsDir)
}

// ConfigWatcher watches for config settings changes.
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
					slog.Debug("ConfigWatcher file changed", slogs.FileName, evt.Name)
					if evt.Name == config.AppConfigFile {
						if err := c.Config.Load(evt.Name, false); err != nil {
							slog.Error("K9s config reload failed", slogs.Error, err)
							s.Flash().Warn("k9s config reload failed. Check k9s logs!")
							s.Logo().Warn("K9s config reload failed!")
						}
					} else {
						if err := c.Config.K9s.Reload(); err != nil {
							slog.Error("K9s context config reload failed", slogs.Error, err)
							s.Flash().Warn("Context config reload failed. Check k9s logs!")
							s.Logo().Warn("Context config reload failed!")
						}
					}
					s.QueueUpdateDraw(func() {
						c.RefreshStyles(s)
					})
				}
			case err := <-w.Errors:
				slog.Warn("ConfigWatcher failed", slogs.Error, err)
				return
			case <-ctx.Done():
				slog.Debug("ConfigWatcher canceled")
				if err := w.Close(); err != nil {
					slog.Error("Canceling ConfigWatcher", slogs.Error, err)
				}
				return
			}
		}
	}()

	slog.Debug("ConfigWatcher watching", slogs.FileName, config.AppConfigFile)
	if err := w.Add(config.AppConfigFile); err != nil {
		return err
	}

	cl, ct, ok := c.activeConfig()
	if !ok {
		return nil
	}
	ctConfigFile := config.AppContextConfig(cl, ct)
	slog.Debug("ConfigWatcher watching", slogs.FileName, ctConfigFile)

	return w.Add(ctConfigFile)
}

func (c *Configurator) activeSkin() (string, bool) {
	var skin string
	if c.Config == nil || c.Config.K9s == nil {
		return skin, false
	}

	if env_skin := os.Getenv("K9S_SKIN"); env_skin != "" {
		if _, err := os.Stat(config.SkinFileFromName(env_skin)); err == nil {
			skin = env_skin
			slog.Debug("Loading env skin", slogs.Skin, skin)
			return skin, true
		}
	}

	if ct, err := c.Config.K9s.ActiveContext(); err == nil && ct.Skin != "" {
		if _, err := os.Stat(config.SkinFileFromName(ct.Skin)); err == nil {
			skin = ct.Skin
			slog.Debug("Loading context skin",
				slogs.Skin, skin,
				slogs.Context, c.Config.K9s.ActiveContextName(),
			)
			return skin, true
		}
	}

	if sk := c.Config.K9s.UI.Skin; sk != "" {
		if _, err := os.Stat(config.SkinFileFromName(sk)); err == nil {
			skin = sk
			slog.Debug("Loading global skin", slogs.Skin, skin)
			return skin, true
		}
	}

	return skin, skin != ""
}

func (c *Configurator) activeConfig() (cluster, contxt string, ok bool) {
	if c.Config == nil || c.Config.K9s == nil {
		return
	}
	ct, err := c.Config.K9s.ActiveContext()
	if err != nil {
		return
	}
	cluster, contxt = ct.GetClusterName(), c.Config.K9s.ActiveContextName()
	if cluster != "" && contxt != "" {
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
		slog.Warn("No benchmark config file found",
			slogs.Cluster, cl,
			slogs.Context, ct,
			slogs.Error, err,
		)
	} else {
		c.BenchFile = bc
	}
}

func (c *Configurator) loadSkinFile(synchronizer) {
	skin, ok := c.activeSkin()
	if !ok {
		slog.Debug("No custom skin found. Using stock skin")
		c.updateStyles("")
		return
	}

	skinFile := config.SkinFileFromName(skin)
	slog.Debug("Loading skin file", slogs.Skin, skinFile)
	if err := c.Styles.Load(skinFile); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Warn("Skin file not found in skins dir",
				slogs.Skin, filepath.Base(skinFile),
				slogs.Dir, config.AppSkinsDir,
				slogs.Error, err,
			)
			c.updateStyles("")
		} else {
			slog.Error("Failed to parse skin file",
				slogs.Path, filepath.Base(skinFile),
				slogs.Error, err,
			)
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
