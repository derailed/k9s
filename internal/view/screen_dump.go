package view

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
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
	s.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.AttrNone)
	s.GetTable().SetColorerFn(render.ScreenDump{}.ColorerFunc())
	s.GetTable().SetSortCol(ageCol, true)
	s.GetTable().SelectRow(1, true)
	s.GetTable().SetEnterFn(s.edit)
	s.SetContextFn(s.dirContext)

	return &s
}

func (s *ScreenDump) dirContext(ctx context.Context) context.Context {
	dir := filepath.Join(config.K9sDumpDir, s.App().Config.K9s.CurrentCluster)
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
