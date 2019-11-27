package view

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

type (
	colorerUC struct {
		ns string
		r  render.RowEvent
		e  tcell.Color
	}
	colorerUCs []colorerUC
)

func TestNSColorer(t *testing.T) {
	var (
		ns   = render.Row{Fields: render.Fields{"blee", "Active"}}
		term = render.Row{Fields: render.Fields{"blee", resource.Terminating}}
		dead = render.Row{Fields: render.Fields{"blee", "Inactive"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{
			Kind: render.EventAdd,
			Row:  ns,
		},
			ui.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ns}, ui.ModColor},
		// MoChange AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ns}, ui.StdColor},
		// Bust NS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: term}, ui.ErrColor},
		// Bust NS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: dead}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, nsColorer(u.ns, u.r))
	}
}

func TestEvColorer(t *testing.T) {
	var (
		ns       = render.Row{Fields: render.Fields{"", "blee", "fred", "Normal"}}
		nonNS    = render.Row{Fields: render.Fields{"", "fred", "Normal"}}
		failNS   = render.Row{Fields: render.Fields{"", "blee", "fred", "Failed"}}
		failNoNS = render.Row{Fields: render.Fields{"", "fred", "Failed"}}
		killNS   = render.Row{Fields: render.Fields{"", "blee", "fred", "Killing"}}
		killNoNS = render.Row{Fields: render.Fields{"", "fred", "Killing"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: ns}, ui.AddColor},
		// Add NS
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ns}, ui.ModColor},
		// Mod NS
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: nonNS}, ui.ModColor},
		// Bust AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: failNS}, ui.ErrColor},
		// Bust NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: failNoNS}, ui.ErrColor},
		// Bust AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: killNS}, ui.KillColor},
		// Bust NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: killNoNS}, ui.KillColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, evColorer(u.ns, u.r))
	}
}

func TestRSColorer(t *testing.T) {
	var (
		ns       = render.Row{Fields: render.Fields{"blee", "fred", "1", "1"}}
		noNs     = render.Row{Fields: render.Fields{"fred", "1", "1"}}
		bustNS   = render.Row{Fields: render.Fields{"blee", "fred", "1", "0"}}
		bustNoNS = render.Row{Fields: render.Fields{"fred", "1", "0"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: ns}, ui.AddColor},
		// Add NS
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: noNs}, ui.AddColor},
		// Bust AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: bustNoNS}, ui.ErrColor},
		// Nochange AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ns}, ui.StdColor},
		// Nochange NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: noNs}, ui.StdColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, rsColorer(u.ns, u.r))
	}
}

func TestStsColorer(t *testing.T) {
	var (
		ns       = render.Row{Fields: render.Fields{"blee", "fred", "1", "1"}}
		nonNS    = render.Row{Fields: render.Fields{"fred", "1", "1"}}
		bustNS   = render.Row{Fields: render.Fields{"blee", "fred", "2", "1"}}
		bustNoNS = render.Row{Fields: render.Fields{"fred", "2", "1"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: ns}, ui.AddColor},
		// Add NS
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ns}, ui.ModColor},
		// Mod NS
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: nonNS}, ui.ModColor},
		// Bust AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: bustNoNS}, ui.ErrColor},
		// Unchanged cool AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ns}, ui.StdColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, stsColorer(u.ns, u.r))
	}
}

func TestDpColorer(t *testing.T) {
	var (
		ns       = render.Row{Fields: render.Fields{"blee", "fred", "1", "1"}}
		nonNS    = render.Row{Fields: render.Fields{"fred", "1", "1"}}
		bustNS   = render.Row{Fields: render.Fields{"blee", "fred", "2", "1"}}
		bustNoNS = render.Row{Fields: render.Fields{"fred", "2", "1"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: ns}, ui.AddColor},
		// Add NS
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ns}, ui.ModColor},
		// Mod NS
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: nonNS}, ui.ModColor},
		// Unchanged cool
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ns}, ui.StdColor},
		// Bust AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: bustNoNS}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, dpColorer(u.ns, u.r))
	}
}

