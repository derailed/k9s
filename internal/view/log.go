package view

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal"
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
	logTitle    = "logs"
	logBuffSize = 100

	// FlushTimeout represents a duration between log flushes.
	FlushTimeout = 200 * time.Millisecond

	logCoFmt = " Logs([fg:bg:]%s:[hilite:bg:b]%s[-:bg:-]) "
	logFmt   = " Logs([fg:bg:]%s) "
)

// Log represents a generic log viewer.
type Log struct {
	*tview.Flex

	app             *App
	logs            *Details
	indicator       *LogIndicator
	ansiWriter      io.Writer
	path, container string
	cancelFn        context.CancelFunc
	previous        bool
	gvr             client.GVR
}

var _ model.Component = &Log{}

// NewLog returns a new viewer.
func NewLog(gvr client.GVR, path, co string, prev bool) *Log {
	return &Log{
		gvr:       gvr,
		Flex:      tview.NewFlex(),
		path:      path,
		container: co,
		previous:  prev,
	}
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

	l.indicator = NewLogIndicator(l.app.Styles)
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

	return nil
}

// StylesChanged reports skin changes.
func (l *Log) StylesChanged(s *config.Styles) {
	l.SetBackgroundColor(config.AsColor(s.Views().Log.BgColor))
	l.logs.SetTextColor(config.AsColor(s.Views().Log.FgColor))
	l.logs.SetBackgroundColor(config.AsColor(s.Views().Log.BgColor))
}

// Hints returns a collection of menu hints.
func (l *Log) Hints() model.MenuHints {
	return l.logs.Actions().Hints()
}

// Start runs the component.
func (l *Log) Start() {
	l.Stop()
	if err := l.doLoad(); err != nil {
		l.app.Flash().Err(err)
		l.log("😂 Doh! No logs are available at this time. Check again later on...")
		return
	}
	l.app.SetFocus(l)
}

// Stop terminates the component.
func (l *Log) Stop() {
	if l.cancelFn != nil {
		log.Debug().Msgf("<<<< Logger STOP!")
		l.cancelFn()
		l.cancelFn = nil
	}
	l.app.Styles.RemoveListener(l)
}

// Name returns the component name.
func (l *Log) Name() string { return logTitle }

func (l *Log) bindKeys() {
	l.logs.Actions().Set(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", l.app.PrevCmd, true),
		ui.KeyC:         ui.NewKeyAction("Clear", l.clearCmd, true),
		ui.KeyS:         ui.NewKeyAction("Toggle AutoScroll", l.toggleAutoScrollCmd, true),
		ui.KeyG:         ui.NewKeyAction("Top", l.topCmd, false),
		ui.KeyShiftF:    ui.NewKeyAction("FullScreen", l.fullScreenCmd, true),
		ui.KeyW:         ui.NewKeyAction("Toggle Wrap", l.textWrapCmd, true),
		ui.KeyShiftG:    ui.NewKeyAction("Bottom", l.bottomCmd, false),
		ui.KeyF:         ui.NewKeyAction("Up", l.pageUpCmd, false),
		ui.KeyB:         ui.NewKeyAction("Down", l.pageDownCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", l.SaveCmd, true),
	})
}

func (l *Log) doLoad() error {
	l.logs.Clear()
	l.setTitle(l.path, l.container)

	var ctx context.Context
	ctx = context.WithValue(context.Background(), internal.KeyFactory, l.app.factory)
	ctx, l.cancelFn = context.WithCancel(ctx)

	c := make(chan string, 10)
	go l.updateLogs(ctx, c, logBuffSize)

	accessor, err := dao.AccessorFor(l.app.factory, l.gvr)
	if err != nil {
		return err
	}
	logger, ok := accessor.(dao.Loggable)
	if !ok {
		return fmt.Errorf("Resource %s is not tailable", l.gvr)
	}

	if err := logger.TailLogs(ctx, c, l.logOpts(l.path, l.container, l.previous)); err != nil {
		l.cancelFn()
		close(c)
		return err
	}

	return nil
}

func (l *Log) logOpts(path, co string, prevLogs bool) dao.LogOptions {
	return dao.LogOptions{
		Path:      path,
		Container: co,
		Lines:     int64(l.app.Config.K9s.LogRequestSize),
		Previous:  prevLogs,
	}
}

func (l *Log) updateLogs(ctx context.Context, c <-chan string, buffSize int) {
	defer func() {
		log.Debug().Msgf("updateLogs view bailing out!")
	}()
	buff, index := make([]string, buffSize), 0
	for {
		select {
		case line, ok := <-c:
			if !ok {
				log.Debug().Msgf("Closed channel detected. Bailing out...")
				l.Flush(index, buff)
				return
			}
			if index < buffSize {
				buff[index] = line
				index++
				continue
			}
			l.Flush(index, buff)
			index = 0
			buff[index] = line
			index++
		case <-time.After(FlushTimeout):
			l.Flush(index, buff)
			index = 0
		case <-ctx.Done():
			return
		}
	}
}

// Indicator returns the scroll mode viewer.
func (l *Log) Indicator() *LogIndicator {
	return l.indicator
}

func (l *Log) setTitle(path, co string) {
	var fmat string
	if co == "" {
		fmat = ui.SkinTitle(fmt.Sprintf(logFmt, path), l.app.Styles.Frame())
	} else {
		fmat = ui.SkinTitle(fmt.Sprintf(logCoFmt, path, co), l.app.Styles.Frame())
	}
	l.path = path
	l.SetTitle(fmat)
}

func (l *Log) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}
	if m, ok := l.logs.Actions()[key]; ok {
		log.Debug().Msgf(">> LogView handled %s", tcell.KeyNames[key])
		return m.Action(evt)
	}

	return evt
}

// Logs returns the log viewer.
func (l *Log) Logs() *Details {
	return l.logs
}

func (l *Log) log(lines string) {
	fmt.Fprintln(l.ansiWriter, tview.Escape(lines))
	log.Debug().Msgf("LOG LINES %d", l.logs.GetLineCount())
}

// Flush write logs to viewer.
func (l *Log) Flush(index int, buff []string) {
	if index == 0 || !l.indicator.AutoScroll() {
		return
	}
	l.log(strings.Join(buff[:index], "\n"))
	l.app.QueueUpdateDraw(func() {
		l.indicator.Refresh()
		l.logs.ScrollToEnd()
	})
}

// ----------------------------------------------------------------------------
// Actions()...

// SaveCmd dumps the logs to file.
func (l *Log) SaveCmd(evt *tcell.EventKey) *tcell.EventKey {
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

func (l *Log) textWrapCmd(*tcell.EventKey) *tcell.EventKey {
	l.indicator.ToggleTextWrap()
	l.logs.SetWrap(l.indicator.textWrap)
	return nil
}

func (l *Log) toggleAutoScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.indicator.ToggleAutoScroll()
	return nil
}

func (l *Log) fullScreenCmd(*tcell.EventKey) *tcell.EventKey {
	l.indicator.ToggleFullScreen()
	sidePadding := 1
	if l.indicator.FullScreen() {
		sidePadding = 0
	}
	l.SetFullScreen(l.indicator.FullScreen())
	l.Box.SetBorder(!l.indicator.FullScreen())
	l.Flex.SetBorderPadding(0, 0, sidePadding, sidePadding)

	return nil
}
