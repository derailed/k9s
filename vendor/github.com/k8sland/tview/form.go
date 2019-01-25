package tview

import (
	"github.com/gdamore/tcell"
)

// DefaultFormFieldWidth is the default field screen width of form elements
// whose field width is flexible (0). This is used in the Form class for
// horizontal layouts.
var DefaultFormFieldWidth = 10

// FormItem is the interface all form items must implement to be able to be
// included in a form.
type FormItem interface {
	Primitive

	// GetLabel returns the item's label text.
	GetLabel() string

	// SetFormAttributes sets a number of item attributes at once.
	SetFormAttributes(labelWidth int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) FormItem

	// GetFieldWidth returns the width of the form item's field (the area which
	// is manipulated by the user) in number of screen cells. A value of 0
	// indicates the the field width is flexible and may use as much space as
	// required.
	GetFieldWidth() int

	// SetFinishedFunc sets the handler function for when the user finished
	// entering data into the item. The handler may receive events for the
	// Enter key (we're done), the Escape key (cancel input), the Tab key (move to
	// next field), and the Backtab key (move to previous field).
	SetFinishedFunc(handler func(key tcell.Key)) FormItem
}

// Form allows you to combine multiple one-line form elements into a vertical
// or horizontal layout. Form elements include types such as InputField or
// Checkbox. These elements can be optionally followed by one or more buttons
// for which you can define form-wide actions (e.g. Save, Clear, Cancel).
//
// See https://github.com/rivo/tview/wiki/Form for an example.
type Form struct {
	*Box

	// The items of the form (one row per item).
	items []FormItem

	// The buttons of the form.
	buttons []*Button

	// If set to true, instead of position items and buttons from top to bottom,
	// they are positioned from left to right.
	horizontal bool

	// The alignment of the buttons.
	buttonsAlign int

	// The number of empty rows between items.
	itemPadding int

	// The index of the item or button which has focus. (Items are counted first,
	// buttons are counted last.)
	focusedElement int

	// The label color.
	labelColor tcell.Color

	// The background color of the input area.
	fieldBackgroundColor tcell.Color

	// The text color of the input area.
	fieldTextColor tcell.Color

	// The background color of the buttons.
	buttonBackgroundColor tcell.Color

	// The color of the button text.
	buttonTextColor tcell.Color

	// An optional function which is called when the user hits Escape.
	cancel func()
}

// NewForm returns a new form.
func NewForm() *Form {
	box := NewBox().SetBorderPadding(1, 1, 1, 1)

	f := &Form{
		Box:                   box,
		itemPadding:           1,
		labelColor:            Styles.SecondaryTextColor,
		fieldBackgroundColor:  Styles.ContrastBackgroundColor,
		fieldTextColor:        Styles.PrimaryTextColor,
		buttonBackgroundColor: Styles.ContrastBackgroundColor,
		buttonTextColor:       Styles.PrimaryTextColor,
	}

	f.focus = f

	return f
}

// SetItemPadding sets the number of empty rows between form items for vertical
// layouts and the number of empty cells between form items for horizontal
// layouts.
func (f *Form) SetItemPadding(padding int) *Form {
	f.itemPadding = padding
	return f
}

// SetHorizontal sets the direction the form elements are laid out. If set to
// true, instead of positioning them from top to bottom (the default), they are
// positioned from left to right, moving into the next row if there is not
// enough space.
func (f *Form) SetHorizontal(horizontal bool) *Form {
	f.horizontal = horizontal
	return f
}

// SetLabelColor sets the color of the labels.
func (f *Form) SetLabelColor(color tcell.Color) *Form {
	f.labelColor = color
	return f
}

// SetFieldBackgroundColor sets the background color of the input areas.
func (f *Form) SetFieldBackgroundColor(color tcell.Color) *Form {
	f.fieldBackgroundColor = color
	return f
}

// SetFieldTextColor sets the text color of the input areas.
func (f *Form) SetFieldTextColor(color tcell.Color) *Form {
	f.fieldTextColor = color
	return f
}

// SetButtonsAlign sets how the buttons align horizontally, one of AlignLeft
// (the default), AlignCenter, and AlignRight. This is only
func (f *Form) SetButtonsAlign(align int) *Form {
	f.buttonsAlign = align
	return f
}

