package render

import (
	"fmt"
	"strings"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Application renders an ArgoCD Application to screen.
type Application struct{}

// ColorerFunc colors a resource row.
func (Application) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Application) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "SYNC STATUS"},
		HeaderColumn{Name: "HEALTH STATUS"},
		HeaderColumn{Name: "SYNC POLICY"},
		HeaderColumn{Name: "REVISION"},
		HeaderColumn{Name: "AGE", Time: true, Decorator: AgeDecorator},
	}
}

// Render renders a K8s resource to screen.
func (Application) Render(o interface{}, ns string, r *Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Application, but got %T", o)
	}
	var app v1alpha1.Application
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &app)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(app.ObjectMeta)
	var syncPolicies []string
	syncPolicy := app.Spec.SyncPolicy
	if syncPolicy.Automated.SelfHeal {
		syncPolicies = append(syncPolicies, "selfHeal")
	}
	if syncPolicy.Automated.Prune {
		syncPolicies = append(syncPolicies, "prune")
	}
	r.Fields = Fields{
		app.Name,
		string(app.Status.Sync.Status),
		string(app.Status.Health.Status),
		strings.Join(syncPolicies, ","),
		string(app.Status.Sync.Revision),
		toAge(app.ObjectMeta.CreationTimestamp),
	}

	return nil
}
