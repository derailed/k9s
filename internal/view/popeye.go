// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

// import (
// 	"context"
// 	"fmt"
// 	"strconv"
// 	"time"

// 	"github.com/derailed/k9s/internal"
// 	"github.com/derailed/k9s/internal/client"
// 	"github.com/derailed/k9s/internal/render"
// 	"github.com/derailed/k9s/internal/ui"
// 	"github.com/derailed/tcell/v2"
// )

// // Popeye represents a sanitizer view.
// type Popeye struct {
// 	ResourceViewer
// }

// // NewPopeye returns a new view.
// func NewPopeye(gvr client.GVR) ResourceViewer {
// 	p := Popeye{
// 		ResourceViewer: NewBrowser(gvr),
// 	}
// 	p.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
// 	p.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorMediumSpringGreen).Attributes(tcell.AttrNone))
// 	p.GetTable().SetSortCol("SCORE%", true)
// 	p.GetTable().SetDecorateFn(p.decorateRows)
// 	p.AddBindKeysFn(p.bindKeys)

// 	return &p
// }

// // Init initializes the view.
// func (p *Popeye) Init(ctx context.Context) error {
// 	if err := p.ResourceViewer.Init(ctx); err != nil {
// 		return err
// 	}
// 	p.GetTable().GetModel().SetRefreshRate(5 * time.Second)

// 	return nil
// }

// func (p *Popeye) decorateRows(data *model1.TableData) {
// 	var sum int
// 	for _, re := range data.RowEvents {
// 		n, err := strconv.Atoi(re.Row.Fields[1])
// 		if err != nil {
// 			continue
// 		}
// 		sum += n
// 	}
// 	score, letter := 0, render.NAValue
// 	if len(data.RowEvents) > 0 {
// 		score = sum / len(data.RowEvents)
// 		letter = grade(score)
// 	}
// 	p.GetTable().Extras = fmt.Sprintf("Score %d -- %s", score, letter)
// }

// func (p *Popeye) bindKeys(aa ui.KeyActions) {
// 	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
// 	aa.Add(ui.KeyActions{
// 		tcell.KeyEnter: ui.NewKeyAction("Goto", p.gotoCmd, true),
// 		ui.KeyShiftR:   ui.NewKeyAction("Sort Resource", p.GetTable().SortColCmd("RESOURCE", true), false),
// 		ui.KeyShiftS:   ui.NewKeyAction("Sort Score", p.GetTable().SortColCmd("SCORE%", true), false),
// 		ui.KeyShiftO:   ui.NewKeyAction("Sort OK", p.GetTable().SortColCmd("OK", true), false),
// 		ui.KeyShiftI:   ui.NewKeyAction("Sort Info", p.GetTable().SortColCmd("INFO", true), false),
// 		ui.KeyShiftW:   ui.NewKeyAction("Sort Warning", p.GetTable().SortColCmd("WARNING", true), false),
// 		ui.KeyShiftE:   ui.NewKeyAction("Sort Error", p.GetTable().SortColCmd("ERROR", true), false),
// 	})
// }

// func (p *Popeye) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	path := p.GetTable().GetSelectedItem()
// 	if path == "" {
// 		return evt
// 	}
// 	v := NewSanitizer(client.NewGVR("sanitizer"))
// 	v.SetContextFn(sanitizerCtx(path))
// 	if err := p.App().inject(v, false); err != nil {
// 		p.App().Flash().Err(err)
// 	}

// 	return nil
// }

// func sanitizerCtx(path string) ContextFunc {
// 	return func(ctx context.Context) context.Context {
// 		ctx = context.WithValue(ctx, internal.KeyPath, path)
// 		return ctx
// 	}
// }

// // Helpers...

// func grade(score int) string {
// 	switch {
// 	case score >= 90:
// 		return "A"
// 	case score >= 80:
// 		return "B"
// 	case score >= 70:
// 		return "C"
// 	case score >= 60:
// 		return "D"
// 	case score >= 50:
// 		return "E"
// 	default:
// 		return "F"
// 	}
// }
