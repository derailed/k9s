// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tcell/v2"
	"golang.org/x/exp/slog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultNSHeader = model1.Header{
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Namespace renders a K8s Namespace to screen.
type Namespace struct {
	Base
}

// ColorerFunc colors a resource row.
func (Namespace) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)
		if c == model1.ErrColor {
			return c
		}
		if re.Kind == model1.EventUpdate {
			c = model1.StdColor
		}
		if strings.Contains(strings.TrimSpace(re.Row.Fields[0]), "*") {
			c = model1.HighlightColor
		}

		return c
	}
}

// Header returns a header row.
func (n Namespace) Header(_ string) model1.Header {
	return n.doHeader(defaultNSHeader)
}

// Render renders a K8s resource to screen.
func (n Namespace) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := n.defaultRow(raw, row); err != nil {
		return err
	}
	if n.specs.isEmpty() {
		return nil
	}

	cols, err := n.specs.realize(raw, defaultNSHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (n Namespace) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var ns v1.Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ns)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(&ns.ObjectMeta)
	r.Fields = model1.Fields{
		ns.Name,
		string(ns.Status.Phase),
		mapToStr(ns.Labels),
		AsStatus(n.diagnose(ns.Status.Phase)),
		ToAge(ns.GetCreationTimestamp()),
	}

	return nil
}

// Healthy checks component health.
func (n Namespace) Healthy(_ context.Context, o any) error {
	res, ok := o.(*unstructured.Unstructured)
	if !ok {
		slog.Error("Expected *Unstructured, but got", slogs.Type, fmt.Sprintf("%T", o))
		return nil
	}
	var ns v1.Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, &ns)
	if err != nil {
		slog.Error("Failed to convert Unstructured to Namespace", slogs.Type, fmt.Sprintf("%T", o), slog.String("error", err.Error()))
		return nil
	}

	return n.diagnose(ns.Status.Phase)
}

func (Namespace) diagnose(phase v1.NamespacePhase) error {
	if phase != v1.NamespaceActive && phase != v1.NamespaceTerminating {
		return errors.New("namespace not ready")
	}

	return nil
}
