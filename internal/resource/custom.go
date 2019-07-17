package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

// Custom tracks a kubernetes resource.
type Custom struct {
	*Base

	instance             *metav1beta1.TableRow
	group, version, name string
	headers              Row
}

// NewCustomList returns a new resource list.
func NewCustomList(c k8s.Connection, ns, group, version, name string) List {
	if !c.IsNamespaced(name) {
		ns = NotNamespaced
	}
	return NewList(
		ns,
		name,
		NewCustom(c, group, version, name), AllVerbsAccess,
	)
}

// NewCustom instantiates a new Kubernetes Resource.
func NewCustom(c k8s.Connection, group, version, name string) *Custom {
	cr := &Custom{Base: &Base{Connection: c, Resource: k8s.NewResource(c, group, version, name)}}
	cr.Factory = cr
	cr.group, cr.version, cr.name = group, version, name

	return cr
}

// New builds a new Custom instance from a k8s resource.
func (r *Custom) New(i interface{}) Columnar {
	cr := NewCustom(r.Connection, "", "", "")
	switch instance := i.(type) {
	case *metav1beta1.TableRow:
		cr.instance = instance
	case metav1beta1.TableRow:
		cr.instance = &instance
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	var obj map[string]interface{}
	err := json.Unmarshal(cr.instance.Object.Raw, &obj)
	if err != nil {
		log.Error().Err(err)
	}
	meta := obj["metadata"].(map[string]interface{})
	ns := ""
	if n, ok := meta["namespace"]; ok {
		ns = n.(string)
	}
	name := meta["name"].(string)
	cr.path = path.Join(ns, name)
	cr.group, cr.version, cr.name = obj["kind"].(string), obj["apiVersion"].(string), name

	return cr
}

// Marshal resource to yaml.
func (r *Custom) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
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
	ii, err := r.Resource.List(ns)
	if err != nil {
		return nil, err
	}

	if len(ii) == 0 {
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
		cc = append(cc, r.New(rows[i]))
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
		log.Error().Err(err)
		return Row{}
	}

	meta := obj["metadata"].(map[string]interface{})
	rns, ok := meta["namespace"].(string)

	if ns == AllNamespaces {
		if ok {
			ff = append(ff, rns)
		}
	}

	for _, c := range r.instance.Cells {
		ff = append(ff, fmt.Sprintf("%v", c))
	}

	return ff
}
