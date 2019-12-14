package dao

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/color"
)

// LogOptions represent logger options.
type LogOptions struct {
	Path            string
	Container       string
	Lines           int64
	Color           color.Paint
	Previous        bool
	SingleContainer bool
	MultiPods       bool
}

// HasContainer checks if a container is present.
func (o LogOptions) HasContainer() bool {
	return o.Container != ""
}

// FixedSizeName returns a normalize fixed size pod name if possible.
func (o LogOptions) FixedSizeName() string {
	_, n := client.Namespaced(o.Path)
	tokens := strings.Split(n, "-")
	if len(tokens) < 3 {
		return n
	}
	var s []string
	for i := 0; i < len(tokens)-1; i++ {
		s = append(s, tokens[i])
	}

	return Truncate(strings.Join(s, "-"), 15) + "-" + tokens[len(tokens)-1]
}

func colorize(c color.Paint, txt string) string {
	if c == 0 {
		return ""
	}

	return color.Colorize(txt, c)
}

// DecorateLog add a log header to display po/co information along with the log message.
func (o LogOptions) DecorateLog(msg string) string {
	_, n := client.Namespaced(o.Path)
	if msg == "" {
		return msg
	}

	if o.MultiPods {
		return colorize(o.Color, n+":"+o.Container+" ") + msg
	}

	if !o.SingleContainer {
		return colorize(o.Color, o.Container+" ") + msg
	}

	return msg
}
