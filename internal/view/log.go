package view

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
)

const (
	logTitle            = "logs"
	logMessage          = "Waiting for logs...\n"
	logFmt              = " Logs([hilite:bg:]%s[-:bg:-])[[green:bg:b]%s[-:bg:-]] "
	logCoFmt            = " Logs([hilite:bg:]%s:[hilite:bg:b]%s[-:bg:-])[[green:bg:b]%s[-:bg:-]] "
	defaultFlushTimeout = 50 * time.Millisecond
)

// Log represents a generic log viewer.
type Log struct {
	*tview.Flex

	app           *App
	logs          *Logger
	indicator     *LogIndicator
	ansiWriter    io.Writer
	model         *model.Log
	cancelFn      context.CancelFunc
	cancelUpdates bool
	mx            sync.Mutex
	logChan       dao.LogChan
	follow        bool
}

var _ model.Component = (*Log)(nil)

// NewLog returns a new viewer.
func NewLog(gvr client.GVR, opts *dao.LogOptions) *Log {
	l := Log{
		Flex:    tview.NewFlex(),
		logChan: make(dao.LogChan, 2),
		model:   model.NewLog(gvr, opts, defaultFlushTimeout),
		follow:  true,
	}

	return &l
}

// Init initializes the viewer.
func (l *Log) Init(ctx context.Context) (err error) {
	if l.app, err = extractApp(ctx); err != nil {
		return err
	}
	l.model.Configure(l.app.Config.K9s.Logger)

	l.SetBorder(true)
	l.SetDirection(tview.FlexRow)

	l.indicator = NewLogIndicator(l.app.Config, l.app.Styles, l.isContainerLogView())
	l.AddItem(l.indicator, 1, 1, false)
	if !l.model.HasDefaultContainer() {
		l.indicator.ToggleAllContainers()
	}
	l.indicator.Refresh()

	l.logs = NewLogger(l.app)
	if err = l.logs.Init(ctx); err != nil {
		return err
	}
	l.logs.SetBorderPadding(0, 0, 1, 1)
	l.logs.SetText("[orange::d]" + logMessage)
	l.logs.SetWrap(l.app.Config.K9s.Logger.TextWrap)
	l.logs.SetMaxLines(l.app.Config.K9s.Logger.BufferSize)

	l.ansiWriter = tview.ANSIWriter(l.logs, l.app.Styles.Views().Log.FgColor.String(), l.app.Styles.Views().Log.BgColor.String())
	l.AddItem(l.logs, 0, 1, true)
	l.bindKeys()

	l.StylesChanged(l.app.Styles)
	l.goFullScreen()

	l.model.Init(l.app.factory)
	l.updateTitle()

	l.model.ToggleShowTimestamp(l.app.Config.K9s.Logger.ShowTime)

	return nil
}

// InCmdMode checks if prompt is active.
func (l *Log) InCmdMode() bool {
	return l.logs.cmdBuff.InCmdMode()
}

// LogCanceled indicates no more logs are coming.
func (l *Log) LogCanceled() {
	log.Debug().Msgf("LOGS_CANCELED!!!")
	l.Flush([][]byte{[]byte("\n🏁 [red::b]Stream exited! No more logs...")})
}

// LogStop disables log flushes.
func (l *Log) LogStop() {
	log.Debug().Msgf("LOG_STOP!!!")
	l.mx.Lock()
	defer l.mx.Unlock()

	l.cancelUpdates = true
}

// LogResume resume log flushes.
func (l *Log) LogResume() {
	l.mx.Lock()
	defer l.mx.Unlock()

	log.Debug().Msgf("LOG_RESUME!!!")
	l.cancelUpdates = false
}

// LogCleared clears the logs.
func (l *Log) LogCleared() {
	l.app.QueueUpdateDraw(func() {
		l.logs.Clear()
	})
}

