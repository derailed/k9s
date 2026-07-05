// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

type ModalProgress struct {
	*tview.Box

	text  *tview.TextView
	frame *tview.Frame
	body  string
}

func NewModalProgress(title string, styles *config.Dialog, msg string) *ModalProgress {
	m := &ModalProgress{Box: tview.NewBox(), text: tview.NewTextView(), body: msg}

	m.text.
		SetDynamicColors(true).
		SetWrap(false).
		SetText(msg).
		SetTextColor(styles.FgColor.Color()).
		SetBackgroundColor(styles.BgColor.Color())

	m.frame = tview.NewFrame(m.text).SetBorders(0, 0, 0, 0, 0, 0)
	m.frame.
		SetBorder(true).
		SetBorderPadding(1, 1, 1, 1).
		SetBorderColor(styles.FgColor.Color()).
		SetBackgroundColor(styles.BgColor.Color()).
		SetTitle(title).
		SetTitleColor(tcell.ColorAqua)

	return m
}

func (m *ModalProgress) SetText(msg string) {
	m.body = msg
	m.text.SetText(msg)
}

func (m *ModalProgress) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	lines := strings.Split(m.body, "\n")
	width := 40
	for _, line := range lines {
		width = max(width, len(line)+4)
	}
	width = min(width, max(20, screenWidth-4))
	height := min(len(lines)+4, max(5, screenHeight-2))
	x := max(0, (screenWidth-width)/2)
	y := max(0, (screenHeight-height)/2)

	m.SetRect(x, y, width, height)
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}

func (m *ModalProgress) Focus(delegate func(p tview.Primitive)) {
	delegate(m.text)
}

func (m *ModalProgress) HasFocus() bool {
	return m.text.HasFocus()
}

func (m *ModalProgress) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			return true, nil
		}

		return false, nil
	})
}

func (m *ModalProgress) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(_ *tcell.EventKey, _ func(p tview.Primitive)) {})
}