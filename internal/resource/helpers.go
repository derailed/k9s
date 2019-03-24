package resource

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	// DefaultNamespace indicator to fetch default namespace.
	DefaultNamespace = "default"
	// AllNamespace namespace name to span all namespaces.
	AllNamespace = "all"
	// AllNamespaces indicator to retrieve K8s resource for all namespaces.
	AllNamespaces = ""
	// NotNamespaced indicator for non namespaced resource.
	NotNamespaced = "-"

	// New track new resource events.
	New watch.EventType = "NEW"
	// Unchanged provides no change events.
	Unchanged watch.EventType = "UNCHANGED"

	// MissingValue indicates an unset value.
	MissingValue = "<none>"
	// NAValue indicates a value that does not pertain.
	NAValue = "<n/a>"
)

func namespaced(n string) (string, string) {
	ns, po := path.Split(n)

	return strings.Trim(ns, "/"), po
}

func missing(s string) string {
	return check(s, MissingValue)
}

func na(s string) string {
	return check(s, NAValue)
}

func check(s, sub string) string {
	if len(s) == 0 {
		return sub
	}

	return s
}

func intToStr(i int64) string {
	return strconv.Itoa(int(i))
}

func boolToStr(b bool) string {
	switch b {
	case true:
		return "true"
	default:
		return "false"
	}
}

func toAge(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "<unknown>"
	}

	return duration.HumanDuration(time.Since(timestamp.Time))
}

// Pad a string up to the given length.
func Pad(s string, l int) string {
	fmat := "%-" + strconv.Itoa(l) + "s"

	return fmt.Sprintf(fmat, s)
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

func mapToStr(m map[string]string) (s string) {
	if len(m) == 0 {
		return MissingValue
	}

	kk := make([]string, 0, len(m))
	for k := range m {
		kk = append(kk, k)
	}
	sort.Strings(kk)

	for i, k := range kk {
		s += k + "=" + m[k]
		if i < len(kk)-1 {
			s += ","
		}
	}

	return
}

// ToMillicore shows cpu reading for human.
func ToMillicore(v int64) string {
	return strconv.Itoa(int(v)) + "m"
}

// ToMi shows mem reading for human.
func ToMi(v float64) string {
	return strconv.Itoa(int(v)) + "Mi"
}
