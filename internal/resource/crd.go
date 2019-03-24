package resource

import (
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CRD tracks a kubernetes resource.
type CRD struct {
	*Base
	instance *unstructured.Unstructured
}

// NewCRDList returns a new resource list.
func NewCRDList(c k8s.Connection, ns string) List {
	return newList(
		NotNamespaced,
		"crd",
		NewCRD(c),
		CRUDAccess|DescribeAccess,
	)
}

// NewCRD instantiates a new CRD.
func NewCRD(c k8s.Connection) *CRD {
	crd := &CRD{&Base{connection: c, resource: k8s.NewCRD(c)}, nil}
	crd.Factory = crd

	return crd
}

// New builds a new CRD instance from a k8s resource.
func (r *CRD) New(i interface{}) Columnar {
	c := NewCRD(r.connection)
	switch instance := i.(type) {
	case *unstructured.Unstructured:
		c.instance = instance
	case unstructured.Unstructured:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown CRD type %#v", i)
	}
	meta := c.instance.Object["metadata"].(map[string]interface{})
	c.path = meta["name"].(string)

	return c
}

// Marshal a resource.
func (r *CRD) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.resource.Get(ns, n)
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
func (*CRD) Header(ns string) Row {
	return Row{"NAME", "AGE"}
}

// Fields retrieves displayable fields.
func (r *CRD) Fields(ns string) Row {
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
func (r *CRD) ExtFields() Properties {
	var (
		pp = Properties{}
		i  = r.instance
	)

	meta := i.Object["metadata"].(map[string]interface{})

	if spec, ok := i.Object["spec"].(map[string]interface{}); ok {
		pp["name"] = meta["name"]
		pp["group"], pp["version"] = spec["group"], spec["version"]
		names := spec["names"].(map[string]interface{})
		pp["kind"] = names["kind"]
		pp["singular"], pp["plural"] = names["singular"], names["plural"]
		pp["aliases"] = names["shortNames"]
	}

	return pp
}
