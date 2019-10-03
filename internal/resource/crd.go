package resource

import (
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CustomResourceDefinition tracks a kubernetes resource.
type CustomResourceDefinition struct {
	*Base
	instance *unstructured.Unstructured
}

// NewCustomResourceDefinitionList returns a new resource list.
func NewCustomResourceDefinitionList(c Connection, ns string) List {
	return NewList(
		NotNamespaced,
		"crd",
		NewCustomResourceDefinition(c),
		CRUDAccess|DescribeAccess,
	)
}

// NewCustomResourceDefinition instantiates a new CustomResourceDefinition.
func NewCustomResourceDefinition(c Connection) *CustomResourceDefinition {
	crd := &CustomResourceDefinition{&Base{Connection: c, Resource: k8s.NewCustomResourceDefinition(c)}, nil}
	crd.Factory = crd

	return crd
}

// New builds a new CustomResourceDefinition instance from a k8s resource.
func (r *CustomResourceDefinition) New(i interface{}) Columnar {
	c := NewCustomResourceDefinition(r.Connection)
	switch instance := i.(type) {
	case *unstructured.Unstructured:
		c.instance = instance
	case unstructured.Unstructured:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown CustomResourceDefinition type %#v", i)
	}
	meta := c.instance.Object["metadata"].(map[string]interface{})
	c.path = meta["name"].(string)

	return c
}

// Marshal a resource.
func (r *CustomResourceDefinition) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}

	// BOZO!! Need to figure out apiGroup+Version
	// return r.marshalObject(i.(*unstructured.Unstructured))
	return string(raw), nil
}

// Header return the resource header.
func (*CustomResourceDefinition) Header(ns string) Row {
	return Row{"NAME", "AGE"}
}

// Fields retrieves displayable fields.
func (r *CustomResourceDefinition) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	meta := i.Object["metadata"].(map[string]interface{})
	t, err := time.Parse(time.RFC3339, meta["creationTimestamp"].(string))
	if err != nil {
		log.Error().Msgf("Fields timestamp %v", err)
	}

	return append(ff, meta["name"].(string), toAge(metav1.Time{t}))
}

// ExtFields returns extended fields.
func (r *CustomResourceDefinition) ExtFields(m *TypeMeta) {
	i := r.instance
	spec, ok := i.Object["spec"].(map[string]interface{})
	if !ok {
		return
	}

	if meta, ok := i.Object["metadata"].(map[string]interface{}); ok {
		m.Name = meta["name"].(string)
	}
	m.Group, m.Version = spec["group"].(string), spec["version"].(string)
	m.Namespaced = isNamespaced(spec["scope"].(string))
	names, ok := spec["names"].(map[string]interface{})
	if !ok {
		return
	}
	m.Kind = names["kind"].(string)
	m.Singular, m.Plural = names["singular"].(string), names["plural"].(string)
	if names["shortNames"] != nil {
		for _, s := range names["shortNames"].([]interface{}) {
			m.ShortNames = append(m.ShortNames, s.(string))
		}
	} else {
		m.ShortNames = nil
	}
}

func isNamespaced(scope string) bool {
	return scope == "Namespaced"
}
