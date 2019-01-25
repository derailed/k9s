package tview

import (
	"github.com/gdamore/tcell"
)

// Modal is a centered message window used to inform the user or prompt them
// for an immediate decision. It needs to have at least one button (added via
// AddButtons()) or it will never disappear.
//
// See https://github.com/rivo/tview/wiki/Modal for an example.
type Modal struct {
	*Box

	// The framed embedded in the modal.
	frame *Frame

	// The form embedded in the modal's frame.
	form *Form

	// The message text (original, not word-wrapped).
	text string

	// The text color.
	textColor tcell.Color

	// The optional callback for when the user clicked one of the buttons. It
	// receives the index of the clicked button and the button's label.
	done func(buttonIndex int, buttonLabel string)
}

// NewModal returns a new modal message window.
func NewModal() *Modal {
	m := &Modal{
		Box:       NewBox(),
		textColor: Styles.PrimaryTextColor,
	}
	m.form = NewForm().
		SetButtonsAlign(AlignCenter).
		SetButtonBackgroundColor(Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(Styles.PrimaryTextColor)
	m.form.SetBackgroundColor(Styles.ContrastBackgroundColor).SetBorderPadding(0, 0, 0, 0)
	m.form.SetCancelFunc(func() {
		if m.done != nil {
			m.done(-1, "")
		}
	})
	m.frame = NewFrame(m.form).SetBorders(0, 0, 1, 0, 0, 0)
	m.frame.SetBorder(true).
		SetBackgroundColor(Styles.ContrastBackgroundColor).
		SetBorderPadding(1, 1, 1, 1)
	m.focus = m
	return m
}

// SetTextColor sets the color of the message text.
func (m *Modal) SetTextColor(color tcell.Color) *Modal {
	m.textColor = color
	return m
}

// SetDoneFunc sets a handler which is called when one of the buttons was
// pressed. It receives the index of the button as well as its label text. The
// handler is also called when the user presses the Escape key. The index will
// then be negative and the label text an emptry string.
func (m *Modal) SetDoneFunc(handler func(buttonIndex int, buttonLabel string)) *Modal {
	m.done = handler
	return m
}

// SetText sets the message text of the window. The text may contain line
// breaks. Note that words are wrapped, too, based on the final size of the
// window.
func (m *Modal) SetText(text string) *Modal {
	m.text = text
	return m
}

// AddButtons adds buttons to the window. There must be at least one button and
// a "done" handler so the window can be closed again.
func (m *Modal) AddButtons(labels []string) *Modal {
	for index, label := range labels {
		func(i int, l string) {
			m.form.AddButton(label, func() {
				if m.done != nil {
					m.done(i, l)
				}
			})
			button := m.form.GetButton(m.form.GetButtonCount() - 1)
			button.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyDown, tcell.KeyRight:
					return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
				case tcell.KeyUp, tcell.KeyLeft:
					return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
				}
				return event
			})
		}(index, label)
	}
	return m
}

// Focus is called when this primitive receives focus.
func (m *Modal) Focus(delegate func(p Primitive)) {
	delegate(m.form)
}

// HasFocus returns whether or not this primitive has focus.
func (m *Modal) HasFocus() bool {
	return m.form.HasFocus()
}

// Draw draws this primitive onto the screen.
func (m *Modal) Draw(screen tcell.Screen) {
	// Calculate the width of this modal.
	buttonsWidth := 0
	for _, button := range m.form.buttons {
		buttonsWidth += StringWidth(button.label) + 4 + 2
	}
	buttonsWidth -= 2
	screenWidth, screenHeight := screen.Size()
	width := screenWidth / 3
	if width < buttonsWidth {
		width = buttonsWidth
	}
	// width is now without the box border.

	// Reset the text and find out how wide it is.
	m.frame.Clear()
	lines := WordWrap(m.text, width)
	for _, line := range lines {
		m.frame.AddText(line, true, AlignCenter, m.textColor)
	}

	// Set the modal's position and size.
	height := len(lines) + 6
	width += 4
	x := (screenWidth - width) / 2
	y := (screenHeight - height) / 2
	m.SetRect(x, y, width, height)

	// Draw the frame.
	m.frame.SetRect(x, y, width, height)
	m.frame.Draw(screen)
}
