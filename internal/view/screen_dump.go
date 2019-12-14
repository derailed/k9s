package view

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const dumpTitle = "Screen Dumps"

// ScreenDump presents a directory listing viewer.
type ScreenDump struct {
	ResourceViewer
}

// NewScreenDump returns a new viewer.
func NewScreenDump(gvr client.GVR) ResourceViewer {
	s := ScreenDump{
		ResourceViewer: NewBrowser(gvr),
	}
	// BOZO!! Rename Table
	s.GetTable().SetBorderFocusColor(tcell.ColorSteelBlue)
	s.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.AttrNone)
	s.GetTable().SetColorerFn(render.ScreenDump{}.ColorerFunc())
	s.GetTable().SetSortCol(s.GetTable().NameColIndex(), 0, true)
	s.GetTable().SelectRow(1, true)
	s.GetTable().SetEnterFn(s.edit)
	s.SetContextFn(s.dirContext)

	return &s
}

// BOZO!!
// BOZO !! Need model watcher!
// // Start starts the directory watcher.
// func (s *ScreenDump) Start() {
// 	s.Stop()

// 	s.GetTable().Actions().Delete(tcell.KeyCtrlS)

// 	s.GetTable().Start()
// 	var ctx context.Context
// 	ctx, s.GetTable().cancelFn = context.WithCancel(context.Background())
// 	if err := s.watchDumpDir(ctx); err != nil {
// 		s.App().Flash().Errf("Unable to watch screen dumps directory %s", err)
// 	}
// }

func (s *ScreenDump) dirContext(ctx context.Context) context.Context {
	dir := filepath.Join(config.K9sDumpDir, s.App().Config.K9s.CurrentCluster)
	return context.WithValue(ctx, internal.KeyDir, dir)
}

func (s *ScreenDump) edit(app *App, ns, resource, path string) {
	log.Debug().Msgf("ScreenDump selection is %q", path)

	s.Stop()
	defer s.Start()
	if !edit(true, app, path) {
		app.Flash().Err(errors.New("Failed to launch editor"))
	}
}

func (s *ScreenDump) watchDumpDir(ctx context.Context) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case evt := <-w.Events:
				log.Debug().Msgf("ScreenDump event detected %#v", evt)
				s.Refresh()
			case err := <-w.Errors:
				log.Error().Err(err).Msg("Dir Watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msg("!!!! ScreenDump WATCHER DONE!!")
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing dump watcher")
				}
				return
			}
		}
	}()

	return w.Add(filepath.Join(config.K9sDumpDir, s.App().Config.K9s.CurrentCluster))
}
