package views

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const detailsTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

type (
	textView struct {
		*tview.TextView

		app     *appView
		actions ui.KeyActions
		cmdBuff *ui.CmdBuff
		title   string
	}

	detailsView struct {
		*textView

		category      string
		backFn        ui.ActionHandler
		numSelections int
	}
)

func newTextView(app *appView) *textView {
	return &textView{
		TextView: tview.NewTextView(),
		app:      app,
		actions:  make(ui.KeyActions),
	}
}

func newDetailsView(app *appView, backFn ui.ActionHandler) *detailsView {
	v := detailsView{textView: newTextView(app)}
	v.backFn = backFn
	v.SetScrollable(true)
	v.SetWrap(true)
	v.SetDynamicColors(true)
	v.SetRegions(true)
	v.SetBorder(true)
	v.SetBorderFocusColor(config.AsColor(v.app.Styles.Frame().Border.FocusColor))
	v.SetHighlightColor(tcell.ColorOrange)
	v.SetTitleColor(tcell.ColorAqua)
	v.SetInputCapture(v.keyboard)

	v.cmdBuff = ui.NewCmdBuff('/', ui.FilterBuff)
	v.cmdBuff.AddListener(app.Cmd())
	v.cmdBuff.Reset()

	v.SetChangedFunc(func() {
		app.Draw()
	})

	v.bindKeys()

	return &v
}

func (v *detailsView) bindKeys() {
	v.actions = ui.KeyActions{
		tcell.KeyBackspace2: ui.NewKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyDelete:     ui.NewKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyEscape:     ui.NewKeyAction("Back", v.backCmd, true),
		tcell.KeyTab:        ui.NewKeyAction("Next Match", v.nextCmd, false),
		tcell.KeyBacktab:    ui.NewKeyAction("Previous Match", v.prevCmd, false),
		tcell.KeyCtrlS:      ui.NewKeyAction("Save", v.saveCmd, true),
		ui.KeyC:             ui.NewKeyAction("Copy", v.cpCmd, false),
	}
}

func (v *detailsView) setCategory(n string) {
	v.category = n
}

func (v *detailsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if v.cmdBuff.IsActive() {
			v.cmdBuff.Add(evt.Rune())
			v.refreshTitle()
			return nil
		}
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> DetailsView handled %s", tcell.KeyNames[key])
		return a.Action(evt)
	}
	return evt
}

func (v *detailsView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveYAML(v.app.Config.K9s.CurrentCluster, v.title, v.GetText(true)); err != nil {
		v.app.Flash().Err(err)
	} else {
		v.app.Flash().Infof("Log %s saved successfully!", path)
	}
	return nil
}

func (v *detailsView) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.Flash().Info("Content copied to clipboard...")
	if err := clipboard.WriteAll(v.GetText(true)); err != nil {
		v.app.Flash().Err(err)
	}
	return nil
}

func (v *detailsView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.Empty() {
		v.cmdBuff.Reset()
		v.search(evt)
		return nil
	}
	v.cmdBuff.Reset()
	if v.backFn != nil {
		return v.backFn(evt)
	}
	return evt
}

func (v *detailsView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.IsActive() {
		return evt
	}
	v.cmdBuff.Delete()
	return nil
}

func (v *detailsView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.app.InCmdMode() {
		v.cmdBuff.SetActive(true)
		v.cmdBuff.Clear()
		return nil
	}
	return evt
}

func (v *detailsView) searchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.IsActive() && !v.cmdBuff.Empty() {
		v.app.Flash().Infof("Searching for %s...", v.cmdBuff)
		v.search(evt)
		highlights := v.GetHighlights()
		if len(highlights) > 0 {
			v.Highlight()
		} else {
			v.Highlight("0").ScrollToHighlight()
		}
	}
	v.cmdBuff.SetActive(false)
	return evt
}

func (v *detailsView) search(evt *tcell.EventKey) {
	v.numSelections = 0
	log.Debug().Msgf("Searching... %s - %d", v.cmdBuff, v.numSelections)
	v.Highlight("")
	v.SetText(v.decorateLines(v.GetText(false), v.cmdBuff.String()))

	if v.cmdBuff.Empty() {
		v.app.Flash().Info("Clearing out search query...")
		v.refreshTitle()
		return
	}
	if v.numSelections == 0 {
		v.app.Flash().Warn("No matches found!")
		return
	}
	v.app.Flash().Infof("Found <%d> matches! <tab>/<TAB> for next/previous", v.numSelections)
}

func (v *detailsView) nextCmd(evt *tcell.EventKey) *tcell.EventKey {
	highlights := v.GetHighlights()
	if len(highlights) == 0 || v.numSelections == 0 {
		return evt
	}
	index, _ := strconv.Atoi(highlights[0])
	index = (index + 1) % v.numSelections
	if index+1 == v.numSelections {
		v.app.Flash().Info("Search hit BOTTOM, continuing at TOP")
	}
	v.Highlight(strconv.Itoa(index)).ScrollToHighlight()
	return nil
}

func (v *detailsView) prevCmd(evt *tcell.EventKey) *tcell.EventKey {
	highlights := v.GetHighlights()
	if len(highlights) == 0 || v.numSelections == 0 {
		return evt
	}
	index, _ := strconv.Atoi(highlights[0])
	index = (index - 1 + v.numSelections) % v.numSelections
	if index == 0 {
		v.app.Flash().Info("Search hit TOP, continuing at BOTTOM")
	}
	v.Highlight(strconv.Itoa(index)).ScrollToHighlight()
	return nil
}

// SetActions to handle keyboard inputs
func (v *detailsView) setActions(aa ui.KeyActions) {
	for k, a := range aa {
		v.actions[k] = a
	}
}

// Hints fetch mmemonic and hints
func (v *detailsView) hints() ui.Hints {
	if v.actions != nil {
		return v.actions.Hints()
	}
	return nil
}

func (v *detailsView) refreshTitle() {
	v.setTitle(v.title)
}

func (v *detailsView) setTitle(t string) {
	v.title = t

	title := skinTitle(fmt.Sprintf(detailsTitleFmt, v.category, t), v.app.Styles.Frame())
	if !v.cmdBuff.Empty() {
		title += skinTitle(fmt.Sprintf(searchFmt, v.cmdBuff.String()), v.app.Styles.Frame())
	}
	v.SetTitle(title)
}

var (
	regionRX = regexp.MustCompile(`\["([a-zA-Z0-9_,;: \-\.]*)"\]`)
	escapeRX = regexp.MustCompile(`\[([a-zA-Z0-9_,;: \-\."#]+)\[(\[*)\]`)
)

func (v *detailsView) decorateLines(buff, q string) string {
	rx := regexp.MustCompile(`(?i)` + q)
	lines := strings.Split(buff, "\n")
	for i, l := range lines {
		l = regionRX.ReplaceAllString(l, "")
		l = escapeRX.ReplaceAllString(l, "")
		if m := rx.FindString(l); len(m) > 0 {
			lines[i] = rx.ReplaceAllString(l, fmt.Sprintf(`["%d"]%s[""]`, v.numSelections, m))
			v.numSelections++
			continue
		}
		lines[i] = l
	}
	return strings.Join(lines, "\n")
}
