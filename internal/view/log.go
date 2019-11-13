package view

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type logFrame struct {
	*tview.Flex

	app     *App
	actions ui.KeyActions
	backFn  ui.ActionHandler
}

func newLogFrame(app *App, backFn ui.ActionHandler) *logFrame {
	f := logFrame{
		Flex:    tview.NewFlex(),
		app:     app,
		backFn:  backFn,
		actions: make(ui.KeyActions),
	}
	f.SetBorder(true)
	f.SetBackgroundColor(config.AsColor(app.Styles.Views().Log.BgColor))
	f.SetBorderPadding(0, 0, 1, 1)
	f.SetDirection(tview.FlexRow)

	return &f
}

type Log struct {
	*logFrame

	logs       *Details
	status     *statusView
	ansiWriter io.Writer
	autoScroll int32
	path       string
}

func NewLog(_ string, app *App, backFn ui.ActionHandler) *Log {
	l := Log{
		logFrame:   newLogFrame(app, backFn),
		autoScroll: 1,
	}

	l.logs = NewDetails(app, backFn)
	{
		l.logs.SetBorder(false)
		l.logs.setCategory("Logs")
		l.logs.SetDynamicColors(true)
		l.logs.SetTextColor(config.AsColor(app.Styles.Views().Log.FgColor))
		l.logs.SetBackgroundColor(config.AsColor(app.Styles.Views().Log.BgColor))
		l.logs.SetWrap(true)
		l.logs.SetMaxBuffer(app.Config.K9s.LogBufferSize)
	}
	l.ansiWriter = tview.ANSIWriter(l.logs, app.Styles.Views().Log.FgColor, app.Styles.Views().Log.BgColor)
	l.status = newStatusView(app.Styles)
	l.AddItem(l.status, 1, 1, false)
	l.AddItem(l.logs, 0, 1, true)

	l.bindKeys()
	l.logs.SetInputCapture(l.keyboard)

	return &l
}

func (l *Log) bindKeys() {
	l.actions = ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", l.backCmd, true),
		ui.KeyC:         ui.NewKeyAction("Clear", l.clearCmd, true),
		ui.KeyS:         ui.NewKeyAction("Toggle AutoScroll", l.toggleScrollCmd, true),
		ui.KeyG:         ui.NewKeyAction("Top", l.topCmd, false),
		ui.KeyShiftG:    ui.NewKeyAction("Bottom", l.bottomCmd, false),
		ui.KeyF:         ui.NewKeyAction("Up", l.pageUpCmd, false),
		ui.KeyB:         ui.NewKeyAction("Down", l.pageDownCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", l.saveCmd, true),
	}
}

func (l *Log) setTitle(path, co string) {
	var fmat string
	if co == "" {
		fmat = skinTitle(fmt.Sprintf(logFmt, path), l.app.Styles.Frame())
	} else {
		fmat = skinTitle(fmt.Sprintf(logCoFmt, path, co), l.app.Styles.Frame())
	}
	l.path = path
	l.SetTitle(fmat)
}

// Hints show action hints
func (l *Log) Hints() model.MenuHints {
	return l.actions.Hints()
}

func (l *Log) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}
	if m, ok := l.actions[key]; ok {
		log.Debug().Msgf(">> LogView handled %s", tcell.KeyNames[key])
		return m.Action(evt)
	}

	return evt
}

func (l *Log) log(lines string) {
	fmt.Fprintln(l.ansiWriter, tview.Escape(lines))
	log.Debug().Msgf("LOG LINES %d", l.logs.GetLineCount())
}

func (l *Log) flush(index int, buff []string) {
	if index == 0 {
		return
	}

	if atomic.LoadInt32(&l.autoScroll) == 1 {
		l.log(strings.Join(buff[:index], "\n"))
		l.app.QueueUpdateDraw(func() {
			l.updateIndicator()
			l.logs.ScrollToEnd()
		})
	}
}

func (l *Log) updateIndicator() {
	status := "Off"
	if l.autoScroll == 1 {
		status = "On"
	}
	l.status.update([]string{fmt.Sprintf("Autoscroll: %s", status)})
}

// ----------------------------------------------------------------------------
// Actions...

func (l *Log) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveData(l.app.Config.K9s.CurrentCluster, l.path, l.logs.GetText(true)); err != nil {
		l.app.Flash().Err(err)
	} else {
		l.app.Flash().Infof("Log %s saved successfully!", path)
	}
	return nil
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0744)
}

func saveData(cluster, name, data string) (string, error) {
	dir := filepath.Join(config.K9sDumpDir, cluster)
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	now := time.Now().UnixNano()
	fName := fmt.Sprintf("%s-%d.log", strings.Replace(name, "/", "-", -1), now)

	path := filepath.Join(dir, fName)
	mod := os.O_CREATE | os.O_WRONLY
	file, err := os.OpenFile(path, mod, 0644)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		log.Error().Err(err).Msgf("LogFile create %s", path)
		return "", nil
	}
	if _, err := fmt.Fprintf(file, data); err != nil {
		return "", err
	}

	return path, nil
}

func (l *Log) toggleScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	if atomic.LoadInt32(&l.autoScroll) == 0 {
		atomic.StoreInt32(&l.autoScroll, 1)
	} else {
		atomic.StoreInt32(&l.autoScroll, 0)
	}

	if atomic.LoadInt32(&l.autoScroll) == 1 {
		l.app.Flash().Info("Autoscroll is on.")
		l.logs.ScrollToEnd()
	} else {
		l.logs.LineUp()
		l.app.Flash().Info("Autoscroll is off.")
	}
	l.updateIndicator()

	return nil
}

func (l *Log) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	return l.backFn(evt)
}

func (l *Log) topCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.app.Flash().Info("Top of logs...")
	l.logs.ScrollToBeginning()
	return nil
}

func (l *Log) bottomCmd(*tcell.EventKey) *tcell.EventKey {
	l.app.Flash().Info("Bottom of logs...")
	l.logs.ScrollToEnd()
	return nil
}

func (l *Log) pageUpCmd(*tcell.EventKey) *tcell.EventKey {
	if l.logs.PageUp() {
		l.app.Flash().Info("Reached Top ...")
	}
	return nil
}

func (l *Log) pageDownCmd(*tcell.EventKey) *tcell.EventKey {
	if l.logs.PageDown() {
		l.app.Flash().Info("Reached Bottom ...")
	}
	return nil
}

func (l *Log) clearCmd(*tcell.EventKey) *tcell.EventKey {
	l.app.Flash().Info("Clearing logs...")
	l.logs.Clear()
	l.logs.ScrollTo(0, 0)
	return nil
}
