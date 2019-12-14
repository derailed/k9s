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
		Header{Name: "AGE", Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (CustomResourceDefinition) Render(o interface{}, ns string, r *Row) error {
	crd, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected CustomResourceDefinition, but got %T", o)
	}

	meta, ok := crd.Object["metadata"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("expecting an interface map but got %T", crd.Object["metadata"])
	}
	t, err := time.Parse(time.RFC3339, extractMetaField(meta, "creationTimestamp"))
	if err != nil {
		log.Error().Err(err).Msgf("Fields timestamp %v", err)
	}

	r.ID = FQN(ClusterScope, extractMetaField(meta, "name"))
	r.Fields = Fields{
		extractMetaField(meta, "name"),
		toAge(metav1.Time{Time: t}),
	}

	return nil
}

func extractMetaField(m map[string]interface{}, field string) string {
	f, ok := m[field]
	if !ok {
		log.Error().Err(fmt.Errorf("failed to extract field from meta %s", field))
		return "n/a"
	}

	fs, ok := f.(string)
	if !ok {
		log.Error().Err(fmt.Errorf("failed to extract string from field %s", field))
		return "n/a"
	}

	return fs
}
