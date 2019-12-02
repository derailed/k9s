package render

// TableData tracks a K8s resource for tabular display.
type TableData struct {
	Header    HeaderRow
	RowEvents RowEvents
	Namespace string
}

func (t TableData) Clone() TableData {
	return TableData{
		Header:    t.Header,
		RowEvents: t.RowEvents.Clone(),
		Namespace: t.Namespace,
	}
}
