package xray

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

// Generic renders a generic resource to screen.
type Generic struct {
	table *metav1beta1.Table
}

// SetTable sets the tabular resource.
func (g *Generic) SetTable(t *metav1beta1.Table) {
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
	parent := ctx.Value(KeyParent).(*TreeNode)
	parent.Add(root)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func resourceNS(raw []byte) (bool, string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		return false, "", err
	}

	meta, ok := obj["metadata"].(map[string]interface{})
	if !ok {
		return false, "", errors.New("no metadata found on generic resource")
	}

	ns, ok := meta["namespace"]
	if !ok {
		return true, "", nil
	}

	nns, ok := ns.(string)
	if !ok {
		return false, "", fmt.Errorf("expecting namespace string type but got %T", ns)
	}
	return false, nns, nil
}