// LogFailed notifies an error occurred.
func (l *Log) LogFailed(err error) {
	l.app.QueueUpdateDraw(func() {
		l.app.Flash().Err(err)
		if l.logs.GetText(true) == logMessage {
			l.logs.Clear()
		}
		if _, err = l.ansiWriter.Write([]byte(tview.Escape(color.Colorize(err.Error(), color.Red)))); err != nil {
			log.Error().Err(err).Msgf("Writing log error")
		}
	})
}

// LogChanged updates the logs.
func (l *Log) LogChanged(lines [][]byte) {
	l.app.QueueUpdateDraw(func() {
		if l.logs.GetText(true) == logMessage {
			l.logs.Clear()
		}
		l.Flush(lines)
	})
}

// BufferCompleted indicates input was accepted.
func (l *Log) BufferCompleted(text, _ string) {
	l.model.Filter(text)
	l.updateTitle()
}

// BufferChanged indicates the buffer was changed.
func (l *Log) BufferChanged(_, _ string) {}

// BufferActive indicates the buff activity changed.
func (l *Log) BufferActive(state bool, k model.BufferKind) {
	l.app.BufferActive(state, k)
}

// StylesChanged reports skin changes.
func (l *Log) StylesChanged(s *config.Styles) {
	l.SetBackgroundColor(s.Views().Log.BgColor.Color())
	l.logs.SetTextColor(s.Views().Log.FgColor.Color())
	l.logs.SetBackgroundColor(s.Views().Log.BgColor.Color())
}

// GetModel returns the log model.
func (l *Log) GetModel() *model.Log {
	return l.model
}

// Hints returns a collection of menu hints.
func (l *Log) Hints() model.MenuHints {
	return l.logs.Actions().Hints()
}

// ExtraHints returns additional hints.
func (l *Log) ExtraHints() map[string]string {
	return nil
}

func (l *Log) getContext() context.Context {
	if l.cancelFn != nil {
		l.cancelFn()
	}
	ctx := context.Background()
	ctx, l.cancelFn = context.WithCancel(ctx)
	return ctx
}

// Start runs the component.
func (l *Log) Start() {
	log.Debug().Msgf("LOG_VIEW STARTED!!")

	l.model.Restart(l.getContext(), l.logChan, true)
	l.model.AddListener(l)
	l.app.Styles.AddListener(l)
	l.logs.cmdBuff.AddListener(l)
	l.logs.cmdBuff.AddListener(l.app.Prompt())
	l.updateTitle()
}

// Stop terminates the component.
func (l *Log) Stop() {
	log.Debug().Msgf("LOG_VIEW STOPPED!")
	l.model.RemoveListener(l)
	l.model.Stop()
	log.Debug().Msgf("CLOSING LOG_CHANNEL!!!")
	l.mx.Lock()
	{
		if l.cancelFn != nil {
			l.cancelFn()
			l.cancelFn = nil
		}
		if l.logChan != nil {
			close(l.logChan)
			l.logChan = nil
		}
	}
	l.mx.Unlock()
	l.app.Styles.RemoveListener(l)
	l.logs.cmdBuff.RemoveListener(l)
	l.logs.cmdBuff.RemoveListener(l.app.Prompt())
}

// Name returns the component name.
func (l *Log) Name() string { return logTitle }

