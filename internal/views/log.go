package views

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type logView struct {
	*tview.Flex

	app        *appView
	logs       *detailsView
	status     *statusView
	parent     masterView
	ansiWriter io.Writer
	autoScroll bool
	actions    keyActions
	path       string
}

func newLogView(title string, parent masterView) *logView {
	v := logView{Flex: tview.NewFlex(), app: parent.appView()}
	v.autoScroll = true
	v.parent = parent
	v.SetBorder(true)
	v.SetBackgroundColor(config.AsColor(parent.appView().styles.Style.Log.BgColor))
	v.SetBorderPadding(0, 0, 1, 1)
	v.logs = newDetailsView(parent.appView(), parent.backFn())
	{
		v.logs.SetBorder(false)
		v.logs.setCategory("Logs")
		v.logs.SetDynamicColors(true)
		v.logs.SetTextColor(config.AsColor(parent.appView().styles.Style.Log.FgColor))
		v.logs.SetBackgroundColor(config.AsColor(parent.appView().styles.Style.Log.BgColor))
		v.logs.SetWrap(true)
		v.logs.SetMaxBuffer(parent.appView().config.K9s.LogBufferSize)
	}
	v.ansiWriter = tview.ANSIWriter(v.logs)
	v.status = newStatusView(parent.appView())
	v.SetDirection(tview.FlexRow)
	v.AddItem(v.status, 1, 1, false)
	v.AddItem(v.logs, 0, 1, true)

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
	v.logs.SetInputCapture(v.keyboard)

	return &v
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

func (v *logView) log(lines fmt.Stringer) {
	v.logs.Clear()
	fmt.Fprintln(v.ansiWriter, lines.String())
}

func (v *logView) flush(index int, buff []string) {
	if index == 0 {
		return
	}
	v.logLine(strings.Join(buff[:index], "\n"))
	if v.autoScroll {
		v.app.QueueUpdateDraw(func() {
			v.update()
			v.logs.ScrollToEnd()
		})
	}
}

func (v *logView) update() {
	status := "Off"
	if v.autoScroll {
		status = "On"
	}
	v.status.update([]string{fmt.Sprintf("Autoscroll: %s", status)})
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	dir := filepath.Join(config.K9sDumpDir, v.app.config.K9s.CurrentCluster)
	if err := os.MkdirAll(dir, 0744); err != nil {
		log.Error().Err(err).Msgf("Mkdir K9s dump")
		return nil
	}

	now := time.Now().UnixNano()
	fName := fmt.Sprintf("%s-%d.log", strings.Replace(v.path, "/", "-", -1), now)

	path := filepath.Join(dir, fName)
	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	file, err := os.OpenFile(path, mod, 0644)
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		log.Error().Err(err).Msgf("LogFile create %s", path)
		return nil
	}

	if _, err := fmt.Fprintf(file, v.logs.GetText(true)); err != nil {
		log.Error().Err(err).Msgf("Log dump %s", v.path)
	}
	v.app.flash().infof("Log %s saved successfully!", path)
	log.Debug().Msgf("Log %s saved successfully!", path)

	return nil
}

func (v *logView) toggleScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.autoScroll = !v.autoScroll
	if v.autoScroll {
		v.app.flash().info("Autoscroll is on.")
		v.logs.ScrollToEnd()
	} else {
		v.logs.PageUp()
		v.app.flash().info("Autoscroll is off.")
	}
	v.update()

	return nil
}

func (v *logView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	return v.parent.backFn()(evt)
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
