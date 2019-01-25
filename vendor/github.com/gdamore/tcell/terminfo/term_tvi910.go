// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// televideo model 910
	AddTerminfo(&Terminfo{
		Name:         "tvi910",
		Columns:      80,
		Lines:        24,
		Bell:         "\a",
		Clear:        "\x1a",
		AttrOff:      "\x1bG0",
		Underline:    "\x1bG8",
		Reverse:      "\x1bG4",
		PadChar:      "\x00",
		SetCursor:    "\x1b=%p1%' '%+%c%p2%' '%+%c",
		CursorBack1:  "\b",
		CursorUp1:    "\v",
		KeyUp:        "\v",
		KeyDown:      "\n",
		KeyRight:     "\f",
		KeyLeft:      "\b",
		KeyBackspace: "\b",
		KeyHome:      "\x1e",
		KeyF1:        "\x01@\r",
		KeyF2:        "\x01A\r",
		KeyF3:        "\x01B\r",
		KeyF4:        "\x01C\r",
		KeyF5:        "\x01D\r",
		KeyF6:        "\x01E\r",
		KeyF7:        "\x01F\r",
		KeyF8:        "\x01G\r",
		KeyF9:        "\x01H\r",
	})
}
