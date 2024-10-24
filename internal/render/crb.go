// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ClusterRoleBinding renders a K8s ClusterRoleBinding to screen.
type ClusterRoleBinding struct {
	Base
}

// Header returns a header rbw.
func (ClusterRoleBinding) Header(string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "CLUSTERROLE"},
		model1.HeaderColumn{Name: "SUBJECT-KIND"},
		model1.HeaderColumn{Name: "SUBJECTS"},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (ClusterRoleBinding) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected ClusterRoleBinding, but got %T", o)
	}
	var crb rbacv1.ClusterRoleBinding
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &crb)
	if err != nil {
		return err
	}

	kind, ss := renderSubjects(crb.Subjects)

	r.ID = client.FQN("-", crb.ObjectMeta.Name)
	r.Fields = model1.Fields{
		crb.Name,
		crb.RoleRef.Name,
		kind,
		ss,
		mapToStr(crb.Labels),
		ToAge(crb.GetCreationTimestamp()),
	}

	return nil
}
