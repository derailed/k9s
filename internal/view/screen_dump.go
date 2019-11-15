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
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/fsnotify/fsnotify"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	dumpTitle    = "Screen Dumps"
	dumpTitleFmt = " [mediumvioletred::b]%s([fuchsia::b]%d[fuchsia::-])[mediumvioletred::-] "
)

var (
	dumpHeader = resource.Row{"NAME", "AGE"}
)

// ScreenDump presents a directory listing viewer.
type ScreenDump struct {
	*MasterDetail

	cancelFn context.CancelFunc
	app      *App
}

func NewScreenDump(_, _ string, _ resource.List) ResourceViewer {
	return &ScreenDump{
		MasterDetail: NewMasterDetail(dumpTitle, ""),
	}
}

// Init initializes the viewer.
func (s *ScreenDump) Init(ctx context.Context) {
	s.app = ctx.Value(ui.KeyApp).(*App)
	s.MasterDetail.Init(ctx)
	s.registerActions()

	table := s.masterPage()
	{
		table.SetBorderFocusColor(tcell.ColorSteelBlue)
		table.SetSelectedStyle(tcell.ColorWhite, tcell.ColorRoyalBlue, tcell.AttrNone)
		table.SetColorerFn(dumpColorer)
		table.SetActiveNS(resource.AllNamespaces)
		table.SetSortCol(table.NameColIndex(), 0, true)
		table.SelectRow(1, true)
	}
	s.Start()
	s.refresh()
}

// Start starts the directory watcher.
func (s *ScreenDump) Start() {
	var ctx context.Context
	ctx, s.cancelFn = context.WithCancel(context.Background())
	if err := s.watchDumpDir(ctx); err != nil {
		s.app.Flash().Errf("Unable to watch screen dumps directory %s", err)
	}
}

// Stop terminates the directory watcher.
func (s *ScreenDump) Stop() {
	if s.cancelFn != nil {
		s.cancelFn()
	}
}

// Name returns the component name.
func (s *ScreenDump) Name() string {
	return dumpTitle
}

func (s *ScreenDump) setEnterFn(enterFn)            {}
func (s *ScreenDump) setColorerFn(ui.ColorerFunc)   {}
func (s *ScreenDump) setDecorateFn(decorateFn)      {}
func (s *ScreenDump) setExtraActionsFn(ActionsFunc) {}

func (s *ScreenDump) refresh() {
	tv := s.masterPage()
	tv.Update(s.hydrate())
	tv.UpdateTitle()
}

func (s *ScreenDump) registerActions() {
	s.masterPage().AddActions(ui.KeyActions{
		tcell.KeyEsc:   ui.NewKeyAction("Back", s.app.PrevCmd, false),
		tcell.KeyEnter: ui.NewKeyAction("View", s.enterCmd, true),
		tcell.KeyCtrlD: ui.NewKeyAction("Delete", s.deleteCmd, true),
		tcell.KeyCtrlS: ui.NewKeyAction("Save", noopCmd, false),
	})
}

func (s *ScreenDump) getTitle() string {
	return dumpTitle
}

func (s *ScreenDump) enterCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msg("Dump enter!")
	tv := s.masterPage()
	if tv.SearchBuff().IsActive() {
		return tv.filterCmd(evt)
	}
	sel := tv.GetSelectedItem()
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
	sel := s.masterPage().GetSelectedItem()
	if sel == "" {
		return nil
	}

	dir := filepath.Join(config.K9sDumpDir, s.app.Config.K9s.CurrentCluster)
	showModal(s.Pages, fmt.Sprintf("Delete screen dump `%s?", sel), "table", func() {
		if err := os.Remove(filepath.Join(dir, sel)); err != nil {
			s.app.Flash().Errf("Unable to delete file %s", err)
			return
		}
		s.refresh()
		s.app.Flash().Infof("ScreenDump file %s deleted!", sel)
	})

	return nil
}

func (s *ScreenDump) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.cancelFn != nil {
		s.cancelFn()
	}
	s.SwitchToPage("table")

	return nil
}

func (s *ScreenDump) Hints() model.MenuHints {
	if s.CurrentPage() == nil {
		return nil
	}
	if c, ok := s.CurrentPage().Item.(model.Hinter); ok {
		return c.Hints()
	}

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

func (s *ScreenDump) resetTitle() {
	s.SetTitle(fmt.Sprintf(dumpTitleFmt, dumpTitle, s.masterPage().GetRowCount()-1))
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
				w.Close()
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
