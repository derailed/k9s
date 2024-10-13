// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Pod represents an xray renderer.
type Pod struct{}

// Render renders an xray node.
func (p *Pod) Render(ctx context.Context, ns string, o interface{}) error {
	pwm, ok := o.(*render.PodWithMetrics)
	if !ok {
		return fmt.Errorf("expected PodWithMetrics, but got %T", o)
	}

	var po v1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(pwm.Raw.Object, &po)
	if err != nil {
		return err
	}

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("no factory found in context")
	}

	node := NewTreeNode("v1/pods", client.FQN(po.Namespace, po.Name))
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	if err := p.containerRefs(ctx, node, po.Namespace, po.Spec); err != nil {
		return err
	}
	p.podVolumeRefs(f, node, po.Namespace, po.Spec.Volumes)
	if err := p.serviceAccountRef(ctx, f, node, po.Namespace, po.Spec); err != nil {
		return err
	}

	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, po.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(node)

	return p.validate(node, po)
}

func (p *Pod) validate(node *TreeNode, po v1.Pod) error {
	var re render.Pod
	phase := re.Phase(&po)
	ss := po.Status.ContainerStatuses
	cr, _, _ := re.Statuses(ss)
	status := OkStatus
	if cr != len(ss) {
		status = ToastStatus
	}
	if phase == "Completed" {
		status = CompletedStatus
	}

	node.Extras[StatusKey] = status
	node.Extras[InfoKey] = strconv.Itoa(cr) + "/" + strconv.Itoa(len(ss))

	return nil
}

func (*Pod) containerRefs(ctx context.Context, parent *TreeNode, ns string, spec v1.PodSpec) error {
	ctx = context.WithValue(ctx, KeyParent, parent)
	var cre Container
	for i := 0; i < len(spec.InitContainers); i++ {
		if err := cre.Render(ctx, ns, render.ContainerRes{Container: &spec.InitContainers[i]}); err != nil {
			return err
		}
	}
	for i := 0; i < len(spec.Containers); i++ {
		if err := cre.Render(ctx, ns, render.ContainerRes{Container: &spec.Containers[i]}); err != nil {
			return err
		}
	}
	for i := 0; i < len(spec.EphemeralContainers); i++ {
		if err := cre.Render(ctx, ns, render.ContainerRes{Container: &spec.Containers[i]}); err != nil {
			return err
		}
	}

	return nil
}

func (*Pod) serviceAccountRef(ctx context.Context, f dao.Factory, parent *TreeNode, ns string, spec v1.PodSpec) error {
	if spec.ServiceAccountName == "" {
		return nil
	}

	id := client.FQN(ns, spec.ServiceAccountName)
	o, err := f.Get("v1/serviceaccounts", id, true, labels.Everything())
	if err != nil {
		return err
	}
	if o == nil {
		addRef(f, parent, "v1/serviceaccounts", id, nil)
		return nil
	}

	var saRE ServiceAccount
	ctx = context.WithValue(ctx, KeyParent, parent)
	ctx = context.WithValue(ctx, KeySAAutomount, spec.AutomountServiceAccountToken)
	return saRE.Render(ctx, ns, o)
}

func (*Pod) podVolumeRefs(f dao.Factory, parent *TreeNode, ns string, vv []v1.Volume) {
	for _, v := range vv {
		sec := v.VolumeSource.Secret
		if sec != nil {
			addRef(f, parent, "v1/secrets", client.FQN(ns, sec.SecretName), sec.Optional)
			continue
		}

		cm := v.VolumeSource.ConfigMap
		if cm != nil {
			addRef(f, parent, "v1/configmaps", client.FQN(ns, cm.LocalObjectReference.Name), cm.Optional)
			continue
		}

		pvc := v.VolumeSource.PersistentVolumeClaim
		if pvc != nil {
			addRef(f, parent, "v1/persistentvolumeclaims", client.FQN(ns, pvc.ClaimName), nil)
		}
	}
}
