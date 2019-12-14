package render

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CustomResourceDefinition renders a K8s CustomResourceDefinition to screen.
type CustomResourceDefinition struct{}

// ColorerFunc colors a resource row.
func (CustomResourceDefinition) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header rbw.
func (CustomResourceDefinition) Header(string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "AGE", Decorator: ageDecorator},
	}
}

// Render renders a K8s resource to screen.
func (CustomResourceDefinition) Render(o interface{}, ns string, r *Row) error {
	crd, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected CustomResourceDefinition, but got %T", o)
	}

	meta := crd.Object["metadata"].(map[string]interface{})
	t, err := time.Parse(time.RFC3339, meta["creationTimestamp"].(string))
	if err != nil {
		log.Error().Err(err).Msgf("Fields timestamp %v", err)
	}

	r.ID = FQN(ClusterWide, meta["name"].(string))
	r.Fields = Fields{
		meta["name"].(string),
		toAge(metav1.Time{t}),
	}

	return nil
}

// BOZO!!
// // TypeMeta represents resource type meta data.
// type TypeMeta struct {
// 	Name       string
// 	Namespaced bool
// 	Group      string
// 	Version    string
// 	Kind       string
// 	Singular   string
// 	Plural     string
// 	ShortNames []string
// }

// func (CustomResourceDefinition) Meta(o interface{}) (TypeMeta, error) {
// 	var m TypeMeta

// 	crd, ok := o.(*unstructured.Unstructured)
// 	if !ok {
// 		return m, fmt.Errorf("Expected CustomResourceDefinition, but got %T", o)
// 	}

// 	spec, ok := crd.Object["spec"].(map[string]interface{})
// 	if !ok {
// 		return m, errors.New("missing crd specs")
// 	}

// 	if meta, ok := crd.Object["metadata"].(map[string]interface{}); ok {
// 		m.Name = meta["name"].(string)
// 	}
// 	m.Group, m.Version = spec["group"].(string), spec["version"].(string)
// 	m.Namespaced = isNamespaced(spec["scope"].(string))
// 	names, ok := spec["names"].(map[string]interface{})
// 	if !ok {
// 		return m, errors.New("missing crd names")
// 	}
// 	m.Kind = names["kind"].(string)
// 	m.Singular, m.Plural = names["singular"].(string), names["plural"].(string)
// 	if names["shortNames"] != nil {
// 		for _, s := range names["shortNames"].([]interface{}) {
// 			m.ShortNames = append(m.ShortNames, s.(string))
// 		}
// 	} else {
// 		m.ShortNames = nil
// 	}
// 	return m, nil
// }

// func isNamespaced(scope string) bool {
// 	return scope == "Namespaced"
// }
