// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/vul"
	"github.com/derailed/tview"
	"github.com/mattn/go-runewidth"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

func getTimestampAge(ts any) string {
	if ts == nil {
		return "-"
	}
	var t metav1.Time
	switch v := ts.(type) {
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return "-"
		}
		t = metav1.Time{Time: parsed}
	case metav1.Time:
		t = v
	default:
		return "-"
	}
	return duration.HumanDuration(time.Since(t.Time))
}

func formatGatewayAddresses(addresses any) string {
	if addresses == nil {
		return "-"
	}
	addrSlice, ok := addresses.([]any)
	if !ok || len(addrSlice) == 0 {
		return "-"
	}
	result := make([]string, 0, len(addrSlice))
	for _, addr := range addrSlice {
		addrMap, ok := addr.(map[string]any)
		if !ok {
			continue
		}
		if val, exists := addrMap["value"]; exists {
			if valStr, ok := val.(string); ok && valStr != "" {
				result = append(result, valStr)
			}
		}
	}
	if len(result) == 0 {
		return "-"
	}
	if len(result) <= 3 {
		return strings.Join(result, ", ")
	}
	return strings.Join(result[:3], ", ") + "..."
}

func formatGatewayPorts(ports any) string {
	if ports == nil {
		return "-"
	}
	portSlice, ok := ports.([]any)
	if !ok || len(portSlice) == 0 {
		return "-"
	}
	result := make([]string, 0, len(portSlice))
	for _, port := range portSlice {
		portMap, ok := port.(map[string]any)
		if !ok {
			continue
		}
		var portNum string
		if val, exists := portMap["port"]; exists {
			switch v := val.(type) {
			case int64:
				portNum = strconv.FormatInt(v, 10)
			case float64:
				portNum = strconv.FormatInt(int64(v), 10)
			case string:
				portNum = v
			}
		}
		if val, exists := portMap["protocol"]; exists {
			if protocol, ok := val.(string); ok && protocol != "" {
				if portNum != "" {
					result = append(result, portNum+"/"+protocol)
				}
			}
		}
	}
	if len(result) == 0 {
		return "-"
	}
	if len(result) <= 3 {
		return strings.Join(result, ", ")
	}
	return strings.Join(result[:3], ", ") + "..."
}

func getGatewayReadyStatus(conditions any) string {
	if conditions == nil {
		return "-"
	}
	condSlice, ok := conditions.([]any)
	if !ok || len(condSlice) == 0 {
		return "-"
	}
	for _, cond := range condSlice {
		condMap, ok := cond.(map[string]any)
		if !ok {
			continue
		}
		if typ, exists := condMap["type"]; exists {
			if typeStr, ok := typ.(string); ok && typeStr == "Ready" {
				if status, exists := condMap["status"]; exists {
					if statusStr, ok := status.(string); ok {
						if statusStr == "True" {
							return "✓"
						}
						return "✗"
					}
				}
			}
		}
	}
	return "-"
}

func formatRouteHostnames(hostnames any) string {
	if hostnames == nil {
		return "-"
	}
	hostnameSlice, ok := hostnames.([]any)
	if !ok || len(hostnameSlice) == 0 {
		return "-"
	}
	result := make([]string, 0, len(hostnameSlice))
	for _, hostname := range hostnameSlice {
		if hostnameStr, ok := hostname.(string); ok && hostnameStr != "" {
			result = append(result, hostnameStr)
		}
	}
	if len(result) == 0 {
		return "-"
	}
	if len(result) <= 2 {
		return strings.Join(result, ", ")
	}
	return strings.Join(result[:2], ", ") + "..."
}

func formatRouteServices(rules any) string {
	if rules == nil {
		return "-"
	}
	ruleSlice, ok := rules.([]any)
	if !ok || len(ruleSlice) == 0 {
		return "-"
	}
	services := make(map[string]bool)
	for _, rule := range ruleSlice {
		ruleMap, ok := rule.(map[string]any)
		if !ok {
			continue
		}
		if backendRefs, exists := ruleMap["backendRefs"]; exists {
			refSlice, ok := backendRefs.([]any)
			if !ok {
				continue
			}
			for _, ref := range refSlice {
				refMap, ok := ref.(map[string]any)
				if !ok {
					continue
				}
				if kind, exists := refMap["kind"]; exists {
					if kindStr, ok := kind.(string); ok && kindStr == "Service" {
						if name, exists := refMap["name"]; exists {
							if nameStr, ok := name.(string); ok && nameStr != "" {
								services[nameStr] = true
							}
						}
					}
				}
			}
		}
	}
	if len(services) == 0 {
		return "-"
	}
	serviceNames := make([]string, 0, len(services))
	for service := range services {
		serviceNames = append(serviceNames, service)
	}
	if len(serviceNames) <= 2 {
		return strings.Join(serviceNames, ", ")
	}
	return strings.Join(serviceNames[:2], ", ") + "..."
}

