package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

// Custom tracks a kubernetes resource.
type Custom struct {
	*Base
	// instance             *unstructured.Unstructured
	instance             *metav1beta1.TableRow
	group, version, name string
	headers              Row
}

// NewCustomList returns a new resource list.
func NewCustomList(ns, g, v, n string) List {
	return NewCustomListWithArgs(ns, n, NewCustom(g, v, n))
}

// NewCustomListWithArgs returns a new resource list.
func NewCustomListWithArgs(ns, n string, res Resource) List {
	return newList(ns, n, res, AllVerbsAccess)
}

// NewCustom instantiates a new Kubernetes Resource.
func NewCustom(g, v, n string) *Custom {
	return NewCustomWithArgs(k8s.NewResource(g, v, n))
}

// NewCustomWithArgs instantiates a new Custom.
func NewCustomWithArgs(r k8s.Res) *Custom {
	cr := &Custom{
		Base: &Base{
			caller: r,
		},
	}
	cr.creator = cr

	cr.group, cr.version, cr.name = r.(*k8s.Resource).GetInfo()
	return cr
}

// NewInstance builds a new Custom instance from a k8s resource.
func (*Custom) NewInstance(i interface{}) Columnar {
	cr := NewCustom("", "", "")
	switch i.(type) {
	case *metav1beta1.TableRow:
		cr.instance = i.(*metav1beta1.TableRow)
	case metav1beta1.TableRow:
		t := i.(metav1beta1.TableRow)
		cr.instance = &t
	default:
		log.Fatalf("Unknown %#v", i)
	}
	var obj map[string]interface{}
	err := json.Unmarshal(cr.instance.Object.Raw, &obj)
	if err != nil {
		log.Error(err)
	}
	meta := obj["metadata"].(map[string]interface{})
	cr.path = path.Join(meta["namespace"].(string), meta["name"].(string))
	cr.group, cr.version, cr.name = obj["kind"].(string), obj["apiVersion"].(string), meta["name"].(string)
	return cr
}

// Marshal resource to yaml.
func (r *Custom) Marshal(path string) (string, error) {
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
}

// List all resources
func (r *Custom) List(ns string) (Columnars, error) {
	ii, err := r.caller.List(ns)
	if err != nil {
		return nil, err
	}

	if len(ii) != 1 {
		return Columnars{}, errors.New("no resources found")
	}

	table := ii[0].(*metav1beta1.Table)
	r.headers = make(Row, len(table.ColumnDefinitions))
	for i, h := range table.ColumnDefinitions {
		r.headers[i] = h.Name
	}
	rows := table.Rows
	cc := make(Columnars, 0, len(rows))
	for i := 0; i < len(rows); i++ {
		cc = append(cc, r.creator.NewInstance(rows[i]))
	}
	return cc, nil
}

// Header return resource header.
func (r *Custom) Header(ns string) Row {
	hh := make(Row, 0, len(r.headers)+1)
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	for _, h := range r.headers {
		hh = append(hh, strings.ToUpper(h))
	}
	return hh
}

// Fields retrieves displayable fields.
func (r *Custom) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	var obj map[string]interface{}
	err := json.Unmarshal(r.instance.Object.Raw, &obj)
	if err != nil {
		log.Error(err)
		return Row{}
	}

	meta := obj["metadata"].(map[string]interface{})

	if ns == AllNamespaces {
		ff = append(ff, meta["namespace"].(string))
	}
	for _, c := range r.instance.Cells {
		ff = append(ff, fmt.Sprintf("%v", c))
	}

	return ff
}

// ExtFields returns extended fields in relation to headers.
func (*Custom) ExtFields() Properties {
	return Properties{}
}

func getCRDS() map[string]k8s.APIGroup {
	m := map[string]k8s.APIGroup{}
	list := NewCRDList("")
	ll, _ := list.Resource().List("")
	for _, l := range ll {
		ff := l.ExtFields()
		grp := k8s.APIGroup{
			Resource: ff["name"].(string),
			Version:  ff["version"].(string),
			Group:    ff["group"].(string),
			Kind:     ff["kind"].(string),
		}
		if aa, ok := ff["aliases"].([]interface{}); ok {
			if n, ok := ff["plural"].(string); ok {
				grp.Plural = n
			}
			if n, ok := ff["singular"].(string); ok {
				grp.Singular = n
			}
			aliases := make([]string, len(aa))
			for i, a := range aa {
				aliases[i] = a.(string)
			}
			grp.Aliases = aliases
		} else if s, ok := ff["singular"].(string); ok {
			grp.Singular = s
			if p, ok := ff["plural"].(string); ok {
				grp.Plural = p
			}
		} else if s, ok := ff["plural"].(string); ok {
			grp.Plural = s
		}
		m[grp.Kind] = grp
	}
	return m
}
