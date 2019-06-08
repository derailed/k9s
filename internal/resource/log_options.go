package resource

import (
	"strings"

	"github.com/derailed/k9s/internal/color"
)

// LogOptions represent logger options.
type LogOptions struct {
	Namespace, Name, Container string
	Lines                      int64
	Previous                   bool
	Color                      color.Paint
}

// HasContainer checks if a container is present.
func (o LogOptions) HasContainer() bool {
	return o.Container != ""
}

// FQN returns resource fully qualified name.
func (o LogOptions) FQN() string {
	return fqn(o.Namespace, o.Name)
}

// Path returns resource descriptor path.
func (o LogOptions) Path() string {
	return o.FQN() + ":" + o.Container
}

// FixedSizeName returns a normalize fixed size pod name if possible.
func (o LogOptions) FixedSizeName() string {
	tokens := strings.Split(o.Name, "-")
	if len(tokens) < 3 {
		return o.Name
	}
	var s []string
	for i := 0; i < len(tokens)-1; i++ {
		s = append(s, tokens[i])
	}
	return Truncate(strings.Join(s, "-"), 15) + "-" + tokens[len(tokens)-1]
}

// NormalizeName colorizes a pod name.
func (o LogOptions) NormalizeName() string {
	if o.Color == 0 {
		return ""
	}
	return color.Colorize(o.Name+":"+o.Container+" ", o.Color)
}
