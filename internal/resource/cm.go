package resource

import (
	"log"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	v1 "k8s.io/api/core/v1"
)

// ConfigMap tracks a kubernetes resource.
type ConfigMap struct {
	*Base
	instance *v1.ConfigMap
}

// NewConfigMapList returns a new resource list.
func NewConfigMapList(c k8s.Connection, ns string) List {
	return newList(
		ns,
		"cm",
		NewConfigMap(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewConfigMap instantiates a new ConfigMap.
func NewConfigMap(c k8s.Connection) *ConfigMap {
	m := &ConfigMap{&Base{connection: c, resource: k8s.NewConfigMap(c)}, nil}
	m.Factory = m

	return m
}

// New builds a new ConfigMap instance from a k8s resource.
func (r *ConfigMap) New(i interface{}) Columnar {
	cm := NewConfigMap(r.connection)
	switch instance := i.(type) {
	case *v1.ConfigMap:
		cm.instance = instance
	case v1.ConfigMap:
		cm.instance = &instance
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)

	return cm
}

// Marshal resource to yaml.
func (r *ConfigMap) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	cm := i.(*v1.ConfigMap)
	cm.TypeMeta.APIVersion = "v1"
	cm.TypeMeta.Kind = "ConfigMap"

	return r.marshalObject(cm)
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
