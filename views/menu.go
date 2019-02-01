package views

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const (
	menuFmt      = " [dodgerblue::b]%s[white::d]%s "
	menuSepFmt   = " [dodgerblue::b]<%s> [white::d]%s "
	menuIndexFmt = " [dodgerblue::b]<%d> [white::d]%s "
	maxRows      = 5
	colLen       = 20
)

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

// Defines char keystrokes
const (
	KeyA int32 = iota + 97
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
)

var numKeys = map[int]int32{
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

type (
	keyboardHandler func(*tcell.EventKey)

	hint struct {
		mnemonic, display string
	}
	hints []hint

	hinter interface {
		hints() hints
	}

	keyAction struct {
		description string
		action      keyboardHandler
	}
	keyActions map[tcell.Key]keyAction

	menuView struct {
		*tview.Grid
	}
)

func newKeyHandler(d string, a keyboardHandler) keyAction {
	return keyAction{description: d, action: a}
}

func newMenuView() *menuView {
	v := menuView{tview.NewGrid()}
	v.SetGap(0, 1)
	return &v
}

func (v *menuView) setMenu(hh hints) {
	v.Clear()
	v.SetRows(1, 1, 1, 1)
	v.SetColumns(colLen, colLen)
	isNS := true
	var row, col int
	for _, h := range hh {
		// Reset cols for namespace menus...
		if len(h.mnemonic) == 1 && isNS {
			col++
			row = 0
			isNS = false
		}
		v.AddItem(v.item(h), row, col, 1, 1, 1, 1, false)
		row++
		if row > maxRows {
			col++
			row = 0
		}
	}
}

func (v *menuView) item(h hint) tview.Primitive {
	c := tview.NewTextView()
	c.SetDynamicColors(true)
	var s string
	if i, err := strconv.Atoi(h.mnemonic); err != nil {
		if strings.ToLower(h.display)[0] == h.mnemonic[0] {
			s = fmt.Sprintf(menuFmt, strings.ToUpper(h.mnemonic), h.display[1:])
		} else {
			s = fmt.Sprintf(menuSepFmt, strings.ToUpper(h.mnemonic), h.display)
		}
	} else {
		s = fmt.Sprintf(menuIndexFmt, i, resource.Truncate(h.display, 14))
	}
	c.SetText(s)
	return c
}

func (a keyActions) toHints() hints {
	kk := make([]int, 0, len(a))
	for k := range a {
		kk = append(kk, int(k))
	}
	sort.Ints(kk)
	hh := make(hints, 0, len(a))
	for _, k := range kk {
		hh = append(hh, hint{
			mnemonic: tcell.KeyNames[tcell.Key(k)],
			display:  a[tcell.Key(k)].description})
	}
	return hh
}
