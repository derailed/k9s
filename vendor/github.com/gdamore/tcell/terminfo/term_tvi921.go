// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// televideo model 921 with sysline same as page & real vi function
	AddTerminfo(&Terminfo{
		Name:         "tvi921",
		Columns:      80,
		Lines:        24,
		Clear:        "\x1a",
		ShowCursor:   "\x1b.3",
		AttrOff:      "\x1bG0",
		Underline:    "\x1bG8",
		Reverse:      "\x1bG4",
		PadChar:      "\x00",
		EnterAcs:     "\x1b$",
		ExitAcs:      "\x1b%%",
		SetCursor:    "\x1b=%p1%' '%+%c%p2%' '%+%c$<3/>",
		CursorBack1:  "\b",
		CursorUp1:    "\v",
		KeyUp:        "\v",
		KeyDown:      "\x16",
		KeyRight:     "\f",
		KeyLeft:      "\b",
		KeyInsert:    "\x1bQ",
		KeyDelete:    "\x1bW",
		KeyBackspace: "\b",
		KeyClear:     "\x1a",
	})
}
