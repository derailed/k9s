// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/itchyny/gojq"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
)

// ColsSpecs represents a collection of column specification ie NAME:spec|flags.
type ColsSpecs []string

// NewColsSpecs returns a new instance.
func NewColsSpecs(cols ...string) ColsSpecs {
	return ColsSpecs(cols)
}

func (cc ColsSpecs) parseSpecs() (ColumnSpecs, error) {
	specs := make(ColumnSpecs, 0, len(cc))

	for _, c := range cc {
		def, err := parse(c)
		if err != nil {
			return nil, err
		}
		specs = append(specs, ColumnSpec{
			Header: def.toHeaderCol(),
			Spec:   def.spec,
		})
	}

	return specs, nil
}

// RenderedCols tracks a collection of column header and cust column parse expression.
type RenderedCols []RenderedCol

func (rr RenderedCols) hydrateRow(row *model1.Row) {
	ff := make(model1.Fields, 0, len(row.Fields))
	for _, c := range rr {
		ff = append(ff, c.Value)
	}
	row.Fields = ff
}

// HasHeader checks if a given header is present in the collection.
func (rr RenderedCols) HasHeader(n string) bool {
	for _, r := range rr {
		if r.has(n) {
			return true
		}
	}

	return false
}

// RenderedCol represents a column header and a column spec.
type RenderedCol struct {
	Header model1.HeaderColumn
	Value  string
}

// Has checks if the header column match the given name.
func (r RenderedCol) has(n string) bool {
	return r.Header.Name == n
}

// ColumnSpec tracks a header column and an options cust column spec.
type ColumnSpec struct {
	Header model1.HeaderColumn
	Spec   string
}

// ColumnSpecs tracks a collection of column specs.
type ColumnSpecs []ColumnSpec

func (c ColumnSpecs) isEmpty() bool {
	return len(c) == 0
}

// Header builds a new header that is a super set of custom and/or default header.
func (cc ColumnSpecs) Header(rh model1.Header) model1.Header {
	hh := make(model1.Header, 0, len(cc))
	for _, h := range cc {
		hh = append(hh, h.Header)
	}

	for _, h := range rh {
		if idx, ok := hh.IndexOf(h.Name, true); ok {
			hh[idx].Attrs = hh[idx].Merge(h.Attrs)
			continue
		}
		hh = append(hh, h)
	}

	return hh
}

func (cc ColumnSpecs) realize(o runtime.Object, rh model1.Header, row *model1.Row) (RenderedCols, error) {
	parsers := make([]*jsonpath.JSONPath, len(cc))
	for ix := range cc {
		if cc[ix].Spec == "" {
			parsers[ix] = nil
			continue
		}
		parsers[ix] = jsonpath.New(
			fmt.Sprintf("column%d", ix),
		).AllowMissingKeys(true)
		if err := parsers[ix].Parse(cc[ix].Spec); err != nil && !isJQSpec(cc[ix].Spec) {
			slog.Warn("Unable to parse custom column",
				slogs.Name, cc[ix].Header.Name,
				slogs.Error, err,
			)
		}
	}

	vv, err := hydrate(o, cc, parsers, rh, row)
	if err != nil {
		return nil, err
	}
	for _, hc := range rh {
		if vv.HasHeader(hc.Name) {
			continue
		}
		if idx, ok := rh.IndexOf(hc.Name, true); ok {
			rc := RenderedCol{Header: hc, Value: row.Fields[idx]}
			rc.Header.Wide = true
			vv = append(vv, rc)
		}
	}

	return vv, nil
}

func hydrate(o runtime.Object, cc ColumnSpecs, parsers []*jsonpath.JSONPath, rh model1.Header, row *model1.Row) (RenderedCols, error) {
	cols := make(RenderedCols, len(parsers))
	for idx := range parsers {
		parser := parsers[idx]
		if parser == nil {
			ix, ok := rh.IndexOf(cc[idx].Header.Name, true)
			if !ok {
				cols[idx] = RenderedCol{
					Header: cc[idx].Header,
					Value:  NAValue,
				}
				slog.Warn("Unable to find custom column", slogs.Name, cc[idx].Header.Name)
				continue
			}
			var v string
			if ix >= len(row.Fields) {
				v = NAValue
			} else {
				v = row.Fields[ix]
			}
			cols[idx] = RenderedCol{
				Header: rh[ix],
				Value:  v,
			}
			continue
		}

		var (
			vals [][]reflect.Value
			err  error
		)
		if unstructured, ok := o.(runtime.Unstructured); ok {
			if vals, ok := jqParse(cc[idx].Spec, unstructured.UnstructuredContent()); ok {
				cols[idx] = RenderedCol{
					Header: cc[idx].Header,
					Value:  vals,
				}
				continue
			}
			vals, err = parser.FindResults(unstructured.UnstructuredContent())
		} else {
			vals, err = parser.FindResults(reflect.ValueOf(o).Elem().Interface())
		}
		if err != nil {
			return nil, err
		}
		values := make([]string, 0, len(vals))
		if len(vals) == 0 || len(vals[0]) == 0 {
			values = append(values, MissingValue)
		}
		for i := range vals {
			for j := range vals[i] {
				var (
					strVal string
					v      = vals[i][j].Interface()
				)
				switch {
				case cc[idx].Header.MXC:
					switch k := v.(type) {
					case resource.Quantity:
						strVal = toMc(k.MilliValue())
					case string:
						if q, err := resource.ParseQuantity(k); err == nil {
							strVal = toMc(q.MilliValue())
						}
					}
				case cc[idx].Header.MXM:
					switch k := v.(type) {
					case resource.Quantity:
						strVal = toMi(k.MilliValue())
					case string:
						if q, err := resource.ParseQuantity(k); err == nil {
							strVal = toMi(q.MilliValue())
						}
					}
				case cc[idx].Header.Time:
					switch k := v.(type) {
					case string:
						if t, err := time.Parse(time.RFC3339, k); err == nil {
							strVal = ToAge(metav1.Time{Time: t})
						}
					case metav1.Time:
						strVal = ToAge(k)
					}
				}
				if strVal == "" {
					strVal = fmt.Sprintf("%v", v)
				}
				values = append(values, strVal)
			}
		}
		cols[idx] = RenderedCol{
			Header: cc[idx].Header,
			Value:  strings.Join(values, ","),
		}
	}

	return cols, nil
}

func isJQSpec(spec string) bool {
	return len(strings.Split(spec, "|")) > 2
}

func jqParse(spec string, o map[string]any) (string, bool) {
	if !isJQSpec(spec) {
		return "", false
	}

	exp := spec[1 : len(spec)-1]
	jq, err := gojq.Parse(exp)
	if err != nil {
		slog.Warn("Fail to parse JQ expression", slogs.JQExp, exp, slogs.Error, err)
		return "", false
	}

	rr := make([]string, 0, 10)
	iter := jq.Run(o)
	for v, ok := iter.Next(); ok; v, ok = iter.Next() {
		if e, cool := v.(error); cool && e != nil {
			if errors.Is(e, new(gojq.HaltError)) {
				break
			}
			slog.Error("JQ expression evaluation failed. Check your query", slogs.Error, e)
			continue
		}
		rr = append(rr, fmt.Sprintf("%v", v))
	}
	if len(rr) == 0 {
		return "", false
	}

	return strings.Join(rr, ","), true
}
