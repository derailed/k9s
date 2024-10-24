// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import "github.com/derailed/tcell/v2"

func init() {
	initKeys()
}

func initKeys() {
	tcell.KeyNames[KeyHelp] = "?"
	tcell.KeyNames[KeySlash] = "/"
	tcell.KeyNames[KeySpace] = "space"

	initNumbKeys()
	initStdKeys()
	initShiftKeys()
	initShiftNumKeys()
}

// Defines numeric keys for container actions.
const (
	Key0 tcell.Key = iota + 48
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
)

// Defines numeric keys for container actions.
const (
	KeyShift0 tcell.Key = 41
	KeyShift1 tcell.Key = 33
	KeyShift2 tcell.Key = 64
	KeyShift3 tcell.Key = 35
	KeyShift4 tcell.Key = 36
	KeyShift5 tcell.Key = 37
	KeyShift6 tcell.Key = 94
	KeyShift7 tcell.Key = 38
	KeyShift8 tcell.Key = 42
	KeyShift9 tcell.Key = 40
)

// Defines char keystrokes.
const (
	KeyA tcell.Key = iota + 97
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
	KeyHelp  = 63
	KeySlash = 47
	KeyColon = 58
	KeySpace = 32
)

// Define Shift Keys.
const (
	KeyShiftA tcell.Key = iota + 65
	KeyShiftB
	KeyShiftC
	KeyShiftD
	KeyShiftE
	KeyShiftF
	KeyShiftG
	KeyShiftH
	KeyShiftI
	KeyShiftJ
	KeyShiftK
	KeyShiftL
	KeyShiftM
	KeyShiftN
	KeyShiftO
	KeyShiftP
	KeyShiftQ
	KeyShiftR
	KeyShiftS
	KeyShiftT
	KeyShiftU
	KeyShiftV
	KeyShiftW
	KeyShiftX
	KeyShiftY
	KeyShiftZ
)

// NumKeys tracks number keys.
var NumKeys = map[int]tcell.Key{
	0: Key0,
	1: Key1,
	2: Key2,
	3: Key3,
	4: Key4,
	5: Key5,
	6: Key6,
	7: Key7,
	8: Key8,
	9: Key9,
}

func initNumbKeys() {
	tcell.KeyNames[Key0] = "0"
	tcell.KeyNames[Key1] = "1"
	tcell.KeyNames[Key2] = "2"
	tcell.KeyNames[Key3] = "3"
	tcell.KeyNames[Key4] = "4"
	tcell.KeyNames[Key5] = "5"
	tcell.KeyNames[Key6] = "6"
	tcell.KeyNames[Key7] = "7"
	tcell.KeyNames[Key8] = "8"
	tcell.KeyNames[Key9] = "9"
}

func initStdKeys() {
	tcell.KeyNames[KeyA] = "a"
	tcell.KeyNames[KeyB] = "b"
	tcell.KeyNames[KeyC] = "c"
	tcell.KeyNames[KeyD] = "d"
	tcell.KeyNames[KeyE] = "e"
	tcell.KeyNames[KeyF] = "f"
	tcell.KeyNames[KeyG] = "g"
	tcell.KeyNames[KeyH] = "h"
	tcell.KeyNames[KeyI] = "i"
	tcell.KeyNames[KeyJ] = "j"
	tcell.KeyNames[KeyK] = "k"
	tcell.KeyNames[KeyL] = "l"
	tcell.KeyNames[KeyM] = "m"
	tcell.KeyNames[KeyN] = "n"
	tcell.KeyNames[KeyO] = "o"
	tcell.KeyNames[KeyP] = "p"
	tcell.KeyNames[KeyQ] = "q"
	tcell.KeyNames[KeyR] = "r"
	tcell.KeyNames[KeyS] = "s"
	tcell.KeyNames[KeyT] = "t"
	tcell.KeyNames[KeyU] = "u"
	tcell.KeyNames[KeyV] = "v"
	tcell.KeyNames[KeyW] = "w"
	tcell.KeyNames[KeyX] = "x"
	tcell.KeyNames[KeyY] = "y"
	tcell.KeyNames[KeyZ] = "z"
}

func initShiftNumKeys() {
	tcell.KeyNames[KeyShift0] = "Shift-0"
	tcell.KeyNames[KeyShift1] = "Shift-1"
	tcell.KeyNames[KeyShift2] = "Shift-2"
	tcell.KeyNames[KeyShift3] = "Shift-3"
	tcell.KeyNames[KeyShift4] = "Shift-4"
	tcell.KeyNames[KeyShift5] = "Shift-5"
	tcell.KeyNames[KeyShift6] = "Shift-6"
	tcell.KeyNames[KeyShift7] = "Shift-7"
	tcell.KeyNames[KeyShift8] = "Shift-8"
	tcell.KeyNames[KeyShift9] = "Shift-9"
}

func initShiftKeys() {
	tcell.KeyNames[KeyShiftA] = "Shift-A"
	tcell.KeyNames[KeyShiftB] = "Shift-B"
	tcell.KeyNames[KeyShiftC] = "Shift-C"
	tcell.KeyNames[KeyShiftD] = "Shift-D"
	tcell.KeyNames[KeyShiftE] = "Shift-E"
	tcell.KeyNames[KeyShiftF] = "Shift-F"
	tcell.KeyNames[KeyShiftG] = "Shift-G"
	tcell.KeyNames[KeyShiftH] = "Shift-H"
	tcell.KeyNames[KeyShiftI] = "Shift-I"
	tcell.KeyNames[KeyShiftJ] = "Shift-J"
	tcell.KeyNames[KeyShiftK] = "Shift-K"
	tcell.KeyNames[KeyShiftL] = "Shift-L"
	tcell.KeyNames[KeyShiftM] = "Shift-M"
	tcell.KeyNames[KeyShiftN] = "Shift-N"
	tcell.KeyNames[KeyShiftO] = "Shift-O"
	tcell.KeyNames[KeyShiftP] = "Shift-P"
	tcell.KeyNames[KeyShiftQ] = "Shift-Q"
	tcell.KeyNames[KeyShiftR] = "Shift-R"
	tcell.KeyNames[KeyShiftS] = "Shift-S"
	tcell.KeyNames[KeyShiftT] = "Shift-T"
	tcell.KeyNames[KeyShiftU] = "Shift-U"
	tcell.KeyNames[KeyShiftV] = "Shift-V"
	tcell.KeyNames[KeyShiftW] = "Shift-W"
	tcell.KeyNames[KeyShiftX] = "Shift-X"
	tcell.KeyNames[KeyShiftY] = "Shift-Y"
	tcell.KeyNames[KeyShiftZ] = "Shift-Z"
}
