package view

import (
	"context"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	aliasTitle    = "Aliases"
	aliasTitleFmt = " [mediumseagreen::b]%s([fuchsia::b]%d[fuchsia::-][mediumseagreen::-]) "
)

// Alias represents a command alias view.
type Alias struct {
	ResourceViewer
}

// NewAlias returns a new alias view.
func NewAlias(gvr dao.GVR) ResourceViewer {
	a := Alias{
		ResourceViewer: NewBrowser(gvr),
	}
	a.GetTable().SetColorerFn(render.Alias{}.ColorerFunc())
	a.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	a.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	a.SetBindKeysFn(a.bindKeys)
	a.SetContextFn(a.aliasContext)
	// a.GetTable().SetEnterFn(a.gotoCmd)

	return &a
}

func (a *Alias) aliasContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyAliases, aliases.Alias)
}

func (a *Alias) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, true),
		// BOZO!!
		// tcell.KeyEscape: ui.NewKeyAction("Reset", a.resetCmd, false),
		// ui.KeySlash:     ui.NewKeyAction("Filter", a.GetTable().activateCmd, false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Resource", a.GetTable().SortColCmd(0, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Command", a.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort ApiGroup", a.GetTable().SortColCmd(2, true), false),
	})
}

func (a *Alias) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("GOTO CMD")
	r, _ := a.GetTable().GetSelection()
	if r != 0 {
		s := ui.TrimCell(a.GetTable().SelectTable, r, 1)
		tokens := strings.Split(s, ",")
		a.App().gotoResource(tokens[0])
		return nil
	}

	if a.GetTable().SearchBuff().IsActive() {
		return a.GetTable().activateCmd(evt)
	}
	return evt
}

func (a *Alias) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !a.GetTable().SearchBuff().Empty() {
		a.GetTable().SearchBuff().Reset()
		return nil
	}

	return a.App().PrevCmd(evt)
}

func (a *Alias) gotoCmd1(app *App, ns, res, path string) {
	log.Debug().Msgf("GOTO %q -- %q -- %q", ns, res, path)
	app.gotoResource(dao.GVR(path).ToR())
	// r, _ := a.GetTable().GetSelection()
	// if r != 0 {
	// 	s := ui.TrimCell(a.GetTable().SelectTable, r, 1)
	// 	tokens := strings.Split(s, ",")
	// 	a.App().Content.Pop()
	// if err := a.App().gotoResource(tokens[0]); err != nil {
	// 	a.App().Flash().Err(err)
	// }
	// return nil
	// }

	// if a.GetTable().SearchBuff().IsActive() {
	// 	return a.GetTable().activateCmd(evt)
	// }

	// return evt
}

// BOZO!!
// func (a *Alias) hydrate() render.TableData {
// 	var re render.Alias

// 	data := render.TableData{
// 		Header:    re.Header(render.AllNamespaces),
// 		RowEvents: make(render.RowEvents, 0, len(aliases.Alias)),
// 		Namespace: resource.NotNamespaced,
// 	}

// 	aa := make(config.ShortNames, len(aliases.Alias))
// 	for alias, gvr := range aliases.Alias {
// 		if _, ok := aa[gvr]; ok {
// 			aa[gvr] = append(aa[gvr], alias)
// 		} else {
// 			aa[gvr] = []string{alias}
// 		}
// 	}

// 	for gvr, aliases := range aa {
// 		var row render.Row
// 		if err := re.Render(aliases, gvr, &row); err != nil {
// 			log.Error().Err(err).Msgf("Alias render failed")
// 			continue
// 		}
// 		data.RowEvents = append(data.RowEvents, render.RowEvent{
// 			Kind: render.EventAdd,
// 			Row:  row,
// 		})
// 	}

// 	return data
// }

// func (a *Alias) resetTitle() {
// 	a.SetTitle(fmt.Sprintf(aliasTitleFmt, aliasTitle, a.GetRowCount()-1))
// }