func formatRouteParents(parents any) string {
	if parents == nil {
		return "-"
	}
	parentSlice, ok := parents.([]any)
	if !ok || len(parentSlice) == 0 {
		return "-"
	}
	result := make([]string, 0, len(parentSlice))
	for _, parent := range parentSlice {
		parentMap, ok := parent.(map[string]any)
		if !ok {
			continue
		}
		if parentName, exists := parentMap["parentRef"]; exists {
			if refMap, ok := parentName.(map[string]any); ok {
				if name, exists := refMap["name"]; exists {
					if nameStr, ok := name.(string); ok && nameStr != "" {
						result = append(result, nameStr)
					}
				}
			}
		}
	}
	if len(result) == 0 {
		return "-"
	}
	if len(result) <= 2 {
		return strings.Join(result, ", ")
	}
	return strings.Join(result[:2], ", ") + "..."
}

func getStatusMessage(conditions any, conditionType string) string {
	if conditions == nil {
		return "-"
	}
	condSlice, ok := conditions.([]any)
	if !ok || len(condSlice) == 0 {
		return "-"
	}
	for _, cond := range condSlice {
		condMap, ok := cond.(map[string]any)
		if !ok {
			continue
		}
		if typ, exists := condMap["type"]; exists {
			if typeStr, ok := typ.(string); ok && typeStr == conditionType {
				if status, exists := condMap["status"]; exists {
					if statusStr, ok := status.(string); ok && statusStr == "True" {
						if msg, exists := condMap["message"]; exists {
							if msgStr, ok := msg.(string); ok && msgStr != "" {
								return msgStr
							}
						}
						return "✓"
					}
					if msg, exists := condMap["message"]; exists {
						if msgStr, ok := msg.(string); ok && msgStr != "" {
							return msgStr
						}
					}
					return "✗"
				}
			}
		}
	}
	return "-"
}

func joinStrings(items any) string {
	if items == nil {
		return "-"
	}
	itemSlice, ok := items.([]any)
	if !ok || len(itemSlice) == 0 {
		return "-"
	}
	result := make([]string, 0, len(itemSlice))
	for _, item := range itemSlice {
		if itemStr, ok := item.(string); ok && itemStr != "" {
			result = append(result, itemStr)
		}
	}
	if len(result) == 0 {
		return "-"
	}
	if len(result) <= 3 {
		return strings.Join(result, ", ")
	}
	return strings.Join(result[:3], ", ") + "..."
}

// ExtractImages returns a collection of container images.
// !!BOZO!! If this has any legs?? enable scans on other container types.
func ExtractImages(spec *v1.PodSpec) []string {
	ii := make([]string, 0, len(spec.Containers))
	for i := range spec.Containers {
		ii = append(ii, spec.Containers[i].Image)
	}

	return ii
}

func computeVulScore(ns string, lbls map[string]string, spec *v1.PodSpec) string {
	if vul.ImgScanner == nil || !vul.ImgScanner.IsInitialized() || vul.ImgScanner.ShouldExcludes(ns, lbls) {
		return NAValue
	}
	ii := ExtractImages(spec)
	vul.ImgScanner.Enqueue(context.Background(), ii...)
	sc := vul.ImgScanner.Score(ii...)

	return sc
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
		slog.Error("Selector conversion failed", slogs.Error, err)
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
func blank(ss []string) bool {
	for _, s := range ss {
		if s != "" {
			return false
		}
	}

	return true
}

// Join a slice of strings, skipping blanks.
func join(ss []string, sep string) string {
	switch len(ss) {
	case 0:
		return ""
	case 1:
		return ss[0]
	}

	b := make([]string, 0, len(ss))
	for _, s := range ss {
		if s != "" {
			b = append(b, s)
		}
	}
	if len(b) == 0 {
		return ""
	}

	n := len(sep) * (len(b) - 1)
	for i := range b {
		n += len(ss[i])
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
	return strconv.Itoa(p)
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
	if s == "" {
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
	if s == "" {
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
			bb = append(bb, ',')
		}
	}

	return string(bb)
}

func toMu(v int64) string {
	if v == 0 {
		return NAValue
	}

	return strconv.Itoa(int(v))
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
