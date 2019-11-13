package view

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
)

type (
	colorerUC struct {
		ns string
		r  *resource.RowEvent
		e  tcell.Color
	}
	colorerUCs []colorerUC
)

func TestNSColorer(t *testing.T) {
	var (
		ns   = resource.Row{"blee", "Active"}
		term = resource.Row{"blee", "Terminating"}
		dead = resource.Row{"blee", "Inactive"}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, ui.AddColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, ui.ModColor},
		// MoChange AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, ui.StdColor},
		// Bust NS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: term}, ui.ErrColor},
		// Bust NS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: dead}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, nsColorer(u.ns, u.r))
	}
}

func TestEvColorer(t *testing.T) {
	var (
		ns       = resource.Row{"", "blee", "fred", "Normal"}
		nonNS    = resource.Row{"", "fred", "Normal"}
		failNS   = resource.Row{"", "blee", "fred", "Failed"}
		failNoNS = resource.Row{"", "fred", "Failed"}
		killNS   = resource.Row{"", "blee", "fred", "Killing"}
		killNoNS = resource.Row{"", "fred", "Killing"}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, ui.AddColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, ui.ModColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, ui.ModColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: failNS}, ui.ErrColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: failNoNS}, ui.ErrColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: killNS}, ui.KillColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: killNoNS}, ui.KillColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, evColorer(u.ns, u.r))
	}
}

func TestRSColorer(t *testing.T) {
	var (
		ns       = resource.Row{"blee", "fred", "1", "1"}
		noNs     = ns[1:]
		bustNS   = resource.Row{"blee", "fred", "1", "0"}
		bustNoNS = bustNS[1:]
	)

	uu := colorerUCs{
		// Add AllNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, ui.AddColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: noNs}, ui.AddColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, ui.ErrColor},
		// Nochange AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, ui.StdColor},
		// Nochange NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: noNs}, ui.StdColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, rsColorer(u.ns, u.r))
	}
}

func TestStsColorer(t *testing.T) {
	var (
		ns       = resource.Row{"blee", "fred", "1", "1"}
		nonNS    = ns[1:]
		bustNS   = resource.Row{"blee", "fred", "2", "1"}
		bustNoNS = bustNS[1:]
	)

	uu := colorerUCs{
		// Add AllNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, ui.AddColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, ui.ModColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, ui.ModColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, ui.ErrColor},
		// Unchanged cool AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, ui.StdColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, stsColorer(u.ns, u.r))
	}
}

func TestDpColorer(t *testing.T) {
	var (
		ns       = resource.Row{"blee", "fred", "1", "1"}
		nonNS    = ns[1:]
		bustNS   = resource.Row{"blee", "fred", "2", "1"}
		bustNoNS = bustNS[1:]
	)

	uu := colorerUCs{
		// Add AllNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, ui.AddColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, ui.ModColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, ui.ModColor},
		// Unchanged cool
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, ui.StdColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, dpColorer(u.ns, u.r))
	}
}

func TestPdbColorer(t *testing.T) {
	var (
		ns       = resource.Row{"blee", "fred", "1", "1", "1", "1", "1"}
		nonNS    = ns[1:]
		bustNS   = resource.Row{"blee", "fred", "1", "1", "1", "1", "2"}
		bustNoNS = bustNS[1:]
	)

	uu := colorerUCs{
		// Add AllNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, ui.AddColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, ui.AddColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, ui.ModColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, ui.ModColor},
		// Unchanged cool
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, ui.StdColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, ui.ErrColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, pdbColorer(u.ns, u.r))
	}
}

func TestPVColorer(t *testing.T) {
	var (
		pv     = resource.Row{"blee", "1G", "RO", "Duh", "Bound"}
		bustPv = resource.Row{"blee", "1G", "RO", "Duh", "UnBound"}
	)

	uu := colorerUCs{
		// Add Normal
		{"", &resource.RowEvent{Action: watch.Added, Fields: pv}, ui.AddColor},
		// Unchanged Bound
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: pv}, ui.StdColor},
		// Unchanged Bound
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustPv}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, pvColorer(u.ns, u.r))
	}
}

func TestPVCColorer(t *testing.T) {
	var (
		pvc     = resource.Row{"blee", "fred", "Bound"}
		bustPvc = resource.Row{"blee", "fred", "UnBound"}
	)

	uu := colorerUCs{
		// Add Normal
		{"", &resource.RowEvent{Action: watch.Added, Fields: pvc}, ui.AddColor},
		// Add Bound
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustPvc}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, pvcColorer(u.ns, u.r))
	}
}

func TestCtxColorer(t *testing.T) {
	var (
		ctx    = resource.Row{"blee"}
		defCtx = resource.Row{"blee*"}
	)

	uu := colorerUCs{
		// Add Normal
		{"", &resource.RowEvent{Action: watch.Added, Fields: ctx}, ui.AddColor},
		// Add Default
		{"", &resource.RowEvent{Action: watch.Added, Fields: defCtx}, ui.AddColor},
		// Mod Normal
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ctx}, ui.ModColor},
		// Mod Default
		{"", &resource.RowEvent{Action: watch.Modified, Fields: defCtx}, ui.ModColor},
		// Unchanged Normal
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ctx}, ui.StdColor},
		// Unchanged Default
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: defCtx}, ui.HighlightColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, ctxColorer(u.ns, u.r))
	}
}

func TestPodColorer(t *testing.T) {
	var (
		nsRow                = resource.Row{"blee", "fred", "1/1", "Running"}
		toastNS              = resource.Row{"blee", "fred", "1/1", "Boom"}
		notReadyNS           = resource.Row{"blee", "fred", "0/1", "Boom"}
		row, toast, notReady = nsRow[1:], toastNS[1:], notReadyNS[1:]
	)

	uu := colorerUCs{
		// Add allNS
		{"", &resource.RowEvent{Action: watch.Added, Fields: nsRow}, ui.AddColor},
		// Add Namespaced
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: row}, ui.AddColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: nsRow}, ui.ModColor},
		// Mod Namespaced
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: row}, ui.ModColor},
		// Mod Busted AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: toastNS}, ui.ErrColor},
		// Mod Busted Namespaced
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: toast}, ui.ErrColor},
		// NotReady AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: notReadyNS}, ui.ErrColor},
		// NotReady Namespaced
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: notReady}, ui.ErrColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, podColorer(u.ns, u.r))
	}
}
