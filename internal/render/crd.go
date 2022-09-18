package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// CustomResourceDefinition renders a K8s CustomResourceDefinition to screen.
type CustomResourceDefinition struct {
	Base
}

// Header returns a header rbw.
func (CustomResourceDefinition) Header(string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "LABELS", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (CustomResourceDefinition) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected CustomResourceDefinition, but got %T", o)
	}

	var crd v1.CustomResourceDefinition
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &crd)
	if err != nil {
		return err
	}

	r.ID = client.FQN(client.ClusterScope, crd.GetName())
	r.Fields = Fields{
		crd.GetName(),
		mapToIfc(crd.GetLabels()),
		toAge(crd.GetCreationTimestamp()),
	}

	return nil
}

func extractMetaField(m map[string]interface{}, field string) string {
	f, ok := m[field]
	if !ok {
		log.Error().Err(fmt.Errorf("failed to extract field from meta %s", field))
		return NAValue
	}

	fs, ok := f.(string)
	if !ok {
		log.Error().Err(fmt.Errorf("failed to extract string from field %s", field))
		return NAValue
	}

	return fs
}
