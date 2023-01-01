package view

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

const liveViewTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

// LiveView represents a live text viewer.
type LiveView struct {
	*tview.Flex

	title                     string
	model                     model.ResourceViewer
	text                      *tview.TextView
	actions                   ui.KeyActions
	app                       *App
	cmdBuff                   *model.FishBuff
	currentRegion, maxRegions int
	cancel                    context.CancelFunc
	fullScreen                bool
	managedField              bool
	autoRefresh               bool
}

// NewLiveView returns a live viewer.
func NewLiveView(app *App, title string, m model.ResourceViewer) *LiveView {
	v := LiveView{
		Flex:          tview.NewFlex(),
		text:          tview.NewTextView(),
		app:           app,
		title:         title,
		actions:       make(ui.KeyActions),
		currentRegion: 0,
		maxRegions:    0,
		cmdBuff:       model.NewFishBuff('/', model.FilterBuffer),
		model:         m,
	}
	v.AddItem(v.text, 0, 1, true)

	return &v
}

// Init initializes the viewer.
func (v *LiveView) Init(_ context.Context) error {
	if v.title != "" {
		v.SetBorder(true)
	}
	v.text.SetScrollable(true).SetWrap(true).SetRegions(true)
	v.text.SetDynamicColors(true)
	v.text.SetHighlightColor(tcell.ColorOrange)
	v.SetTitleColor(tcell.ColorAqua)
	v.SetInputCapture(v.keyboard)
	v.SetBorderPadding(0, 0, 1, 1)
	v.updateTitle()

	v.app.Styles.AddListener(v)
	v.StylesChanged(v.app.Styles)

	v.app.Prompt().SetModel(v.cmdBuff)
	v.cmdBuff.AddListener(v)

	v.bindKeys()
	v.SetInputCapture(v.keyboard)
	v.model.AddListener(v)

	return nil
}

// InCmdMode checks if prompt is active.
func (v *LiveView) InCmdMode() bool {
	return v.cmdBuff.InCmdMode()
}

// ResourceFailed notifies when their is an issue.
func (v *LiveView) ResourceFailed(err error) {
	v.text.SetTextAlign(tview.AlignCenter)
	x, _, w, _ := v.GetRect()
	v.text.SetText(cowTalk(err.Error(), x+w))
}

// ResourceChanged notifies when the filter changes.
func (v *LiveView) ResourceChanged(lines []string, matches fuzzy.Matches) {
	v.app.QueueUpdateDraw(func() {
		v.text.SetTextAlign(tview.AlignLeft)
		v.maxRegions = len(matches)
		ll := make([]string, len(lines))
		copy(ll, lines)
		for i, m := range matches {
			loc, line := m.MatchedIndexes, ll[m.Index]
			ll[m.Index] = line[:loc[0]] + `<<<"search_` + strconv.Itoa(i) + `">>>` + line[loc[0]:loc[1]] + `<<<"">>>` + line[loc[1]:]
		}

		if v.text.GetText(true) == "" {
			v.text.ScrollToBeginning()
		}

		v.text.SetText(colorizeYAML(v.app.Styles.Views().Yaml, strings.Join(ll, "\n")))
		v.text.Highlight()
		if v.currentRegion < v.maxRegions {
			v.text.Highlight("search_" + strconv.Itoa(v.currentRegion))
			v.text.ScrollToHighlight()
		}
		v.updateTitle()
	})
}

// BufferChanged indicates the buffer was changed.
func (v *LiveView) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (v *LiveView) BufferCompleted(text, _ string) {
	v.model.Filter(text)
}

// BufferActive indicates the buff activity changed.
func (v *LiveView) BufferActive(state bool, k model.BufferKind) {
	v.app.BufferActive(state, k)
}

func (v *LiveView) bindKeys() {
	v.actions.Set(ui.KeyActions{
		tcell.KeyEnter:  ui.NewSharedKeyAction("Filter", v.filterCmd, false),
		tcell.KeyEscape: ui.NewKeyAction("Back", v.resetCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", v.saveCmd, false),
		ui.KeyC:         ui.NewKeyAction("Copy", cpCmd(v.app.Flash(), v.text), true),
		ui.KeyF:         ui.NewKeyAction("Toggle FullScreen", v.toggleFullScreenCmd, true),
		ui.KeyR:         ui.NewKeyAction("Toggle Auto-Refresh", v.toggleRefreshCmd, true),
		ui.KeyN:         ui.NewKeyAction("Next Match", v.nextCmd, true),
		ui.KeyShiftN:    ui.NewKeyAction("Prev Match", v.prevCmd, true),
		ui.KeySlash:     ui.NewSharedKeyAction("Filter Mode", v.activateCmd, false),
		tcell.KeyDelete: ui.NewSharedKeyAction("Erase", v.eraseCmd, false),
	})

	if v.title == "YAML" {
		v.actions.Add(ui.KeyActions{
			ui.KeyM: ui.NewKeyAction("Toggle ManagedFields", v.toggleManagedCmd, true),
		})
	}
}

// ToggleRefreshCmd is used for pausing the refreshing of data on config map and secrets.
func (v *LiveView) toggleRefreshCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.autoRefresh = !v.autoRefresh
	if v.autoRefresh {
		v.Start()
		v.app.Flash().Info("Auto-refresh is enabled")
		return nil
	}
	v.Stop()
	v.app.Flash().Info("Auto-refresh is disabled")

	return nil
}

