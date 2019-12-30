package ui

import "github.com/gdamore/tcell"

func init() {
	initKeys()
}

func initKeys() {
	tcell.KeyNames[tcell.Key(KeyHelp)] = "?"
	tcell.KeyNames[tcell.Key(KeySlash)] = "/"
	tcell.KeyNames[tcell.Key(KeySpace)] = "space"

	initNumbKeys()
	initStdKeys()
	initShiftKeys()
}

// Defines numeric keys for container actions
const (
	Key0 int32 = iota + 48
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

// Defines numeric keys for container actions
const (
	KeyShift0 int32 = 41
	KeyShift1 int32 = 33
	KeyShift2 int32 = 64
	KeyShift3 int32 = 35
	KeyShift4 int32 = 36
	KeyShift5 int32 = 37
	KeyShift6 int32 = 94
	KeyShift7 int32 = 38
	KeyShift8 int32 = 42
	KeyShift9 int32 = 40
)

// Defines char keystrokes
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

// Define Shift Keys
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
var NumKeys = map[int]int32{
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
	tcell.KeyNames[tcell.Key(Key0)] = "0"
	tcell.KeyNames[tcell.Key(Key1)] = "1"
	tcell.KeyNames[tcell.Key(Key2)] = "2"
	tcell.KeyNames[tcell.Key(Key3)] = "3"
	tcell.KeyNames[tcell.Key(Key4)] = "4"
	tcell.KeyNames[tcell.Key(Key5)] = "5"
	tcell.KeyNames[tcell.Key(Key6)] = "6"
	tcell.KeyNames[tcell.Key(Key7)] = "7"
	tcell.KeyNames[tcell.Key(Key8)] = "8"
	tcell.KeyNames[tcell.Key(Key9)] = "9"
}

func initStdKeys() {
	tcell.KeyNames[tcell.Key(KeyA)] = "a"
	tcell.KeyNames[tcell.Key(KeyB)] = "b"
	tcell.KeyNames[tcell.Key(KeyC)] = "c"
	tcell.KeyNames[tcell.Key(KeyD)] = "d"
	tcell.KeyNames[tcell.Key(KeyE)] = "e"
	tcell.KeyNames[tcell.Key(KeyF)] = "f"
	tcell.KeyNames[tcell.Key(KeyG)] = "g"
	tcell.KeyNames[tcell.Key(KeyH)] = "h"
	tcell.KeyNames[tcell.Key(KeyI)] = "i"
	tcell.KeyNames[tcell.Key(KeyJ)] = "j"
	tcell.KeyNames[tcell.Key(KeyK)] = "k"
	tcell.KeyNames[tcell.Key(KeyL)] = "l"
	tcell.KeyNames[tcell.Key(KeyM)] = "m"
	tcell.KeyNames[tcell.Key(KeyN)] = "n"
	tcell.KeyNames[tcell.Key(KeyO)] = "o"
	tcell.KeyNames[tcell.Key(KeyP)] = "p"
	tcell.KeyNames[tcell.Key(KeyQ)] = "q"
	tcell.KeyNames[tcell.Key(KeyR)] = "r"
	tcell.KeyNames[tcell.Key(KeyS)] = "s"
	tcell.KeyNames[tcell.Key(KeyT)] = "t"
	tcell.KeyNames[tcell.Key(KeyU)] = "u"
	tcell.KeyNames[tcell.Key(KeyV)] = "v"
	tcell.KeyNames[tcell.Key(KeyW)] = "w"
	tcell.KeyNames[tcell.Key(KeyX)] = "x"
	tcell.KeyNames[tcell.Key(KeyY)] = "y"
	tcell.KeyNames[tcell.Key(KeyZ)] = "z"
}

func initShiftKeys() {
	tcell.KeyNames[tcell.Key(KeyShift0)] = "Shift-0"
	tcell.KeyNames[tcell.Key(KeyShift1)] = "Shift-1"
	tcell.KeyNames[tcell.Key(KeyShift2)] = "Shift-2"
	tcell.KeyNames[tcell.Key(KeyShift3)] = "Shift-3"
	tcell.KeyNames[tcell.Key(KeyShift4)] = "Shift-4"
	tcell.KeyNames[tcell.Key(KeyShift5)] = "Shift-5"
	tcell.KeyNames[tcell.Key(KeyShift6)] = "Shift-6"
	tcell.KeyNames[tcell.Key(KeyShift7)] = "Shift-7"
	tcell.KeyNames[tcell.Key(KeyShift8)] = "Shift-8"
	tcell.KeyNames[tcell.Key(KeyShift9)] = "Shift-9"

	tcell.KeyNames[tcell.Key(KeyShiftA)] = "Shift-A"
	tcell.KeyNames[tcell.Key(KeyShiftB)] = "Shift-B"
	tcell.KeyNames[tcell.Key(KeyShiftC)] = "Shift-C"
	tcell.KeyNames[tcell.Key(KeyShiftD)] = "Shift-D"
	tcell.KeyNames[tcell.Key(KeyShiftE)] = "Shift-E"
	tcell.KeyNames[tcell.Key(KeyShiftF)] = "Shift-F"
	tcell.KeyNames[tcell.Key(KeyShiftG)] = "Shift-G"
	tcell.KeyNames[tcell.Key(KeyShiftH)] = "Shift-H"
	tcell.KeyNames[tcell.Key(KeyShiftI)] = "Shift-I"
	tcell.KeyNames[tcell.Key(KeyShiftJ)] = "Shift-J"
	tcell.KeyNames[tcell.Key(KeyShiftK)] = "Shift-K"
	tcell.KeyNames[tcell.Key(KeyShiftL)] = "Shift-L"
	tcell.KeyNames[tcell.Key(KeyShiftM)] = "Shift-M"
	tcell.KeyNames[tcell.Key(KeyShiftN)] = "Shift-N"
	tcell.KeyNames[tcell.Key(KeyShiftO)] = "Shift-O"
	tcell.KeyNames[tcell.Key(KeyShiftP)] = "Shift-P"
	tcell.KeyNames[tcell.Key(KeyShiftQ)] = "Shift-Q"
	tcell.KeyNames[tcell.Key(KeyShiftR)] = "Shift-R"
	tcell.KeyNames[tcell.Key(KeyShiftS)] = "Shift-S"
	tcell.KeyNames[tcell.Key(KeyShiftT)] = "Shift-T"
	tcell.KeyNames[tcell.Key(KeyShiftU)] = "Shift-U"
	tcell.KeyNames[tcell.Key(KeyShiftV)] = "Shift-V"
	tcell.KeyNames[tcell.Key(KeyShiftW)] = "Shift-W"
	tcell.KeyNames[tcell.Key(KeyShiftX)] = "Shift-X"
	tcell.KeyNames[tcell.Key(KeyShiftY)] = "Shift-Y"
	tcell.KeyNames[tcell.Key(KeyShiftZ)] = "Shift-Z"
}
