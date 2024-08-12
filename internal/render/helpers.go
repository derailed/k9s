// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/vul"
	"github.com/derailed/tview"
	"github.com/mattn/go-runewidth"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// ExtractImages returns a collection of container images.
// !!BOZO!! If this has any legs?? enable scans on other container types.
func ExtractImages(spec *v1.PodSpec) []string {
	ii := make([]string, 0, len(spec.Containers))
	for _, c := range spec.Containers {
		ii = append(ii, c.Image)
	}

	return ii
}

func computeVulScore(m metav1.ObjectMeta, spec *v1.PodSpec) string {
	if vul.ImgScanner == nil || vul.ImgScanner.ShouldExcludes(m) {
		return "0"
	}
	ii := ExtractImages(spec)
	vul.ImgScanner.Enqueue(context.Background(), ii...)

	return vul.ImgScanner.Score(ii...)
}

func runesToNum(rr []rune) int64 {
	var r int64
	var m int64 = 1
	for i := len(rr) - 1; i >= 0; i-- {
		v := int64(rr[i] - '0')
		r += v * m
		m *= 10
	}

	return r
}

// AsThousands prints a number with thousand separator.
func AsThousands(n int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}

// AsStatus returns error as string.
func AsStatus(err error) string {
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

	b := make([]string, 0, len(a))
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

// AsPerc prints a number as percentage with parens.
func AsPerc(p string) string {
	return "(" + p + ")"
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

func naStrings(ss []string) string {
	if len(ss) == 0 {
		return NAValue
	}
	return strings.Join(ss, ",")
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

// ToAge converts time to human duration.
func ToAge(t metav1.Time) string {
	if t.IsZero() {
		return UnknownValue
	}

	return duration.HumanDuration(time.Since(t.Time))
}

func toAgeHuman(s string) string {
	if len(s) == 0 {
		return UnknownValue
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return NAValue
	}

	return duration.HumanDuration(time.Since(t))
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

func mapToStr(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}

	kk := make([]string, 0, len(m))
	for k := range m {
		kk = append(kk, k)
	}
	sort.Strings(kk)

	bb := make([]byte, 0, 100)
	for i, k := range kk {
		bb = append(bb, k+"="+m[k]...)
		if i < len(kk)-1 {
			bb = append(bb, ' ')
		}
	}

	return string(bb)
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

func toMc(v int64) string {
	if v == 0 {
		return ZeroValue
	}
	return strconv.Itoa(int(v))
}

func toMi(v int64) string {
	if v == 0 {
		return ZeroValue
	}
	return strconv.Itoa(int(client.ToMB(v)))
}

func boolPtrToStr(b *bool) string {
	if b == nil {
		return "false"
	}

	return boolToStr(*b)
}

func strPtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// // Check if string is in a string list.
// func in(ll []string, s string) bool {
// 	for _, l := range ll {
// 		if l == s {
// 			return true
// 		}
// 	}
// 	return false
// }

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

// // Converts labels string to map.
// func labelize(labels string) map[string]string {
// 	ll := strings.Split(labels, ",")
// 	data := make(map[string]string, len(ll))

// 	for _, l := range ll {
// 		tokens := strings.Split(l, "=")
// 		if len(tokens) == 2 {
// 			data[tokens[0]] = tokens[1]
// 		}
// 	}

// 	return data
// }

// func sortLabels(m map[string]string) (keys, vals []string) {
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)
// 	for _, k := range keys {
// 		vals = append(vals, m[k])
// 	}

// 	return
// }
