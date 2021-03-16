package render

import (
	"fmt"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// AppProject renders an ArgoCD App Project to screen.
type AppProject struct{}

// ColorerFunc colors a resource row.
func (AppProject) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (AppProject) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (AppProject) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected AppProject, but got %T", o)
	}
	var app v1alpha1.AppProject
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &app)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(app.ObjectMeta)
	r.Fields = Fields{
		app.Name,
		toAge(app.ObjectMeta.CreationTimestamp),
	}

	return nil
}
