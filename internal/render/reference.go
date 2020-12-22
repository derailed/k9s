package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/gdamore/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Reference renders a reference to screen.
type Reference struct{}

// ColorerFunc colors a resource row.
func (Reference) ColorerFunc() ColorerFunc {
	return func(ns string, _ Header, re RowEvent) tcell.Color {
		return tcell.ColorCadetBlue
	}
}

// Header returns a header row.
func (Reference) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "GVR"},
	}
}

// Render renders a K8s resource to screen.
// BOZO!! Pass in a row with pre-alloc fields??
func (Reference) Render(o interface{}, ns string, r *Row) error {
	ref, ok := o.(ReferenceRes)
	if !ok {
		return fmt.Errorf("expected ReferenceRes, but got %T", o)
	}

	r.ID = client.FQN(ref.Namespace, ref.Name)
	r.Fields = append(r.Fields,
		ref.Namespace,
		ref.Name,
		ref.GVR,
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// ReferenceRes represents a reference resource.
type ReferenceRes struct {
	Namespace string
	Name      string
	GVR       string
}

// GetObjectKind returns a schema object.
func (ReferenceRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (a ReferenceRes) DeepCopyObject() runtime.Object {
	return a
}
