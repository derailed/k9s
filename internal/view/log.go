package view

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	logTitle     = "logs"
	logBuffSize  = 100
	FlushTimeout = 200 * time.Millisecond

	logCoFmt = " Logs([fg:bg:]%s:[hilite:bg:b]%s[-:bg:-]) "
	logFmt   = " Logs([fg:bg:]%s) "
)

// Log represents a generic log viewer.
type Log struct {
	*tview.Flex

	app             *App
	logs            *Details
	scrollIndicator *AutoScrollIndicator
	ansiWriter      io.Writer
	path, container string
	cancelFn        context.CancelFunc
	previous        bool
	list            resource.List
}

var _ model.Component = &Log{}

// NewLog returns a new viewer.
func NewLog(path, co string, l resource.List, prev bool) *Log {
	return &Log{
		Flex:      tview.NewFlex(),
		path:      path,
		container: co,
		list:      l,
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
	l.SetBackgroundColor(config.AsColor(l.app.Styles.Views().Log.BgColor))
	l.SetBorderPadding(0, 0, 1, 1)
	l.SetDirection(tview.FlexRow)

	l.scrollIndicator = NewAutoScrollIndicator(l.app.Styles)
	l.AddItem(l.scrollIndicator, 1, 1, false)

	l.logs = NewDetails("")
	l.logs.SetBorder(false)
	l.logs.SetDynamicColors(true)
	l.logs.SetTextColor(config.AsColor(l.app.Styles.Views().Log.FgColor))
	l.logs.SetBackgroundColor(config.AsColor(l.app.Styles.Views().Log.BgColor))
	l.logs.SetWrap(true)
	l.logs.SetMaxBuffer(l.app.Config.K9s.LogBufferSize)
	if err = l.logs.Init(ctx); err != nil {
		return err
	}
	l.ansiWriter = tview.ANSIWriter(l.logs, l.app.Styles.Views().Log.FgColor, l.app.Styles.Views().Log.BgColor)
	l.AddItem(l.logs, 0, 1, true)
	l.bindKeys()
	l.logs.SetInputCapture(l.keyboard)

	return nil
}

// Refresh refreshes the viewer.
func (l *Log) Refresh() {}

// List returns the resource list.
func (l *Log) List() resource.List {
	return l.list
}

// App returns an app handle.
func (l *Log) App() *App {
	return l.app
}

// Hints returns a collection of menu hints.
func (l *Log) Hints() model.MenuHints {
	return l.Actions().Hints()
}

// Actions returns available actions.
func (l *Log) Actions() ui.KeyActions {
	return l.logs.actions
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
}

func (l *Log) Name() string { return logTitle }

func (l *Log) bindKeys() {
	l.logs.Actions().Set(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", l.backCmd, true),
		ui.KeyC:         ui.NewKeyAction("Clear", l.clearCmd, true),
		ui.KeyS:         ui.NewKeyAction("Toggle AutoScroll", l.ToggleAutoScrollCmd, true),
		ui.KeyG:         ui.NewKeyAction("Top", l.topCmd, false),
		ui.KeyShiftG:    ui.NewKeyAction("Bottom", l.bottomCmd, false),
		ui.KeyF:         ui.NewKeyAction("Up", l.pageUpCmd, false),
		ui.KeyB:         ui.NewKeyAction("Down", l.pageDownCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", l.SaveCmd, true),
	})
}

func (l *Log) doLoad() error {
	// BOZO!!
	// l.logs.Clear()
	// l.setTitle(l.path, l.container)

	// var ctx context.Context
	// ctx = context.WithValue(context.Background(), resource.IKey("informer"), l.app.informers.ActiveInformer())
	// ctx, l.cancelFn = context.WithCancel(ctx)

	// c := make(chan string, 10)
	// go l.updateLogs(ctx, c, logBuffSize)

	// res, ok := l.list.Resource().(resource.Tailable)
	// if !ok {
	// 	close(c)
	// 	return fmt.Errorf("Resource %T is not tailable", l.list.Resource())
	// }

	// if err := res.Logs(ctx, c, l.logOpts(l.path, l.container, l.previous)); err != nil {
	// 	l.cancelFn()
	// 	close(c)
	// 	return err
	// }

	return nil
}

func (l *Log) logOpts(path, co string, prevLogs bool) resource.LogOptions {
	ns, po := namespaced(path)
	return resource.LogOptions{
		Fqn: resource.Fqn{
			Namespace: ns,
			Name:      po,
			Container: co,
		},
		Lines:    int64(l.app.Config.K9s.LogRequestSize),
		Previous: prevLogs,
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

// ScrollIndicator returns the scroll mode viewer.
func (l *Log) ScrollIndicator() *AutoScrollIndicator {
	return l.scrollIndicator
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

func (l *Log) Logs() *Details {
	return l.logs
}

func (l *Log) log(lines string) {
	fmt.Fprintln(l.ansiWriter, tview.Escape(lines))
	log.Debug().Msgf("LOG LINES %d", l.logs.GetLineCount())
}

// Flush write logs to viewer.
func (l *Log) Flush(index int, buff []string) {
	if index == 0 || !l.scrollIndicator.AutoScroll() {
		return
	}
	l.log(strings.Join(buff[:index], "\n"))
	l.app.QueueUpdateDraw(func() {
		l.scrollIndicator.Refresh()
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

// ToggleAutoScrollCmd toggles auto scrolling of logs.
func (l *Log) ToggleAutoScrollCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.scrollIndicator.ToggleAutoScroll()
	l.scrollIndicator.Refresh()

	return nil
}

func (l *Log) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	return l.app.PrevCmd(evt)
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
