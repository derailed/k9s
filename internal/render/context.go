package render

import (
	"fmt"

	api "k8s.io/client-go/tools/clientcmd/api"
)

// Context renders a K8s ConfigMap to screen.
type Context struct{}

// ColorerFunc colors a resource row.
func (Context) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Header returns a header row.
func (Context) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "CLUSTER"},
		Header{Name: "AUTHINFO"},
		Header{Name: "NAMESPACE"},
	}
}

// Render renders a K8s resource to screen.
func (Context) Render(o interface{}, _ string, r *Row) error {
	i, ok := o.(*api.Context)
	if !ok {
		return fmt.Errorf("Expected api.Context, but got %T", o)
	}

	r.Fields[0] = r.ID
	r.Fields[1] = i.Cluster
	r.Fields[2] = i.AuthInfo
	r.Fields[3] = i.Namespace

	return nil
}