func (v *LiveView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if a, ok := v.actions[ui.AsKey(evt)]; ok {
		return a.Action(evt)
	}

	return evt
}

// StylesChanged notifies the skin changed.
func (v *LiveView) StylesChanged(s *config.Styles) {
	v.SetBackgroundColor(v.app.Styles.BgColor())
	v.text.SetTextColor(v.app.Styles.FgColor())
	v.SetBorderFocusColor(v.app.Styles.Frame().Border.FocusColor.Color())
	v.ResourceChanged(v.model.Peek(), nil)
}

// Actions returns menu actions.
func (v *LiveView) Actions() ui.KeyActions {
	return v.actions
}

// Name returns the component name.
func (v *LiveView) Name() string { return v.title }

// Start starts the view updater.
func (v *LiveView) Start() {
	if v.autoRefresh {
		var ctx context.Context
		ctx, v.cancel = context.WithCancel(v.defaultCtx())

		if err := v.model.Watch(ctx); err != nil {
			log.Error().Err(err).Msgf("LiveView watcher failed")
		}
		return
	}
	if err := v.model.Refresh(v.defaultCtx()); err != nil {
		log.Error().Err(err).Msgf("refresh failed")
	}
}

func (v *LiveView) defaultCtx() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, v.app.factory)
}

// Stop terminates the updater.
func (v *LiveView) Stop() {
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	v.app.Styles.RemoveListener(v)
}

// Hints returns menu hints.
func (v *LiveView) Hints() model.MenuHints {
	return v.actions.Hints()
}

// ExtraHints returns additional hints.
func (v *LiveView) ExtraHints() map[string]string {
	return nil
}

func (v *LiveView) toggleManagedCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.InCmdMode() {
		return evt
	}

	v.managedField = !v.managedField
	v.model.SetOptions(v.defaultCtx(), map[string]bool{model.ManagedFieldsOpts: v.managedField})

	return nil
}

func (v *LiveView) toggleFullScreenCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.InCmdMode() {
		return evt
	}

	v.fullScreen = !v.fullScreen
	v.SetFullScreen(v.fullScreen)
	v.Box.SetBorder(!v.fullScreen)
	if v.fullScreen {
		v.Box.SetBorderPadding(0, 0, 0, 0)
	} else {
		v.Box.SetBorderPadding(0, 0, 1, 1)
	}

	return nil
}

func (v *LiveView) nextCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.Empty() {
		return evt
	}

	v.currentRegion++
	if v.currentRegion >= v.maxRegions {
		v.currentRegion = 0
	}
	v.text.Highlight("search_" + strconv.Itoa(v.currentRegion))
	v.text.ScrollToHighlight()
	v.updateTitle()

	return nil
}

func (v *LiveView) prevCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.Empty() {
		return evt
	}

	v.currentRegion--
	if v.currentRegion < 0 {
		v.currentRegion = v.maxRegions - 1
	}
	v.text.Highlight("search_" + strconv.Itoa(v.currentRegion))
	v.text.ScrollToHighlight()
	v.updateTitle()

	return nil
}

func (v *LiveView) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.model.Filter(v.cmdBuff.GetText())
	v.cmdBuff.SetActive(false)
	v.updateTitle()

	return nil
}

func (v *LiveView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.InCmdMode() {
		return evt
	}
	v.app.ResetPrompt(v.cmdBuff)

	return nil
}

func (v *LiveView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.IsActive() {
		return nil
	}
	v.cmdBuff.Delete()

	return nil
}

func (v *LiveView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.InCmdMode() {
		v.cmdBuff.Reset()
		return v.app.PrevCmd(evt)
	}

	if v.cmdBuff.GetText() != "" {
		v.model.ClearFilter()
	}
	v.cmdBuff.SetActive(false)
	v.cmdBuff.Reset()
	v.updateTitle()

	return nil
}

func (v *LiveView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveYAML(v.app.Config.K9s.GetScreenDumpDir(), v.app.Config.K9s.CurrentContextDir(), v.title, v.text.GetText(true)); err != nil {
		v.app.Flash().Err(err)
	} else {
		v.app.Flash().Infof("Log %s saved successfully!", path)
	}

	return nil
}

func (v *LiveView) updateTitle() {
	if v.title == "" {
		return
	}
	fmat := fmt.Sprintf(liveViewTitleFmt, v.title, v.model.GetPath())

	buff := v.cmdBuff.GetText()
	if buff == "" {
		v.SetTitle(ui.SkinTitle(fmat, v.app.Styles.Frame()))
		return
	}

	if v.maxRegions > 0 {
		buff += fmt.Sprintf("[%d:%d]", v.currentRegion+1, v.maxRegions)
	}
	fmat += fmt.Sprintf(ui.SearchFmt, buff)
	v.SetTitle(ui.SkinTitle(fmat, v.app.Styles.Frame()))
}
