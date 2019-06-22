package views

import (
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/util/duration"
)

type resTable struct {
	*tview.Table

	app       *appView
	baseTitle string
	currentNS string
	data      resource.TableData
	actions   keyActions
}

func newResTable(app *appView, title string) *resTable {
	v := resTable{
		Table:     tview.NewTable(),
		app:       app,
		actions:   make(keyActions),
		baseTitle: title,
	}

	v.SetFixed(1, 0)
	v.SetBorder(true)
	v.SetBackgroundColor(config.AsColor(app.styles.Table().BgColor))
	v.SetBorderColor(config.AsColor(app.styles.Table().FgColor))
	v.SetBorderFocusColor(config.AsColor(app.styles.Frame().Border.FocusColor))
	v.SetBorderAttributes(tcell.AttrBold)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetSelectable(true, false)
	v.SetSelectedStyle(
		tcell.ColorBlack,
		config.AsColor(app.styles.Table().CursorColor),
		tcell.AttrBold,
	)

	return &v
}

func (v *resTable) formatCell(numerical bool, header, field string, padding int) (string, int) {
	if header == "AGE" {
		dur, err := time.ParseDuration(field)
		if err == nil {
			field = duration.HumanDuration(dur)
		}
	}

	if numerical || cpuRX.MatchString(header) || memRX.MatchString(header) {
		return field, tview.AlignRight
	}

	align := tview.AlignLeft
	if isASCII(field) {
		return pad(field, padding), align
	}

	return field, align
}

func (v *resTable) clearSelection() {
	v.Select(0, 0)
	v.ScrollToBeginning()
}

func (v *resTable) selectFirstRow() {
	if v.GetRowCount() > 0 {
		v.Select(1, 0)
	}
}

func (v *resTable) setDeleted() {
	r, _ := v.GetSelection()
	cols := v.GetColumnCount()
	for x := 0; x < cols; x++ {
		v.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

// SetActions sets up keyboard action listener.
func (v *resTable) setActions(aa keyActions) {
	for k, a := range aa {
		v.actions[k] = a
	}
}

// Hints options
func (v *resTable) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}

	return nil
}

func (v *resTable) nameColIndex() int {
	col := 0
	if v.currentNS == resource.AllNamespaces {
		col++
	}
	return col
}

func (v *resTable) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveTable(v.app.config.K9s.CurrentCluster, v.baseTitle, v.data); err != nil {
		v.app.flash().err(err)
	} else {
		v.app.flash().infof("File %s saved successfully!", path)
	}

	return nil
}
