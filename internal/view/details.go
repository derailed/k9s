package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/sahilm/fuzzy"
)

const detailsTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

// Details represents a generic text viewer.
type Details struct {
	*tview.Flex

	text                      *tview.TextView
	actions                   ui.KeyActions
	app                       *App
	title, subject            string
	cmdBuff                   *model.FishBuff
	model                     *model.Text
	currentRegion, maxRegions int
	searchable                bool
	fullScreen                bool
}

// NewDetails returns a details viewer.
func NewDetails(app *App, title, subject string, searchable bool) *Details {
	d := Details{
		Flex:       tview.NewFlex(),
		text:       tview.NewTextView(),
		app:        app,
		title:      title,
		subject:    subject,
		actions:    make(ui.KeyActions),
		cmdBuff:    model.NewFishBuff('/', model.FilterBuffer),
		model:      model.NewText(),
		searchable: searchable,
	}
	d.AddItem(d.text, 0, 1, true)

	return &d
}

// Init initializes the viewer.
func (d *Details) Init(_ context.Context) error {
	if d.title != "" {
		d.SetBorder(true)
	}
	d.text.SetScrollable(true).SetWrap(true).SetRegions(true)
	d.text.SetDynamicColors(true)
	d.text.SetHighlightColor(tcell.ColorOrange)
	d.SetTitleColor(tcell.ColorAqua)
	d.SetInputCapture(d.keyboard)
	d.SetBorderPadding(0, 0, 1, 1)
	d.updateTitle()

	d.app.Styles.AddListener(d)
	d.StylesChanged(d.app.Styles)

	d.app.Prompt().SetModel(d.cmdBuff)
	d.cmdBuff.AddListener(d)

	d.bindKeys()
	d.SetInputCapture(d.keyboard)
	d.model.AddListener(d)

	return nil
}

// InCmdMode checks if prompt is active.
func (d *Details) InCmdMode() bool {
	return d.cmdBuff.InCmdMode()
}

// TextChanged notifies the model changed.
func (d *Details) TextChanged(lines []string) {
	d.text.SetText(colorizeYAML(d.app.Styles.Views().Yaml, strings.Join(lines, "\n")))
	d.text.ScrollToBeginning()
}

// TextFiltered notifies when the filter changed.
func (d *Details) TextFiltered(lines []string, matches fuzzy.Matches) {
	d.currentRegion, d.maxRegions = 0, 0

	ll := make([]string, len(lines))
	copy(ll, lines)
	for _, m := range matches {
		loc, line := m.MatchedIndexes, ll[m.Index]
		ll[m.Index] = line[:loc[0]] + fmt.Sprintf(`<<<"search_%d">>>`, d.maxRegions) + line[loc[0]:loc[1]] + `<<<"">>>` + line[loc[1]:]
		d.maxRegions++
	}

	d.text.SetText(colorizeYAML(d.app.Styles.Views().Yaml, strings.Join(ll, "\n")))
	d.text.Highlight()
	if d.maxRegions > 0 {
		d.text.Highlight("search_0")
		d.text.ScrollToHighlight()
	}
}

// BufferChanged indicates the buffer was changed.
func (d *Details) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (d *Details) BufferCompleted(text, _ string) {
	d.model.Filter(text)
	d.updateTitle()
}

// BufferActive indicates the buff activity changed.
func (d *Details) BufferActive(state bool, k model.BufferKind) {
	d.app.BufferActive(state, k)
}

func (d *Details) bindKeys() {
	d.actions.Set(ui.KeyActions{
		tcell.KeyEnter:  ui.NewSharedKeyAction("Filter", d.filterCmd, false),
		tcell.KeyEscape: ui.NewKeyAction("Back", d.resetCmd, false),
		tcell.KeyCtrlS:  ui.NewKeyAction("Save", d.saveCmd, false),
		ui.KeyC:         ui.NewKeyAction("Copy", d.cpCmd, true),
		ui.KeyF:         ui.NewKeyAction("Toggle FullScreen", d.toggleFullScreenCmd, true),
		ui.KeyN:         ui.NewKeyAction("Next Match", d.nextCmd, true),
		ui.KeyShiftN:    ui.NewKeyAction("Prev Match", d.prevCmd, true),
		ui.KeySlash:     ui.NewSharedKeyAction("Filter Mode", d.activateCmd, false),
		tcell.KeyDelete: ui.NewSharedKeyAction("Erase", d.eraseCmd, false),
	})

	if !d.searchable {
		d.actions.Delete(ui.KeyN, ui.KeyShiftN)
	}
}

