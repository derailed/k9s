package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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
	v.SetWordWrap(true)
	v.SetWrap(true)
	v.SetDynamicColors(true)
	v.SetBorder(true)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetBackgroundColor(styles.BgColor())
	v.SetTextColor(styles.FgColor())

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
	if v.text == s {
		return
	}
	v.text = s
	v.Clear()
	v.write(v.text)
}

func (v *CmdView) write(s string) {
	fmt.Fprintf(v, defaultPrompt, v.icon, s)
}

func (v *CmdView) reset() {
	v.update("")
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

// BufferChanged indicates the buffer was changed.
func (v *CmdView) BufferChanged(s string) {
	v.update(s)
}

// BufferActive indicates the buff activity changed.
func (v *CmdView) BufferActive(f bool, k BufferKind) {
	if v.activated = f; f {
		v.SetBorder(true)
		v.SetTextColor(v.styles.FgColor())
		v.SetBorderColor(colorFor(k))
		v.icon = iconFor(k)
		v.reset()
		v.activate()
	} else {
		v.SetBorder(false)
		v.SetBackgroundColor(v.styles.BgColor())
		v.Clear()
	}
	log.Debug().Msgf("CmdView activated: %t", v.activated)
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
		return '🐩'
	}
}
