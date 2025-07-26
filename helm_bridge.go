package main

import (
	// This import ensures that our internal bridge package is used
	// instead of the real github.com/rubenv/sql-migrate when helm
	// tries to use it during runtime. This is a build-time workaround.
	_ "github.com/derailed/k9s/internal/bridge"
)