func (d *Details) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if a, ok := d.actions[ui.AsKey(evt)]; ok {
		return a.Action(evt)
	}

	return evt
}

// StylesChanged notifies the skin changed.
func (d *Details) StylesChanged(s *config.Styles) {
	d.SetBackgroundColor(d.app.Styles.BgColor())
	d.text.SetTextColor(d.app.Styles.FgColor())
	d.SetBorderFocusColor(d.app.Styles.Frame().Border.FocusColor.Color())
	d.TextChanged(d.model.Peek())
}

// Update updates the view content.
func (d *Details) Update(buff string) *Details {
	d.model.SetText(buff)
	return d
}

// SetSubject updates the subject.
func (d *Details) SetSubject(s string) {
	d.subject = s
}

// Actions returns menu actions.
func (d *Details) Actions() ui.KeyActions {
	return d.actions
}

// Name returns the component name.
func (d *Details) Name() string { return d.title }

// Start starts the view updater.
func (d *Details) Start() {}

// Stop terminates the updater.
func (d *Details) Stop() {
	d.app.Styles.RemoveListener(d)
}

// Hints returns menu hints.
func (d *Details) Hints() model.MenuHints {
	return d.actions.Hints()
}

// ExtraHints returns additional hints.
func (d *Details) ExtraHints() map[string]string {
	return nil
}

func (d *Details) nextCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.cmdBuff.Empty() {
		return evt
	}

	d.currentRegion++
	if d.currentRegion >= d.maxRegions {
		d.currentRegion = 0
	}
	d.text.Highlight(fmt.Sprintf("search_%d", d.currentRegion))
	d.text.ScrollToHighlight()
	d.updateTitle()

	return nil
}

func (d *Details) toggleFullScreenCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.app.InCmdMode() {
		return evt
	}

	d.fullScreen = !d.fullScreen
	d.SetFullScreen(d.fullScreen)
	d.Box.SetBorder(!d.fullScreen)

	return nil
}

func (d *Details) prevCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.cmdBuff.Empty() {
		return evt
	}

	d.currentRegion--
	if d.currentRegion < 0 {
		d.currentRegion = d.maxRegions - 1
	}
	d.text.Highlight(fmt.Sprintf("search_%d", d.currentRegion))
	d.text.ScrollToHighlight()
	d.updateTitle()

	return nil
}

func (d *Details) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	d.model.Filter(d.cmdBuff.GetText())
	d.cmdBuff.SetActive(false)
	d.updateTitle()

	return nil
}

func (d *Details) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.app.InCmdMode() {
		return evt
	}
	d.app.ResetPrompt(d.cmdBuff)

	return nil
}

func (d *Details) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !d.cmdBuff.IsActive() {
		return nil
	}
	d.cmdBuff.Delete()

	return nil
}

func (d *Details) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !d.cmdBuff.InCmdMode() {
		d.cmdBuff.Reset()
		return d.app.PrevCmd(evt)
	}

	if d.cmdBuff.GetText() != "" {
		d.model.ClearFilter()
	}
	d.cmdBuff.SetActive(false)
	d.cmdBuff.Reset()
	d.updateTitle()

	return nil
}

func (d *Details) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveYAML(d.app.Config.K9s.CurrentCluster, d.title, d.text.GetText(true)); err != nil {
		d.app.Flash().Err(err)
	} else {
		d.app.Flash().Infof("Log %s saved successfully!", path)
	}

	return nil
}

func (d *Details) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	d.app.Flash().Info("Content copied to clipboard...")
	if err := clipboard.WriteAll(d.text.GetText(true)); err != nil {
		d.app.Flash().Err(err)
	}

	return nil
}

func (d *Details) updateTitle() {
	if d.title == "" {
		return
	}
	fmat := fmt.Sprintf(detailsTitleFmt, d.title, d.subject)

	buff := d.cmdBuff.GetText()
	if buff == "" {
		d.SetTitle(ui.SkinTitle(fmat, d.app.Styles.Frame()))
		return
	}

	if d.maxRegions != 0 {
		buff += fmt.Sprintf("[%d:%d]", d.currentRegion+1, d.maxRegions)
	}
	fmat += fmt.Sprintf(ui.SearchFmt, buff)
	d.SetTitle(ui.SkinTitle(fmat, d.app.Styles.Frame()))
}