// SetButtonBackgroundColor sets the background color of the buttons.
func (f *Form) SetButtonBackgroundColor(color tcell.Color) *Form {
	f.buttonBackgroundColor = color
	return f
}

// SetButtonTextColor sets the color of the button texts.
func (f *Form) SetButtonTextColor(color tcell.Color) *Form {
	f.buttonTextColor = color
	return f
}

// AddInputField adds an input field to the form. It has a label, an optional
// initial value, a field width (a value of 0 extends it as far as possible),
// an optional accept function to validate the item's value (set to nil to
// accept any text), and an (optional) callback function which is invoked when
// the input field's text has changed.
func (f *Form) AddInputField(label, value string, fieldWidth int, accept func(textToCheck string, lastChar rune) bool, changed func(text string)) *Form {
	f.items = append(f.items, NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldWidth(fieldWidth).
		SetAcceptanceFunc(accept).
		SetChangedFunc(changed))
	return f
}

// AddPasswordField adds a password field to the form. This is similar to an
// input field except that the user's input not shown. Instead, a "mask"
// character is displayed. The password field has a label, an optional initial
// value, a field width (a value of 0 extends it as far as possible), and an
// (optional) callback function which is invoked when the input field's text has
// changed.
func (f *Form) AddPasswordField(label, value string, fieldWidth int, mask rune, changed func(text string)) *Form {
	if mask == 0 {
		mask = '*'
	}
	f.items = append(f.items, NewInputField().
		SetLabel(label).
		SetText(value).
		SetFieldWidth(fieldWidth).
		SetMaskCharacter(mask).
		SetChangedFunc(changed))
	return f
}

// AddDropDown adds a drop-down element to the form. It has a label, options,
// and an (optional) callback function which is invoked when an option was
// selected. The initial option may be a negative value to indicate that no
// option is currently selected.
func (f *Form) AddDropDown(label string, options []string, initialOption int, selected func(option string, optionIndex int)) *Form {
	f.items = append(f.items, NewDropDown().
		SetLabel(label).
		SetCurrentOption(initialOption).
		SetOptions(options, selected))
	return f
}

// AddCheckbox adds a checkbox to the form. It has a label, an initial state,
// and an (optional) callback function which is invoked when the state of the
// checkbox was changed by the user.
func (f *Form) AddCheckbox(label string, checked bool, changed func(checked bool)) *Form {
	f.items = append(f.items, NewCheckbox().
		SetLabel(label).
		SetChecked(checked).
		SetChangedFunc(changed))
	return f
}

// AddButton adds a new button to the form. The "selected" function is called
// when the user selects this button. It may be nil.
func (f *Form) AddButton(label string, selected func()) *Form {
	f.buttons = append(f.buttons, NewButton(label).SetSelectedFunc(selected))
	return f
}

// GetButton returns the button at the specified 0-based index. Note that
// buttons have been specially prepared for this form and modifying some of
// their attributes may have unintended side effects.
func (f *Form) GetButton(index int) *Button {
	return f.buttons[index]
}

// RemoveButton removes the button at the specified position, starting with 0
// for the button that was added first.
func (f *Form) RemoveButton(index int) *Form {
	f.buttons = append(f.buttons[:index], f.buttons[index+1:]...)
	return f
}

// GetButtonCount returns the number of buttons in this form.
func (f *Form) GetButtonCount() int {
	return len(f.buttons)
}

// GetButtonIndex returns the index of the button with the given label, starting
// with 0 for the button that was added first. If no such label was found, -1
// is returned.
func (f *Form) GetButtonIndex(label string) int {
	for index, button := range f.buttons {
		if button.GetLabel() == label {
			return index
		}
	}
	return -1
}

// Clear removes all input elements from the form, including the buttons if
// specified.
func (f *Form) Clear(includeButtons bool) *Form {
	f.items = nil
	if includeButtons {
		f.buttons = nil
	}
	f.focusedElement = 0
	return f
}

