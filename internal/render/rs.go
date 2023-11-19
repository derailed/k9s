// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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

// ReplicaSet renders a K8s ReplicaSet to screen.
type ReplicaSet struct {
	Base
}

// ColorerFunc colors a resource row.
func (r ReplicaSet) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (ReplicaSet) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "DESIRED", Align: tview.AlignRight},
		HeaderColumn{Name: "CURRENT", Align: tview.AlignRight},
		HeaderColumn{Name: "READY", Align: tview.AlignRight},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (r ReplicaSet) Render(o interface{}, ns string, row *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected ReplicaSet, but got %T", o)
	}
	var rs appsv1.ReplicaSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rs)
	if err != nil {
		return err
	}

	row.ID = client.MetaFQN(rs.ObjectMeta)
	row.Fields = Fields{
		rs.Namespace,
		rs.Name,
		strconv.Itoa(int(*rs.Spec.Replicas)),
		strconv.Itoa(int(rs.Status.Replicas)),
		strconv.Itoa(int(rs.Status.ReadyReplicas)),
		mapToStr(rs.Labels),
		asStatus(r.diagnose(rs)),
		toAge(rs.GetCreationTimestamp()),
	}

	return nil
}

func (ReplicaSet) diagnose(rs appsv1.ReplicaSet) error {
	if rs.Status.Replicas != rs.Status.ReadyReplicas {
		if rs.Status.Replicas == 0 {
			return fmt.Errorf("did not phase down correctly expecting 0 replicas but got %d", rs.Status.ReadyReplicas)
		}
		return fmt.Errorf("mismatch desired(%d) vs ready(%d)", rs.Status.Replicas, rs.Status.ReadyReplicas)
	}

	return nil
}
