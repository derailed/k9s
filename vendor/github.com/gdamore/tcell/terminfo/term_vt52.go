// Generated automatically.  DO NOT HAND-EDIT.

package terminfo

func init() {
	// dec vt52
	AddTerminfo(&Terminfo{
		Name:         "vt52",
		Columns:      80,
		Lines:        24,
		Bell:         "\a",
		Clear:        "\x1bH\x1bJ",
		PadChar:      "\x00",
		AltChars:     "+h.k0affggolpnqprrss",
		EnterAcs:     "\x1bF",
		ExitAcs:      "\x1bG",
		SetCursor:    "\x1bY%p1%' '%+%c%p2%' '%+%c",
		CursorBack1:  "\x1bD",
		CursorUp1:    "\x1bA",
		KeyUp:        "\x1bA",
		KeyDown:      "\x1bB",
		KeyRight:     "\x1bC",
		KeyLeft:      "\x1bD",
		KeyBackspace: "\b",
	})
}
