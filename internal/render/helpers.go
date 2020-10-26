package render

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

var durationRx = regexp.MustCompile(`\A(\d*d)*?(\d*h)*?(\d*m)*?(\d*s)*?\z`)

func durationToSeconds(duration string) string {
	tokens := durationRx.FindAllStringSubmatch(duration, -1)
	if len(tokens) == 0 {
		return duration
	}
	if len(tokens[0]) < 5 {
		return duration
	}

	d, h, m, s := tokens[0][1], tokens[0][2], tokens[0][3], tokens[0][4]
	var n int
	if v, err := strconv.Atoi(strings.Replace(d, "d", "", 1)); err == nil {
		n += v * 24 * 60 * 60
	}
	if v, err := strconv.Atoi(strings.Replace(h, "h", "", 1)); err == nil {
		n += v * 60 * 60
	}
	if v, err := strconv.Atoi(strings.Replace(m, "m", "", 1)); err == nil {
		n += v * 60
	}
	if v, err := strconv.Atoi(strings.Replace(s, "s", "", 1)); err == nil {
		n += v
	}

	return strconv.Itoa(n)
}

// AsThousands prints a number with thousand separator.
func AsThousands(n int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}

// Happy returns true if resoure is happy, false otherwise
func Happy(ns string, h Header, r Row) bool {
	if len(r.Fields) == 0 {
		return true
	}
	validCol := h.IndexOf("VALID", true)
	if validCol < 0 {
		return true
	}
	return strings.TrimSpace(r.Fields[validCol]) == ""
}

func asStatus(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func asSelector(s *metav1.LabelSelector) string {
	sel, err := metav1.LabelSelectorAsSelector(s)
	if err != nil {
		log.Error().Err(err).Msg("Selector conversion failed")
		return NAValue
	}

	return sel.String()
}

type metric struct {
	cpu, mem, cpuLim, memLim string
}

func noMetric() metric {
	return metric{cpu: NAValue, mem: NAValue, cpuLim: NAValue, memLim: NAValue}
}

// ToSelector flattens a map selector to a string selector.
func toSelector(m map[string]string) string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k+"="+v)
	}

	return strings.Join(s, ",")
}

// Blank checks if a collection is empty or all values are blank.
func blank(s []string) bool {
	for _, v := range s {
		if len(v) != 0 {
			return false
		}
	}
	return true
}

// Join a slice of strings, skipping blanks.
func join(a []string, sep string) string {
	switch len(a) {
	case 0:
		return ""
	case 1:
		return a[0]
	}

	var b []string
	for _, s := range a {
		if s != "" {
			b = append(b, s)
		}
	}
	if len(b) == 0 {
		return ""
	}

	n := len(sep) * (len(b) - 1)
	for i := 0; i < len(b); i++ {
		n += len(a[i])
	}

	var buff strings.Builder
	buff.Grow(n)
	buff.WriteString(b[0])
	for _, s := range b[1:] {
		buff.WriteString(sep)
		buff.WriteString(s)
	}

	return buff.String()
}

// PrintPerc prints a number as percentage.
func PrintPerc(p int) string {
	return strconv.Itoa(p) + "%"
}

// IntToStr converts an int to a string.
func IntToStr(p int) string {
	return strconv.Itoa(int(p))
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
		return NAValue
	}

	return duration.HumanDuration(d)
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

func mapToStr(m map[string]string) (s string) {
	if len(m) == 0 {
		return ""
	}

	kk := make([]string, 0, len(m))
	for k := range m {
		kk = append(kk, k)
	}
	sort.Strings(kk)

	for i, k := range kk {
		s += k + "=" + m[k]
		if i < len(kk)-1 {
			s += " "
		}
	}

	return
}

func mapToIfc(m interface{}) (s string) {
	if m == nil {
		return ""
	}

	mm, ok := m.(map[string]interface{})
	if !ok {
		return ""
	}
	if len(mm) == 0 {
		return ""
	}

	kk := make([]string, 0, len(mm))
	for k := range mm {
		kk = append(kk, k)
	}
	sort.Strings(kk)

	for i, k := range kk {
		str, ok := mm[k].(string)
		if !ok {
			continue
		}
		s += k + "=" + str
		if i < len(kk)-1 {
			s += " "
		}
	}

	return
}

// ToResourcesMi prints out request:limit mem resources.
func ToResourcesMi(res resources) string {
	var v1, v2 int64
	if v, ok := res[requestMEM]; ok && v != nil {
		v1 = v.MilliValue()
	}
	if v, ok := res[limitMEM]; ok && v != nil {
		v2 = v.MilliValue()
	}
	if v1 == 0 && v2 == 0 {
		return NAValue
	}
	return bytesToMb(v1) + ":" + bytesToMb(v2)
}

func toMc(v int64) string {
	if v == 0 {
		return NAValue
	}
	p := message.NewPrinter(language.English)
	return p.Sprintf("%dm", v)
}

func bytesToMb(v int64) string {
	if v == 0 {
		return NAValue
	}
	p := message.NewPrinter(language.English)
	return p.Sprintf("%dMi", v/(client.MegaByte*1_000))

}

// ToResourcesMc prints out request:limit cpu resources.
func ToResourcesMc(res resources) string {
	var v1, v2 int64
	if v, ok := res[requestCPU]; ok && v != nil {
		v1 = v.MilliValue()
	}
	if v, ok := res[limitCPU]; ok && v != nil {
		v2 = v.MilliValue()
	}
	if v1 == 0 && v2 == 0 {
		return NAValue
	}

	return toMc(v1) + ":" + toMc(v2)
}

// ToMc returns a the millicore unit.
func ToMc(v int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%dm", v)
}

// ToMi returns the megabytes unit.
func ToMi(v int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%dMi", v)
}

func boolPtrToStr(b *bool) string {
	if b == nil {
		return "false"
	}

	return boolToStr(*b)
}

// Check if string is in a string list.
func in(ll []string, s string) bool {
	for _, l := range ll {
		if l == s {
			return true
		}
	}
	return false
}

// Pad a string up to the given length or truncates if greater than length.
func Pad(s string, width int) string {
	if len(s) == width {
		return s
	}

	if len(s) > width {
		return Truncate(s, width)
	}

	return s + strings.Repeat(" ", width-len(s))
}

// Converts labels string to map
func labelize(labels string) map[string]string {
	ll := strings.Split(labels, ",")
	data := make(map[string]string, len(ll))

	for _, l := range ll {
		tokens := strings.Split(l, "=")
		if len(tokens) == 2 {
			data[tokens[0]] = tokens[1]
		}
	}

	return data
}

func sortLabels(m map[string]string) (keys, vals []string) {
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vals = append(vals, m[k])
	}

	return
}
