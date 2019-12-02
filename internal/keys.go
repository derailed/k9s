package internal

// ContextKey represents context key.
type ContextKey string

const (
	// Factory represents a factory context key.
	KeyFactory   ContextKey = "factory"
	KeySelection            = "selection"
	KeyLabels               = "labels"
	KeyFields               = "fields"
	KeyTable                = "table"
	KeyDir                  = "dir"
)
