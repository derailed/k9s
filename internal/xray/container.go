// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// Container represents an xray renderer.
type Container struct{}

// Render renders an xray node.
func (c *Container) Render(ctx context.Context, ns string, o interface{}) error {
	co, ok := o.(render.ContainerRes)
	if !ok {
		return fmt.Errorf("expected ContainerRes, but got %T", o)
	}

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("no factory found in context")
	}

	root := NewTreeNode("containers", client.FQN(ns, co.Container.Name))
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}
	pns, _ := client.Namespaced(parent.ID)
	c.envRefs(f, root, pns, co.Container)
	parent.Add(root)

	return nil
}

func (c *Container) envRefs(f dao.Factory, parent *TreeNode, ns string, co *v1.Container) {
	for _, e := range co.Env {
		if e.ValueFrom == nil {
			continue
		}
		c.secretRefs(f, parent, ns, e.ValueFrom.SecretKeyRef)
		c.configMapRefs(f, parent, ns, e.ValueFrom.ConfigMapKeyRef)
	}

	for _, e := range co.EnvFrom {
		if e.ConfigMapRef != nil {
			gvr, id := "v1/configmaps", client.FQN(ns, e.ConfigMapRef.Name)
			addRef(f, parent, gvr, id, e.ConfigMapRef.Optional)
		}
		if e.SecretRef != nil {
			gvr, id := "v1/secrets", client.FQN(ns, e.SecretRef.Name)
			addRef(f, parent, gvr, id, e.SecretRef.Optional)
		}
	}
}

func (c *Container) secretRefs(f dao.Factory, parent *TreeNode, ns string, ref *v1.SecretKeySelector) {
	if ref == nil {
		return
	}
	gvr, id := "v1/secrets", client.FQN(ns, ref.LocalObjectReference.Name)
	addRef(f, parent, gvr, id, ref.Optional)
}

func (c *Container) configMapRefs(f dao.Factory, parent *TreeNode, ns string, ref *v1.ConfigMapKeySelector) {
	if ref == nil {
		return
	}
	gvr, id := "v1/configmaps", client.FQN(ns, ref.LocalObjectReference.Name)
	addRef(f, parent, gvr, id, ref.Optional)
}

// ----------------------------------------------------------------------------
// Helpers...

func addRef(f dao.Factory, parent *TreeNode, gvr, id string, optional *bool) {
	if parent.Find(gvr, id) == nil {
		n := NewTreeNode(gvr, id)
		validate(f, n, optional)
		parent.Add(n)
	}
}

func validate(f dao.Factory, n *TreeNode, optional *bool) {
	res, err := f.Get(n.GVR, n.ID, true, labels.Everything())
	if err != nil || res == nil {
		if optional == nil || !*optional {
			log.Warn().Err(err).Msgf("Missing ref %q::%q", n.GVR, n.ID)
			n.Extras[StatusKey] = MissingRefStatus
		}
		return
	}
	n.Extras[StatusKey] = OkStatus
}
