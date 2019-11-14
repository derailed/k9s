package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
)

// Crumbs represents user breadcrumbs.
type Crumbs struct {
	*tview.TextView

	styles *config.Styles
	stack  *model.Stack
}

// NewCrumbs returns a new breadcrumb view.
func NewCrumbs(styles *config.Styles) *Crumbs {
	v := Crumbs{
		stack:    model.NewStack(),
		styles:   styles,
		TextView: tview.NewTextView(),
	}
	v.SetBackgroundColor(styles.BgColor())
	v.SetTextAlign(tview.AlignLeft)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetDynamicColors(true)

	return &v
}

// StackPushed indicates a new item was added.
func (v *Crumbs) StackPushed(c model.Component) {
	v.stack.Push(c)
	v.refresh(v.stack.Flatten())
}

// StackPopped indicates an item was deleted
func (v *Crumbs) StackPopped(_, _ model.Component) {
	v.stack.Pop()
	v.refresh(v.stack.Flatten())
}

// StackTop indicates the top of the stack
func (v *Crumbs) StackTop(top model.Component) {}

// Refresh updates view with new crumbs.
func (v *Crumbs) refresh(crumbs []string) {
	v.Clear()
	last, bgColor := len(crumbs)-1, v.styles.Frame().Crumb.BgColor
	for i, c := range crumbs {
		if i == last {
			bgColor = v.styles.Frame().Crumb.ActiveColor
		}
		fmt.Fprintf(v, "[%s:%s:b] <%s> [-:%s:-] ",
			v.styles.Frame().Crumb.FgColor,
			bgColor, c,
			v.styles.Body().BgColor)
	}
}
