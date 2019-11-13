package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
)

func TestDefaultColorer(t *testing.T) {
	uu := map[string]struct {
		re resource.RowEvent
		e  tcell.Color
	}{
		"def": {resource.RowEvent{}, ui.StdColor},
		"new": {resource.RowEvent{Action: resource.New}, ui.AddColor},
		"add": {resource.RowEvent{Action: watch.Added}, ui.AddColor},
		"upd": {resource.RowEvent{Action: watch.Modified}, ui.ModColor},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, ui.DefaultColorer("", &u.re))
		})
	}
}
