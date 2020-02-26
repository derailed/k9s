package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Deployment renders a K8s Deployment to screen.
type Deployment struct{}

// ColorerFunc colors a resource row.
func (d Deployment) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Deployment) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "READY"},
		HeaderColumn{Name: "UP-TO-DATE", Align: tview.AlignRight},
		HeaderColumn{Name: "AVAILABLE", Align: tview.AlignRight},
		HeaderColumn{Name: "READY", Align: tview.AlignRight},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (d Deployment) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Deployment, but got %T", o)
	}

	var dp appsv1.Deployment
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &dp)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(dp.ObjectMeta)
	r.Fields = Fields{
		dp.Namespace,
		dp.Name,
		strconv.Itoa(int(dp.Status.AvailableReplicas)) + "/" + strconv.Itoa(int(dp.Status.Replicas)),
		strconv.Itoa(int(dp.Status.UpdatedReplicas)),
		strconv.Itoa(int(dp.Status.AvailableReplicas)),
		strconv.Itoa(int(dp.Status.ReadyReplicas)),
		mapToStr(dp.Labels),
		asStatus(d.diagnose(dp.Status.Replicas, dp.Status.AvailableReplicas)),
		toAge(dp.ObjectMeta.CreationTimestamp),
	}

	return nil
}

func (Deployment) diagnose(desired, avail int32) error {
	if desired != avail {
		return fmt.Errorf("desiring %d replicas got %d available", desired, avail)
	}
	return nil
}
