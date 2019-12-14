package internal

// ContextKey represents context key.
type ContextKey string

const (
	// Factory represents a factory context key.
	KeyFactory    ContextKey = "factory"
	KeyLabels                = "labels"
	KeyFields                = "fields"
	KeyTable                 = "table"
	KeyDir                   = "dir"
	KeyPath                  = "path"
	KeySubject               = "subject"
	KeyGVR                   = "gvr"
	KeyForwards              = "forwards"
	KeyContainers            = "containers"
	KeyBenchCfg              = "benchcfg"
	KeyAliases               = "aliases"
	KeyUID                   = "uid"
)
