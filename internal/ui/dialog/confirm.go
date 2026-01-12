// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const (
	dialogKey = "dialog"
)

type confirmFunc func()

// ScrollableModal represents a modal dialog with scrollable content.
type ScrollableModal struct {
	*tview.Box

	frame    *tview.Frame
	textView *tview.TextView
	form     *tview.Form
	done     func(int, string)
}

// NewScrollableModal creates a new scrollable modal dialog.
func NewScrollableModal(title string, text string, form *tview.Form, textColor tcell.Color) *ScrollableModal {
	m := &ScrollableModal{
		Box:  tview.NewBox(),
		form: form,
	}

	// Create scrollable text view for the message
	m.textView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true).
		SetWordWrap(true).
		SetText(text).
		SetTextColor(textColor)

	// Create a flex container to hold text view and form
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(m.textView, 0, 1, false).
		AddItem(m.form, 3, 0, true)

	// Create frame around the content
	m.frame = tview.NewFrame(flex).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBorderPadding(1, 1, 1, 1).
		SetTitle(title)

	return m
}

// Draw draws this primitive onto the screen.
func (m *ScrollableModal) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	// Set modal to 70% of screen size
	width := (screenWidth * 70) / 100
	height := (screenHeight * 70) / 100

	// Center the modal
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2

	m.SetRect(x, y, width, height)
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)

	// Draw scrollbar indicator
	m.drawScrollbar(screen, x, y, width, height)
}

// drawScrollbar draws a scrollbar on the right side of the text view.
func (m *ScrollableModal) drawScrollbar(screen tcell.Screen, x, y, width, height int) {
	// Get the text view's dimensions (accounting for frame borders and padding)
	// Frame has border (1) + padding (1) on each side
	textViewHeight := height - 3 - 4 // Subtract form height (3) and frame borders/padding (4)
	if textViewHeight <= 0 {
		return
	}

	row, _ := m.textView.GetScrollOffset()

	// Get total number of lines in the text view
	// We need to estimate based on the text content
	text := m.textView.GetText(false)
	totalLines := 0
	for _, ch := range text {
		if ch == '\n' {
			totalLines++
		}
	}
	totalLines++ // Account for the last line

	// Only draw scrollbar if content is scrollable
	if totalLines <= textViewHeight {
		return
	}

	// Calculate scrollbar position
	scrollbarX := x + width - 2 // Position on the right side, inside border
	scrollbarStartY := y + 2    // Start after frame border and padding
	scrollbarEndY := scrollbarStartY + textViewHeight - 1

	// Calculate thumb position and size
	thumbHeight := max(1, (textViewHeight*textViewHeight)/totalLines)
	maxScroll := totalLines - textViewHeight
	if maxScroll <= 0 {
		maxScroll = 1
	}
	thumbPosition := (row * (textViewHeight - thumbHeight)) / maxScroll

	// Draw scrollbar track and thumb
	for i := 0; i < textViewHeight; i++ {
		currentY := scrollbarStartY + i
		if currentY > scrollbarEndY {
			break
		}

		var char rune
		var style tcell.Style
		if i >= thumbPosition && i < thumbPosition+thumbHeight {
			// Draw thumb
			char = '█'
			style = tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault)
		} else {
			// Draw track
			char = '│'
			style = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorDefault)
		}
		screen.SetContent(scrollbarX, currentY, char, nil, style)
	}

	// Draw up/down arrows
	screen.SetContent(scrollbarX, scrollbarStartY, '▲', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	screen.SetContent(scrollbarX, scrollbarEndY, '▼', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
}

// SetDoneFunc sets the callback for when the modal is closed.
func (m *ScrollableModal) SetDoneFunc(handler func(int, string)) *ScrollableModal {
	m.done = handler
	return m
}

// Focus is called when this primitive receives focus.
func (m *ScrollableModal) Focus(delegate func(p tview.Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether this primitive has focus.
func (m *ScrollableModal) HasFocus() bool {
	return m.form.HasFocus()
}

// MouseHandler returns the mouse handler for this primitive.
func (m *ScrollableModal) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		// Pass mouse events on to the form
		consumed, capture = m.form.MouseHandler()(action, event, setFocus)
		if !consumed && action == tview.MouseLeftClick && m.InRect(event.Position()) {
			setFocus(m)
			consumed = true
		}
		return
	})
}

// InputHandler returns the handler for this primitive.
func (m *ScrollableModal) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		// Allow arrow keys to scroll the text view when form doesn't have focus on input field
		if event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown || event.Key() == tcell.KeyPgUp || event.Key() == tcell.KeyPgDn {
			if handler := m.textView.InputHandler(); handler != nil {
				handler(event, setFocus)
				return
			}
		}
		// Pass other events to the form
		if handler := m.form.InputHandler(); handler != nil {
			handler(event, setFocus)
		}
	})
}

func ShowConfirmAck(app *ui.App, pages *ui.Pages, acceptStr string, override bool, title, msg string, ack confirmFunc, cancel cancelFunc) {
	styles := app.Styles.Dialog()

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		cancel()
	})

	var accept bool
	if override {
		changedFn := func(t string) {
			accept = (t == acceptStr)
		}
		f.AddInputField("Confirm:", "", 30, nil, changedFn)
	} else {
		accept = true
	}

	f.AddButton("OK", func() {
		if !accept {
			return
		}
		ack()
		dismissConfirm(pages)
		cancel()
	})
	for i := range 2 {
		b := f.GetButton(i)
		if b == nil {
			continue
		}
		b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
		b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
	f.SetFocus(0)
	modal := NewScrollableModal("<"+title+">", msg, f, styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissConfirm(pages)
		cancel()
	})
	pages.AddPage(confirmKey, modal, false, false)
	pages.ShowPage(confirmKey)
}

// ShowConfirm pops a confirmation dialog.
func ShowConfirm(styles *config.Dialog, pages *ui.Pages, title, msg string, ack confirmFunc, cancel cancelFunc) {
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())
	f.AddButton("Cancel", func() {
		dismiss(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		ack()
		dismiss(pages)
		cancel()
	})
	for i := range 2 {
		if b := f.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}
	f.SetFocus(0)
	modal := NewScrollableModal("<"+title+">", msg, f, styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismiss(pages)
		cancel()
	})
	pages.AddPage(dialogKey, modal, false, false)
	pages.ShowPage(dialogKey)
}

func dismiss(pages *ui.Pages) {
	pages.RemovePage(dialogKey)
}
