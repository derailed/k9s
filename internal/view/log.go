package view

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	logTitle   = "logs"
	logMessage = "Waiting for logs..."
	logCoFmt   = " Logs([fg:bg:]%s:[hilite:bg:b]%s[-:bg:-]) "
	logFmt     = " Logs([fg:bg:]%s) "

	// BOZO!! Canned! Need config tail line counts!
	tailLineCount  = 1_000
	defaultTimeout = 200 * time.Millisecond
)

// Log represents a generic log viewer.
type Log struct {
	*tview.Flex

	app        *App
	logs       *Details
	indicator  *LogIndicator
	ansiWriter io.Writer
	cmdBuff    *ui.CmdBuff
	model      *model.Log
}

var _ model.Component = (*Log)(nil)

// NewLog returns a new viewer.
func NewLog(gvr client.GVR, path, co string, prev bool) *Log {
	l := Log{
		Flex:    tview.NewFlex(),
		cmdBuff: ui.NewCmdBuff('/', ui.FilterBuff),
		model:   model.NewLog(gvr, logMessage, buildLogOpts(path, co, prev, tailLineCount), defaultTimeout),
	}

	return &l
}

// Init initialiazes the viewer.
func (l *Log) Init(ctx context.Context) (err error) {
	log.Debug().Msgf(">>> Logs INIT")
	if l.app, err = extractApp(ctx); err != nil {
		return err
	}
	l.SetBorder(true)
	l.SetBorderPadding(0, 0, 1, 1)
	l.SetDirection(tview.FlexRow)

	l.indicator = NewLogIndicator(l.app.Config, l.app.Styles)
	l.AddItem(l.indicator, 1, 1, false)

	l.logs = NewDetails(l.app, "", "")
	if err = l.logs.Init(ctx); err != nil {
		return err
	}
	l.logs.SetWrap(false)
	l.logs.SetMaxBuffer(l.app.Config.K9s.LogBufferSize)

	l.ansiWriter = tview.ANSIWriter(l.logs, l.app.Styles.Views().Log.FgColor, l.app.Styles.Views().Log.BgColor)
	l.AddItem(l.logs, 0, 1, true)
	l.bindKeys()
	l.logs.SetInputCapture(l.keyboard)

	l.StylesChanged(l.app.Styles)
	l.app.Styles.AddListener(l)
	l.goFullScreen()

	l.model.Init(l.app.factory)
	l.model.AddListener(l)
	l.updateTitle()

	l.cmdBuff.AddListener(l.app.Cmd())
	l.cmdBuff.AddListener(l)

	return nil
}

// LogCleared clears the logs.
func (l *Log) LogCleared() {
	log.Debug().Msgf("LOG-CLEARED")
	l.app.QueueUpdateDraw(func() {
		l.logs.Clear()
		l.logs.ScrollTo(0, 0)
	})
}

// LogErrored notifies an error occurred.
func (l *Log) LogFailed(err error) {
	l.app.Flash().Err(err)
}

// LogsChanged updates the logs.
func (l *Log) LogChanged(lines []string) {
	log.Debug().Msgf("LOG-CHANGED %d", len(lines))
	l.app.QueueUpdateDraw(func() {
		l.Flush(lines)
	})
}

// BufferChanged indicates the buffer was changed.
func (l *Log) BufferChanged(s string) {}

// BufferActive indicates the buff activity changed.
func (l *Log) BufferActive(state bool, k ui.BufferKind) {
	l.app.BufferActive(state, k)
}

// StylesChanged reports skin changes.
func (l *Log) StylesChanged(s *config.Styles) {
	l.SetBackgroundColor(config.AsColor(s.Views().Log.BgColor))
	l.logs.SetTextColor(config.AsColor(s.Views().Log.FgColor))
	l.logs.SetBackgroundColor(config.AsColor(s.Views().Log.BgColor))
}

// GetModel returns the log model.
func (l *Log) GetModel() *model.Log {
	return l.model
}

// Hints returns a collection of menu hints.
func (l *Log) Hints() model.MenuHints {
	return l.logs.Actions().Hints()
}

// Start runs the component.
func (l *Log) Start() {
	l.model.Start()
	l.app.SetFocus(l)
}

// Stop terminates the component.
func (l *Log) Stop() {
	l.model.Stop()
	l.model.RemoveListener(l)
	l.app.Styles.RemoveListener(l)
	l.cmdBuff.RemoveListener(l)
	l.cmdBuff.RemoveListener(l.app.Cmd())
}

// Name returns the component name.
func (l *Log) Name() string { return logTitle }

