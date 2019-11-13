package ui

type ContextKey string

const (
	// KeyApp designates an application context.
	KeyApp = ContextKey("app")

	// KeyStyles designates the application styles.
	KeyStyles = ContextKey("styles")

	// KeyNamespace designates a namespace context.
	KeyNamespace = ContextKey("ns")
)