// AddFormItem adds a new item to the form. This can be used to add your own
// objects to the form. Note, however, that the Form class will override some
// of its attributes to make it work in the form context. Specifically, these
// are:
//
//   - The label width
//   - The label color
//   - The background color
//   - The field text color
//   - The field background color
func (f *Form) AddFormItem(item FormItem) *Form {
	f.items = append(f.items, item)
	return f
}

// GetFormItem returns the form element at the given position, starting with
// index 0. Elements are referenced in the order they were added. Buttons are
// not included.
func (f *Form) GetFormItem(index int) FormItem {
	return f.items[index]
}

// RemoveFormItem removes the form element at the given position, starting with
// index 0. Elements are referenced in the order they were added. Buttons are
// not included.
func (f *Form) RemoveFormItem(index int) *Form {
	f.items = append(f.items[:index], f.items[index+1:]...)
	return f
}

// GetFormItemByLabel returns the first form element with the given label. If
// no such element is found, nil is returned. Buttons are not searched and will
// therefore not be returned.
func (f *Form) GetFormItemByLabel(label string) FormItem {
	for _, item := range f.items {
		if item.GetLabel() == label {
			return item
		}
	}
	return nil
}

// GetFormItemIndex returns the index of the first form element with the given
// label. If no such element is found, -1 is returned. Buttons are not searched
// and will therefore not be returned.
func (f *Form) GetFormItemIndex(label string) int {
	for index, item := range f.items {
		if item.GetLabel() == label {
			return index
		}
	}
	return -1
}

// SetCancelFunc sets a handler which is called when the user hits the Escape
// key.
func (f *Form) SetCancelFunc(callback func()) *Form {
	f.cancel = callback
	return f
}

