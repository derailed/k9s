package model

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
)

// Reconcile previous vs current state and emits delta events.
func Reconcile(ctx context.Context, table render.TableData, gvr client.GVR) (render.TableData, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("RECONCILE elapsed: %v", time.Since(t))
	}(time.Now())

	path, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return table, fmt.Errorf("no path in context for %s", gvr)
	}
	log.Debug().Msgf("Reconcile %q in ns %q with path %q", gvr, table.Namespace, path)
	factory, ok := ctx.Value(internal.KeyFactory).(Factory)
	if !ok {
		return table, fmt.Errorf("no Factory in context for %s", gvr)
	}
	m, ok := Registry[string(gvr)]
	if !ok {
		log.Warn().Msgf("Resource %s not found in registry. Going generic!", gvr)
		m = ResourceMeta{
			Model:    &Generic{},
			Renderer: &render.Generic{},
		}
	}
	if m.Model == nil {
		m.Model = &Resource{}
	}
	m.Model.Init(table.Namespace, string(gvr), factory)

	oo, err := m.Model.List(ctx)
	if err != nil {
		return table, err
	}
	log.Debug().Msgf("Model returned [%d] items", len(oo))

	rows := make(render.Rows, len(oo))
	if err := m.Model.Hydrate(oo, rows, m.Renderer); err != nil {
		return table, err
	}
	update(&table, rows)
	table.Header = m.Renderer.Header(table.Namespace)

	log.Debug().Msgf("Table returned [%d] events", len(table.RowEvents))
	return table, nil
}

func update(table *render.TableData, rows render.Rows) {
	cacheEmpty := len(table.RowEvents) == 0
	kk := make([]string, 0, len(rows))
	var blankDelta render.DeltaRow
	for _, row := range rows {
		kk = append(kk, row.ID)
		if cacheEmpty {
			table.RowEvents = append(table.RowEvents, render.NewRowEvent(render.EventAdd, row))
			continue
		}
		if index, ok := table.RowEvents.FindIndex(row.ID); ok {
			delta := render.NewDeltaRow(table.RowEvents[index].Row, row, table.Header.HasAge())
			if delta.IsBlank() {
				table.RowEvents[index].Kind, table.RowEvents[index].Deltas = render.EventUnchanged, blankDelta
			} else {
				table.RowEvents[index] = render.NewDeltaRowEvent(row, delta)
			}
			continue
		}
		table.RowEvents = append(table.RowEvents, render.NewRowEvent(render.EventAdd, row))
	}

	if cacheEmpty {
		return
	}
	ensureDeletes(table, kk)
}

// EnsureDeletes delete items in cache that are no longer valid.
func ensureDeletes(table *render.TableData, newKeys []string) {
	for _, re := range table.RowEvents {
		var found bool
		for i, key := range newKeys {
			if key == re.Row.ID {
				found = true
				newKeys = append(newKeys[:i], newKeys[i+1:]...)
				break
			}
		}
		if !found {
			table.RowEvents = table.RowEvents.Delete(re.Row.ID)
		}
	}
}
