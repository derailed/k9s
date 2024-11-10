// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/go-errors/errors"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// OwnerExtender adds owner actions to a given viewer.
type OwnerExtender struct {
	ResourceViewer
}

// NewOwnerExtender returns a new extender.
func NewOwnerExtender(r ResourceViewer) ResourceViewer {
	v := &OwnerExtender{ResourceViewer: r}
	v.AddBindKeysFn(v.bindKeys)

	return v
}

func (v *OwnerExtender) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyShiftJ, ui.NewKeyAction("Jump Owner", v.ownerCmd, true))
}

func (v *OwnerExtender) ownerCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := v.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	if err := v.findOwnerFor(path); err != nil {
		log.Warn().Msgf("Unable to jump to the owner of resource %q: %s", path, err)
		v.App().Flash().Warnf("Unable to jump owner: %s", err)
	}
	return nil
}

func (v *OwnerExtender) findOwnerFor(path string) error {
	res, err := dao.AccessorFor(v.App().factory, v.GVR())
	if err != nil {
		return err
	}

	o, err := res.Get(v.defaultCtx(), path)
	if err != nil {
		return err
	}

	u, ok := v.asUnstructuredObject(o)
	if !ok {
		return errors.Errorf("unsupported object type: %t", o)
	}

	ns, _ := client.Namespaced(path)
	ownerReferences := u.GetOwnerReferences()
	if len(ownerReferences) == 1 {
		return v.jumpOwner(ns, ownerReferences[0])
	} else if len(ownerReferences) > 1 {
		owners := make([]string, 0, len(ownerReferences))
		for idx, ownerRef := range ownerReferences {
			owners = append(owners, fmt.Sprintf("%d: %s", idx, ownerRef.Kind))
		}

		dialog.ShowSelection(v.App().Styles.Dialog(), v.App().Content.Pages, "Jump To", owners, func(index int) {
			if index >= 0 {
				err = v.jumpOwner(ns, ownerReferences[index])
			}
		})
		return err
	}

	return errors.Errorf("no owner found")
}

func (v *OwnerExtender) jumpOwner(ns string, owner metav1.OwnerReference) error {
	gv, err := schema.ParseGroupVersion(owner.APIVersion)
	if err != nil {
		return err
	}

	gvr, namespaced, found := dao.MetaAccess.GVK2GVR(gv, owner.Kind)
	if !found {
		return errors.Errorf("unsupported GVK: %s/%s", owner.APIVersion, owner.Kind)
	}

	var ownerFQN string
	if namespaced {
		ownerFQN = client.FQN(ns, owner.Name)
	} else {
		ownerFQN = owner.Name
	}

	v.App().gotoResource(gvr.String(), ownerFQN, false)
	return nil
}

func (v *OwnerExtender) defaultCtx() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, v.App().factory)
}

func (v *OwnerExtender) asUnstructuredObject(o runtime.Object) (*unstructured.Unstructured, bool) {
	switch v := o.(type) {
	case *unstructured.Unstructured:
		return v, true
	case *render.PodWithMetrics:
		return v.Raw, true
	default:
		return nil, false
	}
}
