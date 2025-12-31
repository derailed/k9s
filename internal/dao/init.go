// Package dao contains data access objects
package dao

import (
	"github.com/derailed/k9s/internal/hack"
)

func init() {
	// Replace the sql-migrate dialects with our fixed versions
	// This is needed because the vendored code in sql-migrate doesn't
	// implement all the required methods in the gorp.Dialect interface
	hack.InitDialects()
}
