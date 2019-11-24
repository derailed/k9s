package view

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const dumpTitle = "Screen Dumps"

var dumpHeader = resource.Row{"NAME", "AGE"}

// ScreenDump presents a directory listing viewer.
type ScreenDump struct {
	*Table
}

// NewScreenDump returns a new viewer.
func NewScreenDump(_, _ string, _ resource.List) ResourceViewer {
	return &ScreenDump{
		Table: NewTable(dumpTitle),
	}
}

// Init initializes the viewer.
func (s *ScreenDump) Init(ctx context.Context) error {
	if err := s.Table.Init(ctx); err != nil {
		return nil
	}
	s.bindKeys()
	s.SetBorderFocusColor(tcell.ColorSteelBlue)
	s.SetSelectedStyle(tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.AttrNone)
	s.SetColorerFn(dumpColorer)
	s.ActiveNS = resource.AllNamespaces
	s.SetSortCol(s.NameColIndex(), 0, true)
	s.SelectRow(1, true)

	s.Start()
	s.refresh()

	return nil
}

func (r *ScreenDump) GetTable() *Table { return r.Table }
func (r *ScreenDump) SetEnvFn(EnvFunc) {}

func (s *ScreenDump) List() resource.List {
	return nil
}

// Start starts the directory watcher.
func (s *ScreenDump) Start() {
	var ctx context.Context
	ctx, s.cancelFn = context.WithCancel(context.Background())
	if err := s.watchDumpDir(ctx); err != nil {
		s.app.Flash().Errf("Unable to watch screen dumps directory %s", err)
	}
}

// Name returns the component name.
func (s *ScreenDump) Name() string {
	return dumpTitle
}

func (s *ScreenDump) refresh() {
	s.Update(s.hydrate())
	s.UpdateTitle()
}

func (s *ScreenDump) bindKeys() {
	s.Actions().Add(ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", s.app.PrevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", s.enterCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", s.deleteCmd, true),
		tcell.KeyCtrlS: ui.NewKeyAction("Save", noopCmd, false),
	})
}

func (s *ScreenDump) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msg("Dump enter!")
	if s.SearchBuff().IsActive() {
		return s.filterCmd(evt)
	}
	sel := s.GetSelectedItem()
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, s.app.Config.K9s.CurrentCluster)
	if !edit(true, s.app, filepath.Join(dir, sel)) {
		s.app.Flash().Err(errors.New("Failed to launch editor"))
	}

	return nil
}

func (s *ScreenDump) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := s.GetSelectedItem()
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, s.app.Config.K9s.CurrentCluster)
	showModal(s.app.Content.Pages, fmt.Sprintf("Delete screen dump `%s?", sel), func() {
		if err := os.Remove(filepath.Join(dir, sel)); err != nil {
			s.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		s.refresh()
		s.app.Flash().Infof("ScreenDump file %s deleted!", sel)
	})

	return nil
}

func (s *ScreenDump) hydrate() resource.TableData {
	data := resource.TableData{
		Header:    dumpHeader,
		Rows:      make(resource.RowEvents, 10),
		Namespace: resource.NotNamespaced,
	}

	dir := filepath.Join(config.K9sDumpDir, s.app.Config.K9s.CurrentCluster)
	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		s.app.Flash().Errf("Unable to read dump directory %s", err)
	}

	for _, f := range ff {
		fields := resource.Row{f.Name(), time.Since(f.ModTime()).String()}
		data.Rows[f.Name()] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
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
				log.Debug().Msgf("Dump event %#v", evt)
				s.app.QueueUpdateDraw(func() {
					s.refresh()
				})
			case err := <-w.Errors:
				log.Info().Err(err).Msg("Dir Watcher failed")
				return
			case <-ctx.Done():
				log.Debug().Msg("!!!! FS WATCHER DONE!!")
				if err := w.Close(); err != nil {
					log.Error().Err(err).Msg("Closing dump watcher")
				}
				return
			}
		}
	}()

	return w.Add(filepath.Join(config.K9sDumpDir, s.app.Config.K9s.CurrentCluster))
}

// Helpers...

func noopCmd(*tcell.EventKey) *tcell.EventKey {
	return nil
}
