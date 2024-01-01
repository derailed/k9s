package view

import (
	"errors"
	"fmt"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// OwnerExtender adds log actions to a given viewer.
type OwnerExtender struct {
	ResourceViewer
}

// NewOwnerExtender returns a new extender.
func NewOwnerExtender(v ResourceViewer) ResourceViewer {
	o := OwnerExtender{
		ResourceViewer: v,
	}
	o.AddBindKeysFn(o.bindKeys)

	return &o
}

// BindKeys injects new menu actions.
func (o *OwnerExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyO: ui.NewKeyAction("Show Owner", o.ownerCmd(false), true),
	})
}

func (o *OwnerExtender) ownerCmd(prev bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := o.GetTable().GetSelectedItem()
		if path == "" {
			return nil
		}
		if !isResourcePath(path) {
			path = o.GetTable().Path
		}
		o.showOwner(o.GetTable().GVR(), path, prev, evt)

		return nil
	}
}

func (o *OwnerExtender) showOwner(gvr client.GVR, path string, prev bool, evt *tcell.EventKey) {
	var ownerReferences []v1.OwnerReference

	r, err := o.App().factory.Get(gvr.String(), path, true, labels.Everything())
	if err != nil {
		o.App().Flash().Err(err)
		return
	}

	u, ok := r.(*unstructured.Unstructured)
	if !ok {
		o.App().Flash().Err(errors.New("unable to parse resource"))
		return
	}

	ownerReferences, err = o.extractOwnerReference(u)
	if err != nil {
		o.App().Flash().Err(err)
		return
	}

	var owner model.Component
	for _, ownerReference := range ownerReferences {
		ownerPath := u.GetNamespace() + "/" + ownerReference.Name

		switch ownerReference.Kind {
		case "ReplicaSet":
			rs := NewReplicaSet(client.NewGVR(ownerReference.APIVersion + "/replicasets"))
			rs.SetInstance(ownerPath)
			owner = rs
		case "DaemonSet":
			ds := NewDaemonSet(client.NewGVR(ownerReference.APIVersion + "/daemonsets"))
			ds.SetInstance(ownerPath)
			owner = ds
		case "Deployment":
			d := NewDeploy(client.NewGVR(ownerReference.APIVersion + "/deployments"))
			d.SetInstance(ownerPath)
			owner = d
		}
	}

	if owner == nil {
		o.App().Flash().Err(errors.New(fmt.Sprintf("unsupported owner kind")))
		return
	}

	if err := o.App().inject(owner, false); err != nil {
		o.App().Flash().Err(err)
	}

	return
}

// extractOwnerReference extracts the OwnerReferences from an unstructured object
func (o *OwnerExtender) extractOwnerReference(obj *unstructured.Unstructured) ([]v1.OwnerReference, error) {
	ownerRef, found, err := unstructured.NestedSlice(obj.Object, "metadata", "ownerReferences")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	var ownerReferences []v1.OwnerReference
	for _, ref := range ownerRef {
		ownerReferenceMap, ok := ref.(map[string]interface{})
		if !ok {
			return nil, errors.New("could not extract ownerReference")
		}

		ownerReference := v1.OwnerReference{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(ownerReferenceMap, &ownerReference)
		if err != nil {
			return nil, err
		}

		ownerReferences = append(ownerReferences, ownerReference)
	}

	return ownerReferences, nil
}
