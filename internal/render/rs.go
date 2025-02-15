// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
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
func (r ReplicaSet) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Header returns a header row.
func (r ReplicaSet) Header(_ string) model1.Header {
	return r.doHeader(r.defaultHeader())
}

func (ReplicaSet) defaultHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "VS", Attrs: model1.Attrs{VS: true}},
		model1.HeaderColumn{Name: "DESIRED", Attrs: model1.Attrs{Align: tview.AlignRight}},
		model1.HeaderColumn{Name: "CURRENT", Attrs: model1.Attrs{Align: tview.AlignRight}},
		model1.HeaderColumn{Name: "READY", Attrs: model1.Attrs{Align: tview.AlignRight}},
		model1.HeaderColumn{Name: "CONTAINERS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "IMAGES", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "SELECTOR", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
}

// Render renders a K8s resource to screen.
func (r ReplicaSet) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected ReplicaSet, but got %T", o)
	}

	if err := r.defaultRow(raw, row); err != nil {
		return err
	}
	if r.specs.isEmpty() {
		return nil
	}

	cols, err := r.specs.realize(raw, r.defaultHeader(), row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (r ReplicaSet) defaultRow(raw *unstructured.Unstructured, row *model1.Row) error {
	var rs appsv1.ReplicaSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rs)
	if err != nil {
		return err
	}

	var (
		cc        = rs.Spec.Template.Spec.Containers
		cos, imgs = make([]string, 0, len(cc)), make([]string, 0, len(cc))
	)
	for _, co := range cc {
		cos, imgs = append(cos, co.Name), append(imgs, co.Image)
	}

	row.ID = client.MetaFQN(rs.ObjectMeta)
	row.Fields = model1.Fields{
		rs.Namespace,
		rs.Name,
		computeVulScore(rs.ObjectMeta, &rs.Spec.Template.Spec),
		strconv.Itoa(int(*rs.Spec.Replicas)),
		strconv.Itoa(int(rs.Status.Replicas)),
		strconv.Itoa(int(rs.Status.ReadyReplicas)),
		strings.Join(cos, ","),
		strings.Join(imgs, ","),
		mapToStr(rs.Labels),
		AsStatus(r.diagnose(rs)),
		ToAge(rs.GetCreationTimestamp()),
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
