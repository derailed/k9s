// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// NodeClaim represents a nodeclaim view.
type NodeClaim struct {
	ResourceViewer
}

// NewNodeClaim returns a new nodeclaim view.
func NewNodeClaim(gvr *client.GVR) ResourceViewer {
	n := NodeClaim{
		ResourceViewer: NewBrowser(gvr),
	}
	n.GetTable().SetEnterFn(n.showNode)

	return &n
}

func (n *NodeClaim) showNode(app *App, _ ui.Tabular, gvr *client.GVR, fqn string) {
	nodeName, err := n.getNodeName(fqn)
	if err != nil {
		app.Flash().Err(err)
		return
	}
	if nodeName == "" {
		app.Flash().Warn("NodeClaim does not have an associated node")
		return
	}
	app.gotoResource(client.NodeGVR.String(), nodeName, false, true)
}

func (n *NodeClaim) getNodeName(fqn string) (string, error) {
	res, err := dao.AccessorFor(n.App().factory, n.GVR())
	if err != nil {
		return "", err
	}

	o, err := res.Get(context.Background(), fqn)
	if err != nil {
		return "", err
	}

	return extractNodeName(o), nil
}

// extractNodeName extracts the nodeName from a NodeClaim object.
// NodeClaim is a Karpenter CRD with status.nodeName field.
func extractNodeName(obj interface{}) string {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return ""
	}

	// Get status.nodeName
	status, ok := u.Object["status"]
	if !ok {
		return ""
	}

	statusMap, ok := status.(map[string]interface{})
	if !ok {
		return ""
	}

	nodeName, ok := statusMap["nodeName"]
	if !ok {
		return ""
	}

	name, ok := nodeName.(string)
	if !ok {
		return ""
	}

	return name
}
