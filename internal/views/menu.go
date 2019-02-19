package views

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const (
	menuFmt      = " [dodgerblue::b]%s[white::d]%s "
	menuSepFmt   = " [dodgerblue::b]<%s> [white::d]%s "
	menuIndexFmt = " [fuchsia::b]<%d> [white::d]%s "
	maxRows      = 5
	colLen       = 20
)

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
		*tview.Table
	}
)

func (h hints) Len() int {
	return len(h)
}
func (h hints) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h hints) Less(i, j int) bool {
	n, err1 := strconv.Atoi(h[i].mnemonic)
	m, err2 := strconv.Atoi(h[j].mnemonic)

	if err1 == nil && err2 == nil {
		return n < m
	}

	d := strings.Compare(h[i].mnemonic, h[j].mnemonic)
	return d < 0
}

func newKeyHandler(d string, a keyboardHandler) keyAction {
	return keyAction{description: d, action: a}
}

func newMenuView() *menuView {
	v := menuView{tview.NewTable()}
	return &v
}

func (v *menuView) setMenu(hh hints) {
	v.Clear()
	sort.Sort(hh)

	var row, col int
	firstNS, firstCmd := true, true
	for _, h := range hh {
		_, err := strconv.Atoi(h.mnemonic)
		if err == nil && firstNS {
			row = 0
			col = 2
			firstNS = false
		}

		if err != nil && firstCmd {
			row = 0
			col = 0
			firstCmd = false
		}
		c := tview.NewTableCell(v.item(h))
		v.SetCell(row, col, c)
		row++
		if row > maxRows {
			col++
			row = 0
		}
	}
}

func (v *menuView) item(h hint) string {
	i, err := strconv.Atoi(h.mnemonic)
	if err == nil {
		return fmt.Sprintf(menuIndexFmt, i, resource.Truncate(h.display, 14))
	}

	var s string
	if len(h.mnemonic) == 1 {
		s = fmt.Sprintf(menuSepFmt, strings.ToLower(h.mnemonic), h.display)
	} else {
		s = fmt.Sprintf(menuSepFmt, strings.ToUpper(h.mnemonic), h.display)
	}
	return s
}

func (a keyActions) toHints() hints {
	kk := make([]int, 0, len(a))
	for k := range a {
		kk = append(kk, int(k))
	}
	sort.Ints(kk)

	hh := make(hints, 0, len(a))
	for _, k := range kk {
		if name, ok := tcell.KeyNames[tcell.Key(k)]; ok {
			hh = append(hh, hint{
				mnemonic: name,
				display:  a[tcell.Key(k)].description})
		}
	}
	return hh
}

// -----------------------------------------------------------------------------
// Key mapping Constants

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
	KeyHelp = 63
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