func (l *Log) bindKeys() {
	l.logs.Actions().Set(ui.KeyActions{
		tcell.KeyEnter:      ui.NewSharedKeyAction("Filter", l.filterCmd, false),
		tcell.KeyEscape:     ui.NewKeyAction("Back", l.resetCmd, true),
		ui.KeyC:             ui.NewKeyAction("Clear", l.clearCmd, true),
		ui.KeyS:             ui.NewKeyAction("Toggle AutoScroll", l.ToggleAutoScrollCmd, true),
		ui.KeyF:             ui.NewKeyAction("FullScreen", l.fullScreenCmd, true),
		ui.KeyW:             ui.NewKeyAction("Toggle Wrap", l.textWrapCmd, true),
		tcell.KeyCtrlS:      ui.NewKeyAction("Save", l.SaveCmd, true),
		ui.KeySlash:         ui.NewSharedKeyAction("Filter Mode", l.activateCmd, false),
		tcell.KeyCtrlU:      ui.NewSharedKeyAction("Clear Filter", l.clearCmd, false),
		tcell.KeyBackspace2: ui.NewSharedKeyAction("Erase", l.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewSharedKeyAction("Erase", l.eraseCmd, false),
		tcell.KeyDelete:     ui.NewSharedKeyAction("Erase", l.eraseCmd, false),
	})
}

func (l *Log) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyUp || key == tcell.KeyDown {
		return evt
	}
	if key == tcell.KeyRune {
		if l.cmdBuff.IsActive() {
			l.cmdBuff.Add(evt.Rune())
			if err := l.model.Filter(l.cmdBuff.String()); err != nil {
				l.app.Flash().Err(err)
			}
			l.updateTitle()
			return nil
		}
		key = extractKey(evt)
	}

	if a, ok := l.logs.Actions()[key]; ok {
		return a.Action(evt)
	}

	return evt
}

// Indicator returns the scroll mode viewer.
func (l *Log) Indicator() *LogIndicator {
	return l.indicator
}

func (l *Log) updateTitle() {
	var fmat string
	path, co := l.model.GetPath(), l.model.GetContainer()
	if co == "" {
		fmat = ui.SkinTitle(fmt.Sprintf(logFmt, path), l.app.Styles.Frame())
	} else {
		fmat = ui.SkinTitle(fmt.Sprintf(logCoFmt, path, co), l.app.Styles.Frame())
	}

	buff := l.cmdBuff.String()
	if buff != "" {
		fmat += ui.SkinTitle(fmt.Sprintf(ui.SearchFmt, buff), l.app.Styles.Frame())
	}
	l.SetTitle(fmat)
}

// Logs returns the log viewer.
func (l *Log) Logs() *Details {
	return l.logs
}

func (l *Log) write(lines string) {
	fmt.Fprintln(l.ansiWriter, tview.Escape(lines))
	log.Debug().Msgf("LOG LINES %d", l.logs.GetLineCount())
}

// Flush write logs to viewer.
func (l *Log) Flush(lines []string) {
	if !l.indicator.AutoScroll() {
		return
	}
	l.write(strings.Join(lines, "\n"))
	l.indicator.Refresh()
	l.logs.ScrollToEnd()
}

// ----------------------------------------------------------------------------
// Actions()...

func (l *Log) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.cmdBuff.IsActive() {
		return evt
	}
	l.cmdBuff.SetActive(false)
	if err := l.model.Filter(l.cmdBuff.String()); err != nil {
		l.app.Flash().Err(err)
	}
	l.updateTitle()

	return nil
}

func (l *Log) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}
	l.app.Flash().Info("Filter mode activated.")
	l.cmdBuff.SetActive(true)

	return nil
}

func (l *Log) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.cmdBuff.IsActive() {
		return nil
	}
	l.cmdBuff.Delete()
	if err := l.model.Filter(l.cmdBuff.String()); err != nil {
		l.app.Flash().Err(err)
	}
	l.updateTitle()

	return nil
}

func (l *Log) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.cmdBuff.InCmdMode() {
		l.cmdBuff.Reset()
		return l.app.PrevCmd(evt)
	}

	if l.cmdBuff.String() != "" {
		l.model.ClearFilter()
	}
	l.app.Flash().Info("Clearing filter...")
	l.cmdBuff.SetActive(false)
	l.cmdBuff.Reset()
	l.updateTitle()

	return nil
}

// SaveCmd dumps the logs to file.
func (l *Log) SaveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveData(l.app.Config.K9s.CurrentCluster, l.model.GetPath(), l.logs.GetText(true)); err != nil {
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
	file, err := os.OpenFile(path, mod, 0600)
	if err != nil {
		log.Error().Err(err).Msgf("LogFile create %s", path)
		return "", nil
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Error().Err(err).Msg("Closing Log file")
		}
	}()
	if _, err := file.Write([]byte(data)); err != nil {
		return "", err
	}

	return path, nil
}

func (l *Log) clearCmd(*tcell.EventKey) *tcell.EventKey {
	l.app.Flash().Info("Clearing logs...")
	l.model.Clear()
	return nil
}

func (l *Log) textWrapCmd(*tcell.EventKey) *tcell.EventKey {
	l.indicator.ToggleTextWrap()
	l.logs.SetWrap(l.indicator.textWrap)
	return nil
}

func (l *Log) ToggleAutoScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.indicator.ToggleAutoScroll()
	return nil
}

func (l *Log) fullScreenCmd(*tcell.EventKey) *tcell.EventKey {
	l.indicator.ToggleFullScreen()
	l.goFullScreen()
	return nil
}

func (l *Log) goFullScreen() {
	sidePadding := 1
	if l.indicator.FullScreen() {
		sidePadding = 0
	}
	l.SetFullScreen(l.indicator.FullScreen())
	l.Box.SetBorder(!l.indicator.FullScreen())
	l.Flex.SetBorderPadding(0, 0, sidePadding, sidePadding)
}

// ----------------------------------------------------------------------------
// Helpers...

// AsKey converts rune to keyboard key.,
func extractKey(evt *tcell.EventKey) tcell.Key {
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}

func buildLogOpts(path, co string, prevLogs bool, tailLineCount int) dao.LogOptions {
	return dao.LogOptions{
		Path:      path,
		Container: co,
		Lines:     int64(tailLineCount),
		Previous:  prevLogs,
	}
}
