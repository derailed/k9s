package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// StatefulSet renders a K8s StatefulSet to screen.
type StatefulSet struct{}

// ColorerFunc colors a resource row.
func (s StatefulSet) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		c := DefaultColorer(ns, re)
		if re.Kind == EventAdd || re.Kind == EventUpdate {
			return c
		}
		if !Happy(ns, re.Row) {
			return ErrColor
		}

		return StdColor
	}
}

// Header returns a header row.
func (StatefulSet) Header(ns string) HeaderRow {
	var h HeaderRow
	if client.IsAllNamespaces(ns) {
		h = append(h, Header{Name: "NAMESPACE"})
	}

	return append(h,
		Header{Name: "NAME"},
		Header{Name: "READY"},
		Header{Name: "SELECTOR", Wide: true},
		Header{Name: "SERVICE"},
		Header{Name: "CONTAINERS", Wide: true},
		Header{Name: "IMAGES", Wide: true},
		Header{Name: "LABELS", Wide: true},
		Header{Name: "VALID", Wide: true},
		Header{Name: "AGE", Decorator: AgeDecorator},
	)
}

// Render renders a K8s resource to screen.
func (s StatefulSet) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected StatefulSet, but got %T", o)
	}
	var sts appsv1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sts)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(sts.ObjectMeta)
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	if client.IsAllNamespaces(ns) {
		r.Fields = append(r.Fields, sts.Namespace)
	}
	r.Fields = append(r.Fields,
		sts.Name,
		strconv.Itoa(int(sts.Status.ReadyReplicas))+"/"+strconv.Itoa(int(sts.Status.Replicas)),
		asSelector(sts.Spec.Selector),
		na(sts.Spec.ServiceName),
		podContainerNames(sts.Spec.Template.Spec, true),
		podImageNames(sts.Spec.Template.Spec, true),
		mapToStr(sts.Labels),
		asStatus(s.diagnose(sts.Status.Replicas, sts.Status.ReadyReplicas)),
		toAge(sts.ObjectMeta.CreationTimestamp),
	)

	return nil
}

func (StatefulSet) diagnose(d, r int32) error {
	if d != r {
		return fmt.Errorf("desiring %d replicas got %d available", d, r)
	}
	return nil
}
