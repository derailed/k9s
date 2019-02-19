package views

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, addColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, modColor},
		// MoChange AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, stdColor},
		// Bust NS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: term}, errColor},
		// Bust NS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: dead}, errColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, addColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, addColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, modColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, modColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: failNS}, errColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: failNoNS}, errColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: killNS}, killColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: killNoNS}, killColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, addColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: noNs}, addColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, errColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, errColor},
		// Nochange AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, stdColor},
		// Nochange NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: noNs}, stdColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, addColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, addColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, modColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, modColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, errColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, errColor},
		// Unchanged cool AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, stdColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: ns}, addColor},
		// Add NS
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: nonNS}, addColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ns}, modColor},
		// Mod NS
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: nonNS}, modColor},
		// Unchanged cool
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ns}, stdColor},
		// Bust AllNS
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNS}, errColor},
		// Bust NS
		{"blee", &resource.RowEvent{Action: resource.Unchanged, Fields: bustNoNS}, errColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, dpColorer(u.ns, u.r))
	}
}

func TestPVColorer(t *testing.T) {
	var (
		pv     = resource.Row{"blee", "1G", "RO", "Duh", "Bound"}
		bustPv = resource.Row{"blee", "1G", "RO", "Duh", "UnBound"}
	)

	uu := colorerUCs{
		// Add Normal
		{"", &resource.RowEvent{Action: watch.Added, Fields: pv}, addColor},
		// Unchanged Bound
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: pv}, stdColor},
		// Unchanged Bound
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustPv}, errColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: pvc}, addColor},
		// Add Bound
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: bustPvc}, errColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: ctx}, addColor},
		// Add Default
		{"", &resource.RowEvent{Action: watch.Added, Fields: defCtx}, addColor},
		// Mod Normal
		{"", &resource.RowEvent{Action: watch.Modified, Fields: ctx}, modColor},
		// Mod Default
		{"", &resource.RowEvent{Action: watch.Modified, Fields: defCtx}, modColor},
		// Unchanged Normal
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: ctx}, stdColor},
		// Unchanged Default
		{"", &resource.RowEvent{Action: resource.Unchanged, Fields: defCtx}, highlightColor},
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
		{"", &resource.RowEvent{Action: watch.Added, Fields: nsRow}, addColor},
		// Add Namespaced
		{"blee", &resource.RowEvent{Action: watch.Added, Fields: row}, addColor},
		// Mod AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: nsRow}, modColor},
		// Mod Namespaced
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: row}, modColor},
		// Mod Busted AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: toastNS}, errColor},
		// Mod Busted Namespaced
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: toast}, errColor},
		// NotReady AllNS
		{"", &resource.RowEvent{Action: watch.Modified, Fields: notReadyNS}, errColor},
		// NotReady Namespaced
		{"blee", &resource.RowEvent{Action: watch.Modified, Fields: notReady}, errColor},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, podColorer(u.ns, u.r))
	}
}
