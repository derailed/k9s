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
	ns, n := namespaced(path)
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
func (r *CustomResourceDefinition) ExtFields() Properties {
	var (
		pp = Properties{}
		i  = r.instance
	)

	if spec, ok := i.Object["spec"].(map[string]interface{}); ok {
		if meta, ok := i.Object["metadata"].(map[string]interface{}); ok {
			pp["name"] = meta["name"]
		}
		pp["group"], pp["version"] = spec["group"], spec["version"]

		if names, ok := spec["names"].(map[string]interface{}); ok {
			pp["kind"] = names["kind"]
			pp["singular"], pp["plural"] = names["singular"], names["plural"]
			pp["aliases"] = names["shortNames"]
		}
	}

	return pp
}
