package resource

import (
	"log"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
)

// ConfigMap tracks a kubernetes resource.
type ConfigMap struct {
	*Base
	instance *v1.ConfigMap
}

// NewConfigMapList returns a new resource list.
func NewConfigMapList(ns string) List {
	return NewConfigMapListWithArgs(ns, NewConfigMap())
}

// NewConfigMapListWithArgs returns a new resource list.
func NewConfigMapListWithArgs(ns string, res Resource) List {
	return newList(ns, "cm", res, AllVerbsAccess)
}

// NewConfigMap instantiates a new ConfigMap.
func NewConfigMap() *ConfigMap {
	return NewConfigMapWithArgs(k8s.NewConfigMap())
}

// NewConfigMapWithArgs instantiates a new ConfigMap.
func NewConfigMapWithArgs(r k8s.Res) *ConfigMap {
	cm := &ConfigMap{
		Base: &Base{
			caller: r,
		},
	}
	cm.creator = cm
	return cm
}

// NewInstance builds a new ConfigMap instance from a k8s resource.
func (*ConfigMap) NewInstance(i interface{}) Columnar {
	cm := NewConfigMap()
	switch i.(type) {
	case *v1.ConfigMap:
		cm.instance = i.(*v1.ConfigMap)
	case v1.ConfigMap:
		ii := i.(v1.ConfigMap)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *ConfigMap) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	cm := i.(*v1.ConfigMap)
	cm.TypeMeta.APIVersion = "v1"
	cm.TypeMeta.Kind = "ConfigMap"
	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// Header return resource header.
func (*ConfigMap) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh, "NAME", "DATA", "AGE")
}

// Fields retrieves displayable fields.
func (r *ConfigMap) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		strconv.Itoa(len(i.Data)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*ConfigMap) ExtFields() Properties {
	return Properties{}
}
