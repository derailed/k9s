package tview

// Focusable provides a method which determines if a primitive has focus.
// Composed primitives may be focused based on the focused state of their
// contained primitives.
type Focusable interface {
	HasFocus() bool
}
