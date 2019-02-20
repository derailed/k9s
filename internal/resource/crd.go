package resource

import (
	"time"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
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
func NewCRDList(ns string) List {
	return NewCRDListWithArgs(ns, NewCRD())
}

// NewCRDListWithArgs returns a new resource list.
func NewCRDListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "crd", res, CRUDAccess|DescribeAccess)
}

// NewCRD instantiates a new CRD.
func NewCRD() *CRD {
	return NewCRDWithArgs(k8s.NewCRD())
}

// NewCRDWithArgs instantiates a new Context.
func NewCRDWithArgs(r k8s.Res) *CRD {
	ctx := &CRD{
		Base: &Base{
			caller: r,
		},
	}
	ctx.creator = ctx
	return ctx
}

// NewInstance builds a new Context instance from a k8s resource.
func (r *CRD) NewInstance(i interface{}) Columnar {
	c := NewCRD()
	switch i.(type) {
	case *unstructured.Unstructured:
		c.instance = i.(*unstructured.Unstructured)
	case unstructured.Unstructured:
		ii := i.(unstructured.Unstructured)
		c.instance = &ii
	default:
		log.Fatalf("unknown context type %#v", i)
	}
	meta := c.instance.Object["metadata"].(map[string]interface{})
	c.path = meta["name"].(string)

	return c
}

// Marshal a resource.
func (r *CRD) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(raw), nil
	// BOZO!! Need to figure out apiGroup+Version
	// return r.marshalObject(i.(*unstructured.Unstructured))
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
		log.Error("Fields timestamp", err)
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
