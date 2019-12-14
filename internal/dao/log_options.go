package dao

import (
	"strings"

	"github.com/derailed/k9s/internal/color"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
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
	_, n := k8s.Namespaced(o.Path)
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
	_, n := k8s.Namespaced(o.Path)
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

// Helpers...

// BOZO!! Consolidate!!
// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

// BOZO!!
// // Namespaced return a namesapace and a name.
// func Namespaced(n string) (string, string) {
// 	ns, po := path.Split(n)

// 	return strings.Trim(ns, "/"), po
// }

// // FQN returns a fully qualified resource name.
// func FQN(ns, n string) string {
// 	if ns == "" {
// 		return n
// 	}
// 	return ns + "/" + n
// }
