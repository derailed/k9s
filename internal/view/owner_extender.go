package view

import (
	"errors"
	"fmt"
	"github.com/derailed/k9s/internal/client"
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

func (o *OwnerExtender) showOwner(gvr client.GVR, path string, _ bool, _ *tcell.EventKey) {
	var ownerRefs []v1.OwnerReference

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

	ownerRefs, err = extractOwnerRefs(u)
	if err != nil {
		o.App().Flash().Err(err)
		return
	}

	var owner ResourceViewer
	for _, ownerRef := range ownerRefs {
		gvrString, newViewerFunc, err := getKindInfo(ownerRef)
		if err != nil {
			o.App().Flash().Err(err)
			return
		}

		owner = newViewerFunc(client.NewGVR(gvrString))

		ownerPath := u.GetNamespace() + "/" + ownerRef.Name
		owner.SetInstance(ownerPath)
	}

	if err := o.App().inject(owner, false); err != nil {
		o.App().Flash().Err(err)
	}

	return
}

// extractOwnerRefs extracts the OwnerReferences from an unstructured object
func extractOwnerRefs(obj *unstructured.Unstructured) ([]v1.OwnerReference, error) {
	ownerRefInterfaces, found, err := unstructured.NestedSlice(obj.Object, "metadata", "ownerReferences")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	var ownerRefs []v1.OwnerReference
	for _, refInterface := range ownerRefInterfaces {
		refMap, ok := refInterface.(map[string]interface{})
		if !ok {
			return nil, errors.New("could not extract ownerReference")
		}

		var ref v1.OwnerReference
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(refMap, &ref)
		if err != nil {
			return nil, err
		}

		ownerRefs = append(ownerRefs, ref)
	}

	return ownerRefs, nil
}

func getKindInfo(ownerRef v1.OwnerReference) (string, func(client.GVR) ResourceViewer, error) {
	var (
		gvrStrings = map[string]string{
			"ReplicaSet": "apps/v1/replicasets",
			"DaemonSet":  "apps/v1/daemonsets",
			"Deployment": "apps/v1/deployments",
			"Jobs":       "apps/v1/jobs",
			"CronJobs":   "apps/v1/cronjobs",
		}
		newViewerFuncs = map[string]func(client.GVR) ResourceViewer{
			"ReplicaSet": NewReplicaSet,
			"DaemonSet":  NewDaemonSet,
			"Deployment": NewDeploy,
			"Jobs":       NewJob,
			"CronJobs":   NewCronJob,
		}
	)

	if ownerRef.APIVersion != "apps/v1" {
		return "", nil, errors.New(fmt.Sprintf("unsupported ownerReference API version: %s", ownerRef.APIVersion))
	}

	gvrString, found := gvrStrings[ownerRef.Kind]
	if !found {
		return "", nil, errors.New(fmt.Sprintf("unsupported ownerReference kind: %s", ownerRef.Kind))
	}

	newViewerFunc, found := newViewerFuncs[ownerRef.Kind]
	if !found {
		return "", nil, errors.New(fmt.Sprintf("unsupported ownerReference kind: %s", ownerRef.Kind))
	}

	return gvrString, newViewerFunc, nil
}