func TestPdbColorer(t *testing.T) {
	var (
		ns       = render.Row{Fields: render.Fields{"blee", "fred", "1", "1", "1", "1", "1"}}
		nonNS    = render.Row{Fields: render.Fields{"fred", "1", "1", "1", "1", "1"}}
		bustNS   = render.Row{Fields: render.Fields{"blee", "fred", "1", "1", "1", "1", "2"}}
		bustNoNS = render.Row{Fields: render.Fields{"fred", "1", "1", "1", "1", "2"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: ns}, ui.AddColor},
		// Add NS
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ns}, ui.ModColor},
		// Mod NS
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: nonNS}, ui.ModColor},
		// Unchanged cool
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ns}, ui.StdColor},
		// Bust AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", render.RowEvent{Kind: render.EventUnchanged, Row: bustNoNS}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, pdbColorer(u.ns, u.r))
	}
}

func TestPVColorer(t *testing.T) {
	var (
		pv     = render.Row{Fields: render.Fields{"blee", "1G", "RO", "Duh", "Bound"}}
		bustPv = render.Row{Fields: render.Fields{"blee", "1G", "RO", "Duh", "UnBound"}}
	)

	uu := colorerUCs{
		// Add Normal
		{"", render.RowEvent{Kind: render.EventAdd, Row: pv}, ui.AddColor},
		// Unchanged Bound
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: pv}, ui.StdColor},
		// Unchanged Bound
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: bustPv}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, pvColorer(u.ns, u.r))
	}
}

func TestPVCColorer(t *testing.T) {
	var (
		pvc     = render.Row{Fields: render.Fields{"blee", "fred", "Bound"}}
		bustPvc = render.Row{Fields: render.Fields{"blee", "fred", "UnBound"}}
	)

	uu := colorerUCs{
		// Add Normal
		{"", render.RowEvent{Kind: render.EventAdd, Row: pvc}, ui.AddColor},
		// Add Bound
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: bustPvc}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, pvcColorer(u.ns, u.r))
	}
}

func TestCtxColorer(t *testing.T) {
	var (
		ctx    = render.Row{Fields: render.Fields{"blee"}}
		defCtx = render.Row{Fields: render.Fields{"blee*"}}
	)

	uu := colorerUCs{
		// Add Normal
		{"", render.RowEvent{Kind: render.EventAdd, Row: ctx}, ui.AddColor},
		// Add Default
		{"", render.RowEvent{Kind: render.EventAdd, Row: defCtx}, ui.AddColor},
		// Mod Normal
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ctx}, ui.ModColor},
		// Mod Default
		{"", render.RowEvent{Kind: render.EventUpdate, Row: defCtx}, ui.ModColor},
		// Unchanged Normal
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ctx}, ui.StdColor},
		// Unchanged Default
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: defCtx}, ui.HighlightColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, ctxColorer(u.ns, u.r))
	}
}

func TestPodColorer(t *testing.T) {
	var (
		nsRow      = render.Row{Fields: render.Fields{"blee", "fred", "1/1", "Running"}}
		toastNS    = render.Row{Fields: render.Fields{"blee", "fred", "1/1", "Boom"}}
		notReadyNS = render.Row{Fields: render.Fields{"blee", "fred", "0/1", "Boom"}}
		row        = render.Row{Fields: render.Fields{"fred", "1/1", "Running"}}
		toast      = render.Row{Fields: render.Fields{"fred", "1/1", "Boom"}}
		notReady   = render.Row{Fields: render.Fields{"fred", "0/1", "Boom"}}
	)

	uu := colorerUCs{
		// Add allNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: nsRow}, ui.AddColor},
		// Add Namespaced
		{"blee", render.RowEvent{Kind: render.EventAdd, Row: row}, ui.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: nsRow}, ui.ModColor},
		// Mod Namespaced
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: row}, ui.ModColor},
		// Mod Busted AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: toastNS}, ui.ErrColor},
		// Mod Busted Namespaced
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: toast}, ui.ErrColor},
		// NotReady AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: notReadyNS}, ui.ErrColor},
		// NotReady Namespaced
		{"blee", render.RowEvent{Kind: render.EventUpdate, Row: notReady}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, podColorer(u.ns, u.r))
	}
}
