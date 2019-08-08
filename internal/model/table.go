package model

import (
	"github.com/derailed/k9s/internal/resource"
)

// TableListener tracks tabular data changes.
type TableListener interface {
	Refreshed(resource.TableData)
	RowAdded(resource.RowEvent)
	RowUpdated(resource.RowEvent)
	RowDeleted(resource.RowEvent)
}

// Table represents tabular data.
type Table struct {
	data resource.TableData

	listeners []TableListener
}

// NewTable returns a new table.
func NewTable() *Table {
	return &Table{}
}

// Load the initial tabular data
func (t *Table) Load(data resource.TableData) {
	t.data = data
	t.fireTableRefreshed()
}

func (t *Table) fireTableRefreshed() {
	for _, l := range t.listeners {
		l.Refreshed(t.data)
	}
}
