package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const defaultPrompt = "%c> %s"

// CmdView captures users free from command input.
type CmdView struct {
	*tview.TextView

	activated bool
	icon      rune
	text      string
	styles    *config.Styles
}

// NewCmdView returns a new command view.
func NewCmdView(styles *config.Styles) *CmdView {
	v := CmdView{styles: styles, TextView: tview.NewTextView()}
	{
		v.SetWordWrap(true)
		v.SetWrap(true)
		v.SetDynamicColors(true)
		v.SetBorder(true)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetBackgroundColor(styles.BgColor())
		// v.SetBorderColor(config.AsColor(styles.Frame().Border.FocusColor))
		v.SetTextColor(styles.FgColor())
	}
	return &v
}

// InCmdMode returns true if command is active, false otherwise.
func (v *CmdView) InCmdMode() bool {
	return v.activated
}

func (v *CmdView) activate() {
	v.write(v.text)
}

func (v *CmdView) update(s string) {
	v.text = s
	v.Clear()
	v.write(s)
}

func (v *CmdView) append(r rune) {
	fmt.Fprintf(v, string(r))
}

func (v *CmdView) write(s string) {
	fmt.Fprintf(v, defaultPrompt, v.icon, s)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

// BufferChanged indicates the buffer was changed.
func (v *CmdView) BufferChanged(s string) {
	v.update(s)
}

// BufferActive indicates the buff activity changed.
func (v *CmdView) BufferActive(f bool, k BufferKind) {
	v.activated = f
	if f {
		v.SetBorder(true)
		v.icon = iconFor(k)
		v.SetTextColor(v.styles.FgColor())
		v.SetBorderColor(colorFor(k))
		v.activate()
	} else {
		v.SetBorder(false)
		v.SetBackgroundColor(v.styles.BgColor())
		v.Clear()
	}
}

func colorFor(k BufferKind) tcell.Color {
	switch k {
	case CommandBuff:
		return tcell.ColorAqua
	default:
		return tcell.ColorSeaGreen
	}
}
func iconFor(k BufferKind) rune {
	switch k {
	case CommandBuff:
		return '🐶'
	default:
		return '🤓'
	}
}
