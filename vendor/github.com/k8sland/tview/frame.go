package tview

import (
	"github.com/gdamore/tcell"
)

// frameText holds information about a line of text shown in the frame.
type frameText struct {
	Text   string      // The text to be displayed.
	Header bool        // true = place in header, false = place in footer.
	Align  int         // One of the Align constants.
	Color  tcell.Color // The text color.
}

// Frame is a wrapper which adds a border around another primitive. The top area
// (header) and the bottom area (footer) may also contain text.
//
// See https://github.com/rivo/tview/wiki/Frame for an example.
type Frame struct {
	*Box

	// The contained primitive.
	primitive Primitive

	// The lines of text to be displayed.
	text []*frameText

	// Border spacing.
	top, bottom, header, footer, left, right int
}

// NewFrame returns a new frame around the given primitive. The primitive's
// size will be changed to fit within this frame.
func NewFrame(primitive Primitive) *Frame {
	box := NewBox()

	f := &Frame{
		Box:       box,
		primitive: primitive,
		top:       1,
		bottom:    1,
		header:    1,
		footer:    1,
		left:      1,
		right:     1,
	}

	f.focus = f

	return f
}

// AddText adds text to the frame. Set "header" to true if the text is to appear
// in the header, above the contained primitive. Set it to false for it to
// appear in the footer, below the contained primitive. "align" must be one of
// the Align constants. Rows in the header are printed top to bottom, rows in
// the footer are printed bottom to top. Note that long text can overlap as
// different alignments will be placed on the same row.
func (f *Frame) AddText(text string, header bool, align int, color tcell.Color) *Frame {
	f.text = append(f.text, &frameText{
		Text:   text,
		Header: header,
		Align:  align,
		Color:  color,
	})
	return f
}

// Clear removes all text from the frame.
func (f *Frame) Clear() *Frame {
	f.text = nil
	return f
}

// SetBorders sets the width of the frame borders as well as "header" and
// "footer", the vertical space between the header and footer text and the
// contained primitive (does not apply if there is no text).
func (f *Frame) SetBorders(top, bottom, header, footer, left, right int) *Frame {
	f.top, f.bottom, f.header, f.footer, f.left, f.right = top, bottom, header, footer, left, right
	return f
}

// Draw draws this primitive onto the screen.
func (f *Frame) Draw(screen tcell.Screen) {
	f.Box.Draw(screen)

	// Calculate start positions.
	x, top, width, height := f.GetInnerRect()
	bottom := top + height - 1
	x += f.left
	top += f.top
	bottom -= f.bottom
	width -= f.left + f.right
	if width <= 0 || top >= bottom {
		return // No space left.
	}

	// Draw text.
	var rows [6]int // top-left, top-center, top-right, bottom-left, bottom-center, bottom-right.
	topMax := top
	bottomMin := bottom
	for _, text := range f.text {
		// Where do we place this text?
		var y int
		if text.Header {
			y = top + rows[text.Align]
			rows[text.Align]++
			if y >= bottomMin {
				continue
			}
			if y+1 > topMax {
				topMax = y + 1
			}
		} else {
			y = bottom - rows[3+text.Align]
			rows[3+text.Align]++
			if y <= topMax {
				continue
			}
			if y-1 < bottomMin {
				bottomMin = y - 1
			}
		}

		// Draw text.
		Print(screen, text.Text, x, y, width, text.Align, text.Color)
	}

	// Set the size of the contained primitive.
	if topMax > top {
		top = topMax + f.header
	}
	if bottomMin < bottom {
		bottom = bottomMin - f.footer
	}
	if top > bottom {
		return // No space for the primitive.
	}
	f.primitive.SetRect(x, top, width, bottom+1-top)

	// Finally, draw the contained primitive.
	f.primitive.Draw(screen)
}

// Focus is called when this primitive receives focus.
func (f *Frame) Focus(delegate func(p Primitive)) {
	delegate(f.primitive)
}

// HasFocus returns whether or not this primitive has focus.
func (f *Frame) HasFocus() bool {
	focusable, ok := f.primitive.(Focusable)
	if ok {
		return focusable.HasFocus()
	}
	return false
}
