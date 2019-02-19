package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
)

// DaemonSet tracks a kubernetes resource.
type DaemonSet struct {
	*Base
	instance *extv1beta1.DaemonSet
}

// NewDaemonSetList returns a new resource list.
func NewDaemonSetList(ns string) List {
	return NewDaemonSetListWithArgs(ns, NewDaemonSet())
}

// NewDaemonSetListWithArgs returns a new resource list.
func NewDaemonSetListWithArgs(ns string, res Resource) List {
	return newList(ns, "ds", res, AllVerbsAccess)
}

// NewDaemonSet instantiates a new DaemonSet.
func NewDaemonSet() *DaemonSet {
	return NewDaemonSetWithArgs(k8s.NewDaemonSet())
}

// NewDaemonSetWithArgs instantiates a new DaemonSet.
func NewDaemonSetWithArgs(r k8s.Res) *DaemonSet {
	cm := &DaemonSet{
		Base: &Base{
			caller: r,
		},
	}
	cm.creator = cm
	return cm
}

// NewInstance builds a new DaemonSet instance from a k8s resource.
func (*DaemonSet) NewInstance(i interface{}) Columnar {
	cm := NewDaemonSet()
	switch i.(type) {
	case *extv1beta1.DaemonSet:
		cm.instance = i.(*extv1beta1.DaemonSet)
	case extv1beta1.DaemonSet:
		ii := i.(extv1beta1.DaemonSet)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *DaemonSet) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	ds := i.(*extv1beta1.DaemonSet)
	ds.TypeMeta.APIVersion = "extensions/v1beta1"
	ds.TypeMeta.Kind = "DaemonSet"
	return r.marshalObject(ds)
}

// Header return resource header.
func (*DaemonSet) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	hh = append(hh, "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE")
	hh = append(hh, "AVAILABLE", "NODE_SELECTOR", "AGE")
	return hh
}

// Fields retrieves displayable fields.
func (r *DaemonSet) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}
	return append(ff,
		i.Name,
		strconv.Itoa(int(i.Status.DesiredNumberScheduled)),
		strconv.Itoa(int(i.Status.CurrentNumberScheduled)),
		strconv.Itoa(int(i.Status.NumberReady)),
		strconv.Itoa(int(i.Status.UpdatedNumberScheduled)),
		strconv.Itoa(int(i.Status.NumberAvailable)),
		mapToStr(i.Spec.Template.Spec.NodeSelector),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*DaemonSet) ExtFields() Properties {
	return Properties{}
}
