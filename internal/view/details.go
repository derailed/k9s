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
	"github.com/gdamore/tcell"
	"github.com/sahilm/fuzzy"
)

const detailsTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

// Details represents a generic text viewer.
type Details struct {
	*tview.TextView

	actions                   ui.KeyActions
	app                       *App
	title, subject            string
	cmdBuff                   *model.CmdBuff
	model                     *model.Text
	currentRegion, maxRegions int
	searchable                bool
}

// NewDetails returns a details viewer.
func NewDetails(app *App, title, subject string, searchable bool) *Details {
	d := Details{
		TextView:   tview.NewTextView(),
		app:        app,
		title:      title,
		subject:    subject,
		actions:    make(ui.KeyActions),
		cmdBuff:    model.NewCmdBuff('/', model.Filter),
		model:      model.NewText(),
		searchable: searchable,
	}

	return &d
}

// Init initializes the viewer.
func (d *Details) Init(_ context.Context) error {
	if d.title != "" {
		d.SetBorder(true)
	}
	d.SetScrollable(true).SetWrap(true).SetRegions(true)
	d.SetDynamicColors(true)
	d.SetHighlightColor(tcell.ColorOrange)
	d.SetTitleColor(tcell.ColorAqua)
	d.SetInputCapture(d.keyboard)
	d.SetChangedFunc(func() {
		d.app.Draw()
	})
	d.updateTitle()

	d.app.Styles.AddListener(d)
	d.StylesChanged(d.app.Styles)

	d.cmdBuff.AddListener(d.app.Cmd())
	d.cmdBuff.AddListener(d)

	d.bindKeys()
	d.SetInputCapture(d.keyboard)
	d.model.AddListener(d)

	return nil
}

// TextChanged notifies the model changed.
func (d *Details) TextChanged(lines []string) {
	d.SetText(colorizeYAML(d.app.Styles.Views().Yaml, strings.Join(lines, "\n")))
	d.ScrollToBeginning()
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

	d.SetText(colorizeYAML(d.app.Styles.Views().Yaml, strings.Join(ll, "\n")))
	d.Highlight()
	if d.maxRegions > 0 {
		d.Highlight("search_0")
		d.ScrollToHighlight()
	}
}

// BufferChanged indicates the buffer was changed.
func (d *Details) BufferChanged(s string) {}

// BufferActive indicates the buff activity changed.
func (d *Details) BufferActive(state bool, k model.BufferKind) {
	d.app.BufferActive(state, k)
}

func (d *Details) bindKeys() {
	d.actions.Set(ui.KeyActions{
		tcell.KeyEnter:      ui.NewSharedKeyAction("Filter", d.filterCmd, false),
		tcell.KeyEscape:     ui.NewKeyAction("Back", d.resetCmd, false),
		tcell.KeyCtrlS:      ui.NewKeyAction("Save", d.saveCmd, false),
		ui.KeyC:             ui.NewKeyAction("Copy", d.cpCmd, true),
		ui.KeyN:             ui.NewKeyAction("Next Match", d.nextCmd, true),
		ui.KeyShiftN:        ui.NewKeyAction("Prev Match", d.prevCmd, true),
		ui.KeySlash:         ui.NewSharedKeyAction("Filter Mode", d.activateCmd, false),
		tcell.KeyCtrlU:      ui.NewSharedKeyAction("Clear Filter", d.clearCmd, false),
		tcell.KeyBackspace2: ui.NewSharedKeyAction("Erase", d.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewSharedKeyAction("Erase", d.eraseCmd, false),
		tcell.KeyDelete:     ui.NewSharedKeyAction("Erase", d.eraseCmd, false),
	})

	if !d.searchable {
		d.actions.Delete(ui.KeyN, ui.KeyShiftN)
	}
}

func (d *Details) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyUp || key == tcell.KeyDown {
		return evt
	}

	if key == tcell.KeyRune {
		if d.filterInput(evt.Rune()) {
			return nil
		}
		key = ui.AsKey(evt)
	}

	if a, ok := d.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

func (d *Details) filterInput(r rune) bool {
	if !d.cmdBuff.IsActive() {
		return false
	}
	d.cmdBuff.Add(r)
	d.updateTitle()

	return true
}

// StylesChanged notifies the skin changed.
func (d *Details) StylesChanged(s *config.Styles) {
	d.SetBackgroundColor(d.app.Styles.BgColor())
	d.SetTextColor(d.app.Styles.FgColor())
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

// Actions returns menu actions
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
	d.Highlight(fmt.Sprintf("search_%d", d.currentRegion))
	d.ScrollToHighlight()
	d.updateTitle()

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
	d.Highlight(fmt.Sprintf("search_%d", d.currentRegion))
	d.ScrollToHighlight()
	d.updateTitle()

	return nil
}

func (d *Details) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	d.model.Filter(d.cmdBuff.String())
	d.cmdBuff.SetActive(false)
	d.updateTitle()

	return nil
}

func (d *Details) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.app.InCmdMode() {
		return evt
	}
	d.cmdBuff.SetActive(true)

	return nil
}

func (d *Details) clearCmd(*tcell.EventKey) *tcell.EventKey {
	if !d.app.InCmdMode() {
		return nil
	}
	d.cmdBuff.Clear()

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

	if d.cmdBuff.String() != "" {
		d.model.ClearFilter()
	}
	d.cmdBuff.SetActive(false)
	d.cmdBuff.Reset()
	d.updateTitle()

	return nil
}

func (d *Details) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveYAML(d.app.Config.K9s.CurrentCluster, d.title, d.GetText(true)); err != nil {
		d.app.Flash().Err(err)
	} else {
		d.app.Flash().Infof("Log %s saved successfully!", path)
	}

	return nil
}

func (d *Details) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	d.app.Flash().Info("Content copied to clipboard...")
	if err := clipboard.WriteAll(d.GetText(true)); err != nil {
		d.app.Flash().Err(err)
	}

	return nil
}

func (d *Details) updateTitle() {
	if d.title == "" {
		return
	}
	fmat := fmt.Sprintf(detailsTitleFmt, d.title, d.subject)

	buff := d.cmdBuff.String()
	if buff == "" {
		d.SetTitle(ui.SkinTitle(fmat, d.app.Styles.Frame()))
		return
	}

	search := d.cmdBuff.String()
	if d.maxRegions != 0 {
		search += fmt.Sprintf("[%d:%d]", d.currentRegion+1, d.maxRegions)
	}
	fmat += fmt.Sprintf(ui.SearchFmt, search)
	d.SetTitle(ui.SkinTitle(fmat, d.app.Styles.Frame()))
}
