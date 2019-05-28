package views

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const detailsTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

// detailsView displays text output.
type detailsView struct {
	*tview.TextView

	app           *appView
	actions       keyActions
	title         string
	category      string
	cmdBuff       *cmdBuff
	backFn        actionHandler
	numSelections int
}

func newDetailsView(app *appView, backFn actionHandler) *detailsView {
	v := detailsView{TextView: tview.NewTextView(), app: app, actions: make(keyActions)}
	{
		v.backFn = backFn
		v.SetScrollable(true)
		v.SetWrap(true)
		v.SetDynamicColors(true)
		v.SetRegions(true)
		v.SetBorder(true)
		v.SetBorderFocusColor(config.AsColor(v.app.styles.Style.Border.FocusColor))
		v.SetHighlightColor(tcell.ColorOrange)
		v.SetTitleColor(tcell.ColorAqua)
		v.SetInputCapture(v.keyboard)
		v.cmdBuff = newCmdBuff('/')
		{
			v.cmdBuff.addListener(app.cmdView)
			v.cmdBuff.reset()
		}
		v.SetChangedFunc(func() {
			app.Draw()
		})
	}

	v.actions[tcell.KeyBackspace2] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[tcell.KeyBackspace] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[tcell.KeyDelete] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[tcell.KeyEscape] = newKeyAction("Back", v.backCmd, true)
	v.actions[tcell.KeyTab] = newKeyAction("Next Match", v.nextCmd, false)
	v.actions[tcell.KeyBacktab] = newKeyAction("Previous Match", v.prevCmd, false)

	return &v
}

func (v *detailsView) setCategory(n string) {
	v.category = n
}

func (v *detailsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if v.cmdBuff.isActive() {
			v.cmdBuff.add(evt.Rune())
			v.refreshTitle()
			return nil
		}
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> DetailsView handled %s", tcell.KeyNames[key])
		return a.action(evt)
	}
	return evt
}

func (v *detailsView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		v.search(evt)
		return nil
	}
	v.cmdBuff.reset()
	if v.backFn != nil {
		return v.backFn(evt)
	}
	return evt
}

func (v *detailsView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.isActive() {
		return evt
	}
	v.cmdBuff.del()
	return nil
}

func (v *detailsView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.app.cmdView.inCmdMode() {
		v.cmdBuff.setActive(true)
		v.cmdBuff.clear()
		return nil
	}
	return evt
}

func (v *detailsView) searchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.isActive() && !v.cmdBuff.empty() {
		v.app.flash().infof("Searching for %s...", v.cmdBuff)
		v.search(evt)
		highlights := v.GetHighlights()
		if len(highlights) > 0 {
			v.Highlight()
		} else {
			v.Highlight("0").ScrollToHighlight()
		}
	}
	v.cmdBuff.setActive(false)
	return evt
}

func (v *detailsView) search(evt *tcell.EventKey) {
	v.numSelections = 0
	v.Highlight()
	log.Debug().Msgf("Searching... %s - %d", v.cmdBuff, v.numSelections)
	v.Highlight("")
	v.SetText(v.decorateLines(v.GetText(false), v.cmdBuff.String()))

	if v.cmdBuff.empty() {
		v.app.flash().info("Clearing out search query...")
		v.refreshTitle()
		return
	}
	if v.numSelections == 0 {
		v.app.flash().warn("No matches found!")
		return
	}
	v.app.flash().infof("Found <%d> matches! <tab>/<TAB> for next/previous", v.numSelections)
}

func (v *detailsView) nextCmd(evt *tcell.EventKey) *tcell.EventKey {
	highlights := v.GetHighlights()
	if len(highlights) == 0 || v.numSelections == 0 {
		return evt
	}
	index, _ := strconv.Atoi(highlights[0])
	index = (index + 1) % v.numSelections
	if index+1 == v.numSelections {
		v.app.flash().info("Search hit BOTTOM, continuing at TOP")
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
		v.app.flash().info("Search hit TOP, continuing at BOTTOM")
	}
	v.Highlight(strconv.Itoa(index)).ScrollToHighlight()
	return nil
}

// SetActions to handle keyboard inputs
func (v *detailsView) setActions(aa keyActions) {
	for k, a := range aa {
		v.actions[k] = a
	}
}

// Hints fetch mmemonic and hints
func (v *detailsView) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}
	return nil
}

func (v *detailsView) refreshTitle() {
	v.setTitle(v.title)
}

func (v *detailsView) setTitle(t string) {
	v.title = t

	title := skinTitle(fmt.Sprintf(detailsTitleFmt, v.category, t), v.app.styles.Style)
	if !v.cmdBuff.empty() {
		title += skinTitle(fmt.Sprintf(searchFmt, v.cmdBuff.String()), v.app.styles.Style)
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
		} else {
			lines[i] = l
		}
	}
	return strings.Join(lines, "\n")
}
