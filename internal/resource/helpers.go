package resource

import (
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

// AsPerc prints a number as a percentage.
func AsPerc(f float64) string {
	return strconv.Itoa(int(f)) + "%"
}

// ToPerc computes the ratio of two numbers as a percentage.
func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return (v1 / v2) * 100
}

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
	return time.Since(timestamp.Time).String()
}

func toAgeHuman(s string) string {
	d, err := time.ParseDuration(s)
	if err != nil {
		return "<unknown>"
	}

	return duration.HumanDuration(d)
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

func boolPtrToStr(b *bool) string {
	if b == nil {
		return "false"
	}

	return boolToStr(*b)
}
