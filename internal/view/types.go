package view

import "github.com/derailed/k9s/internal/model"

// Hinter represents a view that can produce menu hints.
type Hinter interface {
	// Hints returns a collection of hints.
	Hints() model.MenuHints
}
