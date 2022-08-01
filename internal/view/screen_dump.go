package view

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
)

// ScreenDump presents a directory listing viewer.
type ScreenDump struct {
	ResourceViewer
}

// NewScreenDump returns a new viewer.
func NewScreenDump(gvr client.GVR) ResourceViewer {
	s := ScreenDump{
		ResourceViewer: NewBrowser(gvr),
	}
	s.GetTable().SetBorderFocusColor(tcell.ColorSteelBlue)
	s.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorRoyalBlue).Attributes(tcell.AttrNone))
	s.GetTable().SetSortCol(ageCol, true)
	s.GetTable().SelectRow(1, true)
	s.GetTable().SetEnterFn(s.edit)
	s.SetContextFn(s.dirContext)

	return &s
}

func (s *ScreenDump) dirContext(ctx context.Context) context.Context {
	dir := filepath.Join(s.App().Config.K9s.GetScreenDumpDir(), s.App().Config.K9s.CurrentCluster)
	log.Debug().Msgf("SD-DIR %q", dir)
	config.EnsureFullPath(dir, config.DefaultDirMod)

	return context.WithValue(ctx, internal.KeyDir, dir)
}

func (s *ScreenDump) edit(app *App, model ui.Tabular, gvr, path string) {
	log.Debug().Msgf("ScreenDump selection is %q", path)

	s.Stop()
	defer s.Start()
	if !edit(app, shellOpts{clear: true, args: []string{path}}) {
		app.Flash().Err(errors.New("Failed to launch editor"))
	}
}
