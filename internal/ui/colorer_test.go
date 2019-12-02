package ui_test

// BOZO!!
// import (
// 	"testing"

// 	"github.com/derailed/k9s/internal/render"
// 	"github.com/derailed/k9s/internal/ui"
// 	"github.com/gdamore/tcell"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDefaultColorer(t *testing.T) {
// 	uu := map[string]struct {
// 		re render.RowEvent
// 		e  tcell.Color
// 	}{
// 		"default": {render.RowEvent{}, ui.StdColor},
// 		"add":     {render.RowEvent{Kind: render.EventAdd}, ui.AddColor},
// 		"delete":  {render.RowEvent{Kind: render.EventDelete}, ui.KillColor},
// 		"update":  {render.RowEvent{Kind: render.EventUpdate}, ui.ModColor},
// 	}

// 	for k := range uu {
// 		u := uu[k]
// 		t.Run(k, func(t *testing.T) {
// 			assert.Equal(t, u.e, ui.DefaultColorer("", u.re))
// 		})
// 	}
// }
