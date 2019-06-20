package views

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type (
	logFrame struct {
		*tview.Flex
		app     *appView
		actions keyActions
		backFn  actionHandler
	}

	logView struct {
		*logFrame
		logs       *detailsView
		status     *statusView
		ansiWriter io.Writer
		autoScroll int32
		path       string
	}
)

func newLogFrame(app *appView, backFn actionHandler) *logFrame {
	f := logFrame{
		Flex:    tview.NewFlex(),
		app:     app,
		backFn:  backFn,
		actions: make(keyActions),
	}
	f.SetBorder(true)
	f.SetBackgroundColor(config.AsColor(app.styles.Views().Log.BgColor))
	f.SetBorderPadding(0, 0, 1, 1)
	f.SetDirection(tview.FlexRow)

	return &f
}

func newLogView(_ string, app *appView, backFn actionHandler) *logView {
	v := logView{
		logFrame:   newLogFrame(app, backFn),
		autoScroll: 1,
	}

	v.logs = newDetailsView(app, backFn)
	{
		v.logs.SetBorder(false)
		v.logs.setCategory("Logs")
		v.logs.SetDynamicColors(true)
		v.logs.SetTextColor(config.AsColor(app.styles.Views().Log.FgColor))
		v.logs.SetBackgroundColor(config.AsColor(app.styles.Views().Log.BgColor))
		v.logs.SetWrap(true)
		v.logs.SetMaxBuffer(app.config.K9s.LogBufferSize)
	}
	v.ansiWriter = tview.ANSIWriter(v.logs)
	v.status = newStatusView(app.styles)
	v.AddItem(v.status, 1, 1, false)
	v.AddItem(v.logs, 0, 1, true)

	v.bindKeys()
	v.logs.SetInputCapture(v.keyboard)

	return &v
}

func (v *logView) bindKeys() {
	v.actions = keyActions{
		tcell.KeyEscape: newKeyAction("Back", v.backCmd, true),
		KeyC:            newKeyAction("Clear", v.clearCmd, true),
		KeyS:            newKeyAction("Toggle AutoScroll", v.toggleScrollCmd, true),
		KeyG:            newKeyAction("Top", v.topCmd, false),
		KeyShiftG:       newKeyAction("Bottom", v.bottomCmd, false),
		KeyF:            newKeyAction("Up", v.pageUpCmd, false),
		KeyB:            newKeyAction("Down", v.pageDownCmd, false),
		tcell.KeyCtrlS:  newKeyAction("Save", v.saveCmd, true),
	}
}

func (v *logView) setTitle(path, co string) {
	var fmat string
	if co == "" {
		fmat = skinTitle(fmt.Sprintf(logFmt, path), v.app.styles.Frame())
	} else {
		fmat = skinTitle(fmt.Sprintf(logCoFmt, path, co), v.app.styles.Frame())
	}
	v.path = path
	v.SetTitle(fmat)
}

// Hints show action hints
func (v *logView) hints() hints {
	return v.actions.toHints()
}

func (v *logView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}
	if m, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> LogView handled %s", tcell.KeyNames[key])
		return m.action(evt)
	}

	return evt
}

func (v *logView) logLine(line string) {
	fmt.Fprintln(v.ansiWriter, tview.Escape(line))
}

func (v *logView) flush(index int, buff []string) {
	if index == 0 {
		return
	}
	v.logLine(strings.Join(buff[:index], "\n"))
	if atomic.LoadInt32(&v.autoScroll) == 1 {
		v.app.QueueUpdateDraw(func() {
			v.update()
			v.logs.ScrollToEnd()
		})
	}
}

func (v *logView) update() {
	status := "Off"
	if v.autoScroll == 1 {
		status = "On"
	}
	v.status.update([]string{fmt.Sprintf("Autoscroll: %s", status)})
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveData(v.app.config.K9s.CurrentCluster, v.path, v.logs.GetText(true)); err != nil {
		v.app.flash().err(err)
	} else {
		v.app.flash().infof("Log %s saved successfully!", path)
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

func (v *logView) toggleScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	if atomic.LoadInt32(&v.autoScroll) == 0 {
		atomic.StoreInt32(&v.autoScroll, 1)
	} else {
		atomic.StoreInt32(&v.autoScroll, 0)
	}

	if atomic.LoadInt32(&v.autoScroll) == 1 {
		v.app.flash().info("Autoscroll is on.")
		v.logs.ScrollToEnd()
	} else {
		v.logs.LineUp()
		v.app.flash().info("Autoscroll is off.")
	}
	v.update()

	return nil
}

func (v *logView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	return v.backFn(evt)
}

func (v *logView) topCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.flash().info("Top of logs...")
	v.logs.ScrollToBeginning()
	return nil
}

func (v *logView) bottomCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.flash().info("Bottom of logs...")
	v.logs.ScrollToEnd()
	return nil
}

func (v *logView) pageUpCmd(*tcell.EventKey) *tcell.EventKey {
	if v.logs.PageUp() {
		v.app.flash().info("Reached Top ...")
	}
	return nil
}

func (v *logView) pageDownCmd(*tcell.EventKey) *tcell.EventKey {
	if v.logs.PageDown() {
		v.app.flash().info("Reached Bottom ...")
	}
	return nil
}

func (v *logView) clearCmd(*tcell.EventKey) *tcell.EventKey {
	v.app.flash().info("Clearing logs...")
	v.logs.Clear()
	v.logs.ScrollTo(0, 0)
	return nil
}
