// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// Darwin PowerPC Console (color)
	AddTerminfo(&Terminfo{
		Name:         "xnuppc",
		Aliases:      []string{"darwin"},
		Colors:       8,
		Clear:        "\x1b[H\x1b[J",
		AttrOff:      "\x1b[m",
		Underline:    "\x1b[4m",
		Bold:         "\x1b[1m",
		Reverse:      "\x1b[7m",
		EnterKeypad:  "\x1b[?1h\x1b=",
		ExitKeypad:   "\x1b[?1l\x1b>",
		SetFg:        "\x1b[3%p1%dm",
		SetBg:        "\x1b[4%p1%dm",
		SetFgBg:      "\x1b[3%p1%d;4%p2%dm",
		PadChar:      "\x00",
		SetCursor:    "\x1b[%i%p1%d;%p2%dH",
		CursorBack1:  "\x1b[D",
		CursorUp1:    "\x1b[A",
		KeyUp:        "\x1bOA",
		KeyDown:      "\x1bOB",
		KeyRight:     "\x1bOC",
		KeyLeft:      "\x1bOD",
		KeyBackspace: "\u007f",
	})
}