func (l *Log) bindKeys() {
	l.logs.Actions().Set(ui.KeyActions{
		ui.Key0:         ui.NewKeyAction("tail", l.sinceCmd(-1), true),
		ui.Key1:         ui.NewKeyAction("head", l.head(), true),
		ui.Key2:         ui.NewKeyAction("1m", l.sinceCmd(60), true),
		ui.Key3:         ui.NewKeyAction("5m", l.sinceCmd(5*60), true),
		ui.Key4:         ui.NewKeyAction("15m", l.sinceCmd(15*60), true),
		ui.Key5:         ui.NewKeyAction("30m", l.sinceCmd(30*60), true),
		ui.Key6:         ui.NewKeyAction("1h", l.sinceCmd(60*60), true),
		tcell.KeyEnter:  ui.NewSharedKeyAction("Filter", l.filterCmd, false),
		tcell.KeyEscape: ui.NewKeyAction("Back", l.resetCmd, false),
		ui.KeyShiftC:    ui.NewKeyAction("Clear", l.clearCmd, true),
		ui.KeyM:         ui.NewKeyAction("Mark", l.markCmd, true),
		ui.KeyS:         ui.NewKeyAction("Toggle AutoScroll", l.toggleAutoScrollCmd, true),
		ui.KeyF:         ui.NewKeyAction("Toggle FullScreen", l.toggleFullScreenCmd, true),
		ui.KeyT:         ui.NewKeyAction("Toggle Timestamp", l.toggleTimestampCmd, true),
		ui.KeyW:         ui.NewKeyAction("Toggle Wrap", l.toggleTextWrapCmd, true),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", l.SaveCmd, true),
		ui.KeyC:         ui.NewKeyAction("Copy", l.cpCmd, true),
	})
	if l.model.HasDefaultContainer() {
		l.logs.Actions().Set(ui.KeyActions{
			ui.KeyA: ui.NewKeyAction("Toggle AllContainers", l.toggleAllContainers, true),
		})
	}
}

func (l *Log) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.logs.cmdBuff.IsActive() {
		if l.logs.cmdBuff.GetText() == "" {
			return l.app.PrevCmd(evt)
		}
	}

	l.logs.cmdBuff.Reset()
	l.logs.cmdBuff.SetActive(false)
	l.model.Filter(l.logs.cmdBuff.GetText())
	l.updateTitle()

	return nil
}

// SendStrokes (testing only!)
func (l *Log) SendStrokes(s string) {
	l.app.Prompt().SendStrokes(s)
}

// SendKeys (testing only!)
func (l *Log) SendKeys(kk ...tcell.Key) {
	for _, k := range kk {
		l.logs.keyboard(tcell.NewEventKey(k, ' ', tcell.ModNone))
	}
}

// Indicator returns the scroll mode viewer.
func (l *Log) Indicator() *LogIndicator {
	return l.indicator
}

func (l *Log) updateTitle() {
	sinceSeconds, since := l.model.SinceSeconds(), "tail"
	if sinceSeconds > 0 && sinceSeconds < 60*60 {
		since = fmt.Sprintf("%dm", sinceSeconds/60)
	}
	if sinceSeconds >= 60*60 {
		since = fmt.Sprintf("%dh", sinceSeconds/(60*60))
	}
	if l.model.IsHead() {
		since = "head"
	}

	var title string
	path, co := l.model.GetPath(), l.model.GetContainer()
	if co == "" {
		title = ui.SkinTitle(fmt.Sprintf(logFmt, path, since), l.app.Styles.Frame())
	} else {
		title = ui.SkinTitle(fmt.Sprintf(logCoFmt, path, co, since), l.app.Styles.Frame())
	}

	buff := l.logs.cmdBuff.GetText()
	if buff != "" {
		title += ui.SkinTitle(fmt.Sprintf(ui.SearchFmt, buff), l.app.Styles.Frame())
	}
	l.SetTitle(title)
}

// Logs returns the log viewer.
func (l *Log) Logs() *Logger {
	return l.logs
}

// EOL tracks end of lines.
var EOL = []byte{'\n'}

// Flush write logs to viewer.
func (l *Log) Flush(lines [][]byte) {
	defer func() {
		if l.cancelUpdates {
			l.cancelUpdates = false
		}
	}()

	if len(lines) == 0 || !l.indicator.AutoScroll() || l.cancelUpdates {
		return
	}
	for i := 0; i < len(lines); i++ {
		if l.cancelUpdates {
			break
		}
		_, _ = l.ansiWriter.Write(lines[i])
	}
	if l.follow {
		l.logs.ScrollToEnd()
	}
}

// ----------------------------------------------------------------------------
// Actions()...

