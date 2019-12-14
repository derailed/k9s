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

	meta := crd.Object["metadata"].(map[string]interface{})
	t, err := time.Parse(time.RFC3339, meta["creationTimestamp"].(string))
	if err != nil {
		log.Error().Err(err).Msgf("Fields timestamp %v", err)
	}

	r.ID = FQN(ClusterScope, meta["name"].(string))
	r.Fields = Fields{
		meta["name"].(string),
		toAge(metav1.Time{t}),
	}

	return nil
}
