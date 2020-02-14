package dao

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Healthz struct {
	Resource string
	Count    int64
	Errors   int64
}

type Pulse interface{}
type Pulses []Pulse


func ClusterHealth(ctx context.Context, gvr string) (Pulses, error) {
	var h Healthz

	oo, err := p.List(ctx, "")
	if err != nil {
		return nil, err
	}

	h.Count = len(oo)
	for _, o := range oo {
		var pod v1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return nil, err
		}

		if !happy(pod) {
			h.Errors++
		}
	}
}

func happy(p v1.Pod) bool {

}
