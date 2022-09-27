package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

// Generic renders a generic resource to screen.
type Generic struct {
	table *metav1beta1.Table
}

// SetTable sets the tabular resource.
func (g *Generic) SetTable(_ string, t *metav1beta1.Table) {
	g.table = t
}

// Render renders a K8s resource to screen.
func (g *Generic) Render(ctx context.Context, ns string, o interface{}) error {
	row, ok := o.(metav1beta1.TableRow)
	if !ok {
		return fmt.Errorf("expecting a TableRow but got %T", o)
	}

	n, ok := row.Cells[0].(string)
	if !ok {
		return fmt.Errorf("expecting row 0 to be a string but got %T", row.Cells[0])
	}

	root := NewTreeNode("generic", client.FQN(ns, n))
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("expecting TreeNode but got %T", ctx.Value(KeyParent))
	}
	parent.Add(root)

	return nil
}
