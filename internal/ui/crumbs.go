// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"strings"

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
	c := Crumbs{
		stack:    model.NewStack(),
		styles:   styles,
		TextView: tview.NewTextView(),
	}
	c.SetBackgroundColor(styles.BgColor())
	c.SetTextAlign(tview.AlignLeft)
	c.SetBorderPadding(0, 0, 1, 1)
	c.SetDynamicColors(true)
	styles.AddListener(&c)

	return &c
}

// StylesChanged notifies skin changed.
func (c *Crumbs) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.refresh(c.stack.Flatten())
}

// StackPushed indicates a new item was added.
func (c *Crumbs) StackPushed(comp model.Component) {
	c.stack.Push(comp)
	c.refresh(c.stack.Flatten())
}

// StackPopped indicates an item was deleted.
func (c *Crumbs) StackPopped(_, _ model.Component) {
	c.stack.Pop()
	c.refresh(c.stack.Flatten())
}

// StackTop indicates the top of the stack.
func (c *Crumbs) StackTop(top model.Component) {}

// Refresh updates view with new crumbs.
func (c *Crumbs) refresh(crumbs []string) {
	c.Clear()
	last, bgColor := len(crumbs)-1, c.styles.Frame().Crumb.BgColor
	for i, crumb := range crumbs {
		if i == last {
			bgColor = c.styles.Frame().Crumb.ActiveColor
		}
		fmt.Fprintf(c, "[%s:%s:b] <%s> [-:%s:-] ",
			c.styles.Frame().Crumb.FgColor,
			bgColor, strings.Replace(strings.ToLower(crumb), " ", "", -1),
			c.styles.Body().BgColor)
	}
}
