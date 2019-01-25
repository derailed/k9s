// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// televideo 970
	AddTerminfo(&Terminfo{
		Name:         "tvi970",
		Columns:      80,
		Lines:        24,
		Clear:        "\x1b[H\x1b[2J",
		EnterCA:      "\x1b[?20l\x1b[?7h\x1b[1Q",
		AttrOff:      "\x1b[m",
		Underline:    "\x1b[4m",
		PadChar:      "\x00",
		EnterAcs:     "\x1b(B",
		ExitAcs:      "\x1b(B",
		SetCursor:    "\x1b[%i%p1%d;%p2%df",
		CursorBack1:  "\b",
		CursorUp1:    "\x1bM",
		KeyUp:        "\x1b[A",
		KeyDown:      "\x1b[B",
		KeyRight:     "\x1b[C",
		KeyLeft:      "\x1b[D",
		KeyBackspace: "\b",
		KeyHome:      "\x1b[H",
		KeyF1:        "\x1b?a",
		KeyF2:        "\x1b?b",
		KeyF3:        "\x1b?c",
		KeyF4:        "\x1b?d",
		KeyF5:        "\x1b?e",
		KeyF6:        "\x1b?f",
		KeyF7:        "\x1b?g",
		KeyF8:        "\x1b?h",
		KeyF9:        "\x1b?i",
	})
}
