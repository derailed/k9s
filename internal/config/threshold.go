package config

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
)

const (
	// DefCon1 tracks high severity.
	DefCon1 DefConLevel = iota + 1

	// DefCon2 tracks warn level.
	DefCon2

	// DefCon3 tracks medium level.
	DefCon3

	// DefCon4 tracks low level.
	DefCon4

	// DefCon5 tracks all cool.
	DefCon5
)

// DefConLevel tracks defcon severity.
type DefConLevel int

// DefCon tracks a resource alert level.
type DefCon struct {
	Levels []int `yaml:"defcon,omitempty"`
}

// NewDefCon returns a new instance.
func NewDefCon() *DefCon {
	return &DefCon{Levels: []int{90, 80, 75, 70}}
}

// Validate checks all thresholds and make sure we're cool. If not use defaults.
func (d *DefCon) Validate() {
	norm := NewDefCon()
	if len(d.Levels) < 4 {
		d.Levels = norm.Levels
		return
	}
	for i, level := range d.Levels {
		if !d.isValidRange(level) {
			d.Levels[i] = norm.Levels[i]
		}
	}
}

// String returns defcon settings a string.
func (d *DefCon) String() string {
	ss := make([]string, len(d.Levels))
	for i := 0; i < len(d.Levels); i++ {
		ss[i] = render.PrintPerc(d.Levels[i])
	}
	return strings.Join(ss, "|")
}

func (d *DefCon) isValidRange(v int) bool {
	if v < 0 || v > 100 {
		return false
	}

	return true
}

// Threshold tracks threshold to alert user when excided.
type Threshold map[string]*DefCon

// NewThreshold returns a new threshold.
func NewThreshold() Threshold {
	return Threshold{
		"cpu":    NewDefCon(),
		"memory": NewDefCon(),
	}
}

// Validate a namespace is setup correctly
func (t Threshold) Validate(c client.Connection, ks KubeSettings) {
	for _, k := range []string{"cpu", "memory"} {
		v, ok := t[k]
		if !ok {
			t[k] = NewDefCon()
		} else {
			v.Validate()
		}
	}
}

// DefConFor returns a defcon level for the current state.
func (t Threshold) DefConFor(k string, v int) DefConLevel {
	dc, ok := t[k]
	if !ok || v < 0 || v > 100 {
		return DefCon5
	}
	for i, l := range dc.Levels {
		if v >= l {
			return dcLevelFor(i)
		}
	}

	return DefCon5
}

// DefConColorFor returns an defcon level associated level.
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
