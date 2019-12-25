package render

// TableData tracks a K8s resource for tabular display.
type TableData struct {
	Header    HeaderRow
	RowEvents RowEvents
	Namespace string
}
