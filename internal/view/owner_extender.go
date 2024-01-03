package view

import (
	"errors"
	"fmt"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const selectOwnerDialogKey = "owner"

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
		ui.KeyO: ui.NewKeyAction("Show Owner", o.ownerCmd(), true),
	})
}

func (o *OwnerExtender) ownerCmd() func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := o.GetTable().GetSelectedItem()
		if path == "" {
			return nil
		}
		if !isResourcePath(path) {
			path = o.GetTable().Path
		}

		err := o.showOwner(o.GetTable().GVR(), path)
		if err != nil {
			o.App().Flash().Err(err)
		}

		return nil
	}
}

func (o *OwnerExtender) showOwner(gvr client.GVR, path string) error {
	var ownerRefs []v1.OwnerReference

	r, err := o.App().factory.Get(gvr.String(), path, true, labels.Everything())
	if err != nil {
		return err
	}

	u, ok := r.(*unstructured.Unstructured)
	if !ok {
		return errors.New("unable to parse resource")
	}

	ownerRefs, err = extractOwnerRefs(u)
	if err != nil {
		return err
	}

	if len(ownerRefs) == 0 {
		return errors.New("resource does not have an owner")
	}

	namespace := u.GetNamespace()

	if len(ownerRefs) == 1 {
		return o.goToOwner(ownerRefs[0], namespace)
	}

	return o.showSelectOwnerDialog(ownerRefs, namespace)
}

func (o *OwnerExtender) goToOwner(ownerRef v1.OwnerReference, namespace string) error {
	var owner ResourceViewer

	gvrString, newViewerFunc, err := getKindInfo(ownerRef)
	if err != nil {
		return err
	}

	owner = newViewerFunc(client.NewGVR(gvrString))

	ownerPath := namespace + "/" + ownerRef.Name
	owner.SetInstance(ownerPath)

	if err := o.App().inject(owner, false); err != nil {
		return err
	}

	return nil
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

func (o *OwnerExtender) showSelectOwnerDialog(refs []v1.OwnerReference, namespace string) error {
	form, err := o.makeSelectOwnerForm(refs, namespace)
	if err != nil {
		return err
	}
	modal := tview.NewModalForm("<Owner>", form)
	msg := "Select owner"
	modal.SetText(msg)
	modal.SetDoneFunc(func(int, string) {
		o.dismissDialog()
	})
	o.App().Content.AddPage(selectOwnerDialogKey, modal, false, false)
	o.App().Content.ShowPage(selectOwnerDialogKey)

	return nil
}

func (o *OwnerExtender) makeSelectOwnerForm(refs []v1.OwnerReference, namespace string) (*tview.Form, error) {
	f := o.makeStyledForm()

	var ownerLabels []string
	for _, ref := range refs {
		ownerLabels = append(ownerLabels, fmt.Sprintf("<%s> %s", ref.Kind, ref.Name))
	}

	var selectedRef v1.OwnerReference

	f.AddDropDown("Owner:", ownerLabels, 0, func(option string, optionIndex int) {
		selectedRef = refs[optionIndex]
		return
	})

	f.AddButton("OK", func() {
		defer o.dismissDialog()
		err := o.goToOwner(selectedRef, namespace)
		if err != nil {
			o.App().Flash().Err(err)
		}
	})

	f.AddButton("Cancel", func() {
		o.dismissDialog()
	})

	return f, nil
}

func (o *OwnerExtender) makeStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	return f
}

func (o *OwnerExtender) dismissDialog() {
	o.App().Content.RemovePage(selectOwnerDialogKey)
}