func (l *Log) head() func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		log.Debug().Msgf("!!!!HEAD!!!!")
		l.cancelUpdates = true
		l.logs.Clear()
		l.model.Head(l.getContext(), l.logChan)
		l.updateTitle()

		return nil
	}
}

func (l *Log) sinceCmd(a int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		l.logs.Clear()
		l.model.SetSinceSeconds(l.getContext(), l.logChan, int64(a))
		l.updateTitle()

		return nil
	}
}

func (l *Log) toggleAllContainers(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}
	l.indicator.ToggleAllContainers()
	l.model.ToggleAllContainers(l.getContext(), l.logChan)
	l.updateTitle()

	return nil
}

func (l *Log) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !l.logs.cmdBuff.IsActive() {
		return evt
	}

	l.logs.cmdBuff.SetActive(false)
	l.model.Filter(l.logs.cmdBuff.GetText())
	l.updateTitle()

	return nil
}

// SaveCmd dumps the logs to file.
func (l *Log) SaveCmd(*tcell.EventKey) *tcell.EventKey {
	if path, err := saveData(l.app.Config.K9s.CurrentContext, l.model.GetPath(), l.logs.GetText(true)); err != nil {
		l.app.Flash().Err(err)
	} else {
		l.app.Flash().Infof("Log %s saved successfully!", path)
	}
	return nil
}

func (l *Log) cpCmd(*tcell.EventKey) *tcell.EventKey {
	l.app.Flash().Info("Content copied to clipboard...")
	if err := clipboard.WriteAll(l.logs.GetText(true)); err != nil {
		l.app.Flash().Err(err)
	}
	return nil
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0744)
}

func saveData(cluster, name, data string) (string, error) {
	dir := filepath.Join(config.K9sDumpDir, dao.SanitizeFilename(cluster))
	if err := ensureDir(dir); err != nil {
		return "", err
	}

	now := time.Now().UnixNano()
	fName := fmt.Sprintf("%s-%d.log", dao.SanitizeFilename(name), now)

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
	l.model.Clear()
	return nil
}

func (l *Log) markCmd(*tcell.EventKey) *tcell.EventKey {
	_, _, w, _ := l.GetRect()
	fmt.Fprintf(l.ansiWriter, "\n[white:-:b]%s[-:-:-]", strings.Repeat("─", w-4))
	l.follow = true

	return nil
}

func (l *Log) toggleTimestampCmd(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}

	l.indicator.ToggleTimestamp()
	l.model.ToggleShowTimestamp(l.indicator.showTime)
	l.indicator.Refresh()

	return nil
}

func (l *Log) toggleTextWrapCmd(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}

	l.indicator.ToggleTextWrap()
	l.logs.SetWrap(l.indicator.textWrap)
	l.indicator.Refresh()

	return nil
}

// ToggleAutoScrollCmd toggles autoscroll status.
func (l *Log) toggleAutoScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}

	l.indicator.ToggleAutoScroll()
	l.follow = l.indicator.AutoScroll()
	// if l.indicator.AutoScroll() {

	// 	// l.model.Restart(l.getContext(), l.logChan, false)
	// } else {
	// 	// l.model.Stop()
	// }
	l.indicator.Refresh()

	return nil
}

func (l *Log) toggleFullScreenCmd(evt *tcell.EventKey) *tcell.EventKey {
	if l.app.InCmdMode() {
		return evt
	}
	l.indicator.ToggleFullScreen()
	l.goFullScreen()
	l.indicator.Refresh()

	return nil
}

func (l *Log) goFullScreen() {
	l.SetFullScreen(l.indicator.FullScreen())
	l.Box.SetBorder(!l.indicator.FullScreen())
	if l.indicator.FullScreen() {
		l.logs.SetBorderPadding(0, 0, 0, 0)
	} else {
		l.logs.SetBorderPadding(0, 0, 1, 1)
	}
}

func (l *Log) isContainerLogView() bool {
	return l.model.HasDefaultContainer()
}