// Draw draws this primitive onto the screen.
func (f *Form) Draw(screen tcell.Screen) {
	f.Box.Draw(screen)

	// Determine the dimensions.
	x, y, width, height := f.GetInnerRect()
	topLimit := y
	bottomLimit := y + height
	rightLimit := x + width
	startX := x

	// Find the longest label.
	var maxLabelWidth int
	for _, item := range f.items {
		labelWidth := StringWidth(item.GetLabel())
		if labelWidth > maxLabelWidth {
			maxLabelWidth = labelWidth
		}
	}
	maxLabelWidth++ // Add one space.

	// Calculate positions of form items.
	positions := make([]struct{ x, y, width, height int }, len(f.items)+len(f.buttons))
	var focusedPosition struct{ x, y, width, height int }
	for index, item := range f.items {
		// Calculate the space needed.
		labelWidth := StringWidth(item.GetLabel())
		var itemWidth int
		if f.horizontal {
			fieldWidth := item.GetFieldWidth()
			if fieldWidth == 0 {
				fieldWidth = DefaultFormFieldWidth
			}
			labelWidth++
			itemWidth = labelWidth + fieldWidth
		} else {
			// We want all fields to align vertically.
			labelWidth = maxLabelWidth
			itemWidth = width
		}

		// Advance to next line if there is no space.
		if f.horizontal && x+labelWidth+1 >= rightLimit {
			x = startX
			y += 2
		}

		// Adjust the item's attributes.
		if x+itemWidth >= rightLimit {
			itemWidth = rightLimit - x
		}
		item.SetFormAttributes(
			labelWidth,
			f.labelColor,
			f.backgroundColor,
			f.fieldTextColor,
			f.fieldBackgroundColor,
		)

		// Save position.
		positions[index].x = x
		positions[index].y = y
		positions[index].width = itemWidth
		positions[index].height = 1
		if item.GetFocusable().HasFocus() {
			focusedPosition = positions[index]
		}

		// Advance to next item.
		if f.horizontal {
			x += itemWidth + f.itemPadding
		} else {
			y += 1 + f.itemPadding
		}
	}

	// How wide are the buttons?
	buttonWidths := make([]int, len(f.buttons))
	buttonsWidth := 0
	for index, button := range f.buttons {
		w := StringWidth(button.GetLabel()) + 4
		buttonWidths[index] = w
		buttonsWidth += w + 1
	}
	buttonsWidth--

	// Where do we place them?
	if !f.horizontal && x+buttonsWidth < rightLimit {
		if f.buttonsAlign == AlignRight {
			x = rightLimit - buttonsWidth
		} else if f.buttonsAlign == AlignCenter {
			x = (x + rightLimit - buttonsWidth) / 2
		}

		// In vertical layouts, buttons always appear after an empty line.
		if f.itemPadding == 0 {
			y++
		}
	}

	// Calculate positions of buttons.
	for index, button := range f.buttons {
		space := rightLimit - x
		buttonWidth := buttonWidths[index]
		if f.horizontal {
			if space < buttonWidth-4 {
				x = startX
				y += 2
				space = width
			}
		} else {
			if space < 1 {
				break // No space for this button anymore.
			}
		}
		if buttonWidth > space {
			buttonWidth = space
		}
		button.SetLabelColor(f.buttonTextColor).
			SetLabelColorActivated(f.buttonBackgroundColor).
			SetBackgroundColorActivated(f.buttonTextColor).
			SetBackgroundColor(f.buttonBackgroundColor)

		buttonIndex := index + len(f.items)
		positions[buttonIndex].x = x
		positions[buttonIndex].y = y
		positions[buttonIndex].width = buttonWidth
		positions[buttonIndex].height = 1

		if button.HasFocus() {
			focusedPosition = positions[buttonIndex]
		}

		x += buttonWidth + 1
	}

	// Determine vertical offset based on the position of the focused item.
	var offset int
	if focusedPosition.y+focusedPosition.height > bottomLimit {
		offset = focusedPosition.y + focusedPosition.height - bottomLimit
		if focusedPosition.y-offset < topLimit {
			offset = focusedPosition.y - topLimit
		}
	}

	// Draw items.
	for index, item := range f.items {
		// Set position.
		y := positions[index].y - offset
		height := positions[index].height
		item.SetRect(positions[index].x, y, positions[index].width, height)

		// Is this item visible?
		if y+height <= topLimit || y >= bottomLimit {
			continue
		}

		// Draw items with focus last (in case of overlaps).
		if item.GetFocusable().HasFocus() {
			defer item.Draw(screen)
		} else {
			item.Draw(screen)
		}
	}

	// Draw buttons.
	for index, button := range f.buttons {
		// Set position.
		buttonIndex := index + len(f.items)
		y := positions[buttonIndex].y - offset
		height := positions[buttonIndex].height
		button.SetRect(positions[buttonIndex].x, y, positions[buttonIndex].width, height)

		// Is this button visible?
		if y+height <= topLimit || y >= bottomLimit {
			continue
		}

		// Draw button.
		button.Draw(screen)
	}
}

// Focus is called by the application when the primitive receives focus.
func (f *Form) Focus(delegate func(p Primitive)) {
	if len(f.items)+len(f.buttons) == 0 {
		return
	}

	// Hand on the focus to one of our child elements.
	if f.focusedElement < 0 || f.focusedElement >= len(f.items)+len(f.buttons) {
		f.focusedElement = 0
	}
	handler := func(key tcell.Key) {
		switch key {
		case tcell.KeyTab, tcell.KeyEnter:
			f.focusedElement++
			f.Focus(delegate)
		case tcell.KeyBacktab:
			f.focusedElement--
			if f.focusedElement < 0 {
				f.focusedElement = len(f.items) + len(f.buttons) - 1
			}
			f.Focus(delegate)
		case tcell.KeyEscape:
			if f.cancel != nil {
				f.cancel()
			} else {
				f.focusedElement = 0
				f.Focus(delegate)
			}
		}
	}

	if f.focusedElement < len(f.items) {
		// We're selecting an item.
		item := f.items[f.focusedElement]
		item.SetFinishedFunc(handler)
		delegate(item)
	} else {
		// We're selecting a button.
		button := f.buttons[f.focusedElement-len(f.items)]
		button.SetBlurFunc(handler)
		delegate(button)
	}
}

// HasFocus returns whether or not this primitive has focus.
func (f *Form) HasFocus() bool {
	for _, item := range f.items {
		if item.GetFocusable().HasFocus() {
			return true
		}
	}
	for _, button := range f.buttons {
		if button.focus.HasFocus() {
			return true
		}
	}
	return false
}
