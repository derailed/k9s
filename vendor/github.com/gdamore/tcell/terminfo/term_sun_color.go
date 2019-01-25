// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// Sun Microsystems Workstation console with color support (IA systems)
	AddTerminfo(&Terminfo{
		Name:         "sun-color",
		Columns:      80,
		Lines:        34,
		Colors:       8,
		Bell:         "\a",
		Clear:        "\f",
		AttrOff:      "\x1b[m",
		Bold:         "\x1b[1m",
		Reverse:      "\x1b[7m",
		SetFg:        "\x1b[3%p1%dm",
		SetBg:        "\x1b[4%p1%dm",
		SetFgBg:      "\x1b[3%p1%d;4%p2%dm",
		PadChar:      "\x00",
		SetCursor:    "\x1b[%i%p1%d;%p2%dH",
		CursorBack1:  "\b",
		CursorUp1:    "\x1b[A",
		KeyUp:        "\x1b[A",
		KeyDown:      "\x1b[B",
		KeyRight:     "\x1b[C",
		KeyLeft:      "\x1b[D",
		KeyInsert:    "\x1b[247z",
		KeyDelete:    "\u007f",
		KeyBackspace: "\b",
		KeyHome:      "\x1b[214z",
		KeyEnd:       "\x1b[220z",
		KeyPgUp:      "\x1b[216z",
		KeyPgDn:      "\x1b[222z",
		KeyF1:        "\x1b[224z",
		KeyF2:        "\x1b[225z",
		KeyF3:        "\x1b[226z",
		KeyF4:        "\x1b[227z",
		KeyF5:        "\x1b[228z",
		KeyF6:        "\x1b[229z",
		KeyF7:        "\x1b[230z",
		KeyF8:        "\x1b[231z",
		KeyF9:        "\x1b[232z",
		KeyF10:       "\x1b[233z",
		KeyF11:       "\x1b[234z",
		KeyF12:       "\x1b[235z",
	})
}
