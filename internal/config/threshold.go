package config

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
)

const (
	DefCon1 DefConLevel = iota + 1
	DefCon2
	DefCon3
	DefCon4
	DefCon5
)

type DefConLevel int

// DefCon tracks a resource alert level.
type DefCon [4]int

func newDefCon() DefCon {
	return DefCon{90, 80, 75, 70}
}

func (d DefCon) validate() {
	dc := newDefCon()
	for i := range d {
		if !d.isValidRange(d[i]) {
			d[i] = dc[i]
		}
	}
}

func (d DefCon) String() string {
	ss := make([]string, len(d))
	for i := 0; i < len(d); i++ {
		ss[i] = render.PrintPerc(d[i])
	}
	return strings.Join(ss, "|")
}

func (d DefCon) isValidRange(v int) bool {
	if v == 0 || v > 100 {
		return false
	}

	return true
}

// Threshold tracks threshold to alert user when excided.
type Threshold map[string]DefCon

func NewThreshold() Threshold {
	return Threshold{
		"cpu":    newDefCon(),
		"memory": newDefCon(),
	}
}

// Validate a namespace is setup correctly
func (t Threshold) Validate(c client.Connection, ks KubeSettings) {
	for _, k := range []string{"cpu", "memory"} {
		v, ok := t[k]
		if !ok {
			t[k] = newDefCon()
		} else {
			v.validate()
		}
	}
}

// DefConFor returns a defcon level for the current state.
func (t Threshold) DefConFor(k string, v int) DefConLevel {
	dc, ok := t[k]
	if !ok || v < 0 || v > 100 {
		return DefCon5
	}
	for i, l := range dc {
		if v >= l {
			return dcLevelFor(i)
		}
	}

	return DefCon5
}

func (t *Threshold) DefConColorFor(k string, v int) string {
	switch t.DefConFor(k, v) {
	case DefCon1:
		return "red"
	case DefCon2:
		return "orangered"
	case DefCon3:
		return "orange"
	default:
		return "green"
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func dcLevelFor(l int) DefConLevel {
	switch l {
	case 0:
		return DefCon1
	case 1:
		return DefCon2
	case 2:
		return DefCon3
	case 3:
		return DefCon4
	default:
		return DefCon5
	}
}
