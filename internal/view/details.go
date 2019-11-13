package view

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const detailsTitleFmt = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "

// Details presents a generic text viewer.
type Details struct {
	*tview.TextView

	app           *App
	actions       ui.KeyActions
	cmdBuff       *ui.CmdBuff
	title         string
	category      string
	backFn        ui.ActionHandler
	numSelections int
}

// NewDetails returns a details viewer.
func NewDetails(app *App, backFn ui.ActionHandler) *Details {
	return &Details{
		TextView: tview.NewTextView(),
		app:      app,
		backFn:   backFn,
	}
}

func (d *Details) Init(ctx context.Context) {
	d.app = ctx.Value(ui.KeyApp).(*App)

	d.SetScrollable(true)
	d.SetWrap(true)
	d.SetDynamicColors(true)
	d.SetRegions(true)
	d.SetBorder(true)
	d.SetBorderFocusColor(config.AsColor(d.app.Styles.Frame().Border.FocusColor))
	d.SetHighlightColor(tcell.ColorOrange)
	d.SetTitleColor(tcell.ColorAqua)
	d.SetInputCapture(d.keyboard)
	d.bindKeys()

	d.cmdBuff = ui.NewCmdBuff('/', ui.FilterBuff)
	d.cmdBuff.AddListener(d.app.Cmd())
	d.cmdBuff.Reset()

	d.SetChangedFunc(func() {
		d.app.Draw()
	})
}

func (d *Details) Name() string { return "details" }
func (d *Details) Start()       {}
func (d *Details) Stop()        {}

func (d *Details) bindKeys() {
	d.actions = ui.KeyActions{
		tcell.KeyBackspace2: ui.NewKeyAction("Erase", d.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewKeyAction("Erase", d.eraseCmd, false),
		tcell.KeyDelete:     ui.NewKeyAction("Erase", d.eraseCmd, false),
		tcell.KeyEscape:     ui.NewKeyAction("Back", d.backCmd, true),
		tcell.KeyTab:        ui.NewKeyAction("Next Match", d.nextCmd, false),
		tcell.KeyBacktab:    ui.NewKeyAction("Previous Match", d.prevCmd, false),
		tcell.KeyCtrlS:      ui.NewKeyAction("Save", d.saveCmd, true),
		ui.KeyC:             ui.NewKeyAction("Copy", d.cpCmd, false),
	}
}

func (d *Details) setCategory(n string) {
	d.category = n
}

func (d *Details) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if d.cmdBuff.IsActive() {
			d.cmdBuff.Add(evt.Rune())
			d.refreshTitle()
			return nil
		}
		key = tcell.Key(evt.Rune())
	}

	if a, ok := d.actions[key]; ok {
		log.Debug().Msgf(">> DetailsView handled %s", tcell.KeyNames[key])
		return a.Action(evt)
	}
	return evt
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

func (d *Details) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !d.cmdBuff.Empty() {
		d.cmdBuff.Reset()
		d.search(evt)
		return nil
	}
	d.cmdBuff.Reset()
	if d.backFn != nil {
		return d.backFn(evt)
	}
	return evt
}

func (d *Details) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !d.cmdBuff.IsActive() {
		return evt
	}
	d.cmdBuff.Delete()
	return nil
}

func (d *Details) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !d.app.InCmdMode() {
		d.cmdBuff.SetActive(true)
		d.cmdBuff.Clear()
		return nil
	}
	return evt
}

func (d *Details) searchCmd(evt *tcell.EventKey) *tcell.EventKey {
	if d.cmdBuff.IsActive() && !d.cmdBuff.Empty() {
		d.app.Flash().Infof("Searching for %s...", d.cmdBuff)
		d.search(evt)
		highlights := d.GetHighlights()
		if len(highlights) > 0 {
			d.Highlight()
		} else {
			d.Highlight("0").ScrollToHighlight()
		}
	}
	d.cmdBuff.SetActive(false)
	return evt
}

func (d *Details) search(evt *tcell.EventKey) {
	d.numSelections = 0
	log.Debug().Msgf("Searching... %s - %d", d.cmdBuff, d.numSelections)
	d.Highlight("")
	d.SetText(d.decorateLines(d.GetText(false), d.cmdBuff.String()))

	if d.cmdBuff.Empty() {
		d.app.Flash().Info("Clearing out search query...")
		d.refreshTitle()
		return
	}
	if d.numSelections == 0 {
		d.app.Flash().Warn("No matches found!")
		return
	}
	d.app.Flash().Infof("Found <%d> matches! <tab>/<TAB> for next/previous", d.numSelections)
}

func (d *Details) nextCmd(evt *tcell.EventKey) *tcell.EventKey {
	highlights := d.GetHighlights()
	if len(highlights) == 0 || d.numSelections == 0 {
		return evt
	}
	index, _ := strconv.Atoi(highlights[0])
	index = (index + 1) % d.numSelections
	if index+1 == d.numSelections {
		d.app.Flash().Info("Search hit BOTTOM, continuing at TOP")
	}
	d.Highlight(strconv.Itoa(index)).ScrollToHighlight()
	return nil
}

func (d *Details) prevCmd(evt *tcell.EventKey) *tcell.EventKey {
	highlights := d.GetHighlights()
	if len(highlights) == 0 || d.numSelections == 0 {
		return evt
	}
	index, _ := strconv.Atoi(highlights[0])
	index = (index - 1 + d.numSelections) % d.numSelections
	if index == 0 {
		d.app.Flash().Info("Search hit TOP, continuing at BOTTOM")
	}
	d.Highlight(strconv.Itoa(index)).ScrollToHighlight()
	return nil
}

// SetActions to handle keyboard inputs
func (d *Details) setActions(aa ui.KeyActions) {
	for k, a := range aa {
		d.actions[k] = a
	}
}

// Hints fetch mmemonic and hints
func (d *Details) Hints() model.MenuHints {
	if d.actions != nil {
		return d.actions.Hints()
	}
	return nil
}

func (d *Details) refreshTitle() {
	d.setTitle(d.title)
}

func (d *Details) setTitle(t string) {
	d.title = t

	title := skinTitle(fmt.Sprintf(detailsTitleFmt, d.category, t), d.app.Styles.Frame())
	if !d.cmdBuff.Empty() {
		title += skinTitle(fmt.Sprintf(ui.SearchFmt, d.cmdBuff.String()), d.app.Styles.Frame())
	}
	d.SetTitle(title)
}

var (
	regionRX = regexp.MustCompile(`\["([a-zA-Z0-9_,;: \-\.]*)"\]`)
	escapeRX = regexp.MustCompile(`\[([a-zA-Z0-9_,;: \-\."#]+)\[(\[*)\]`)
)

func (d *Details) decorateLines(buff, q string) string {
	rx := regexp.MustCompile(`(?i)` + q)
	lines := strings.Split(buff, "\n")
	for i, l := range lines {
		l = regionRX.ReplaceAllString(l, "")
		l = escapeRX.ReplaceAllString(l, "")
		if m := rx.FindString(l); len(m) > 0 {
			lines[i] = rx.ReplaceAllString(l, fmt.Sprintf(`["%d"]%s[""]`, d.numSelections, m))
			d.numSelections++
			continue
		}
		lines[i] = l
	}
	return strings.Join(lines, "\n")
}
