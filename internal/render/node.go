// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	labelNodeRolePrefix = "node-role.kubernetes.io/"
	labelNodeRoleSuffix = "kubernetes.io/role"
)

var defaultNOHeader = model1.Header{
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "ROLE"},
	model1.HeaderColumn{Name: "ARCH", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "TAINTS"},
	model1.HeaderColumn{Name: "VERSION"},
	model1.HeaderColumn{Name: "OS-IMAGE", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "KERNEL", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "INTERNAL-IP", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "EXTERNAL-IP", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "PODS", Attrs: model1.Attrs{Align: tview.AlignRight}},
	model1.HeaderColumn{Name: "CPU", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "CPU/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%CPU", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "MEM", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "MEM/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%MEM", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "GPU/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "GPU/C", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Node renders a K8s Node to screen.
type Node struct {
	Base
}

// Header returns a header row.
func (n Node) Header(_ string) model1.Header {
	return n.doHeader(defaultNOHeader)
}

// Render renders a K8s resource to screen.
func (n Node) Render(o any, _ string, row *model1.Row) error {
	nwm, ok := o.(*NodeWithMetrics)
	if !ok {
		return fmt.Errorf("expected NodeWithMetrics, but got %T", o)
	}
	if err := n.defaultRow(nwm, row); err != nil {
		return err
	}
	if n.specs.isEmpty() {
		return nil
	}

	cols, err := n.specs.realize(nwm.Raw, defaultNOHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

// Render renders a K8s resource to screen.
func (n Node) defaultRow(nwm *NodeWithMetrics, r *model1.Row) error {
	var no v1.Node
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(nwm.Raw.Object, &no)
	if err != nil {
		return err
	}

	iIP, eIP := getIPs(no.Status.Addresses)
	iIP, eIP = missing(iIP), missing(eIP)

	c, a := gatherNodeMX(&no, nwm.MX)

	statuses := make(sort.StringSlice, 10)
	status(no.Status.Conditions, no.Spec.Unschedulable, statuses)
	sort.Sort(statuses)
	roles := make(sort.StringSlice, 10)
	nodeRoles(&no, roles)
	sort.Sort(roles)

	podCount := strconv.Itoa(nwm.PodCount)
	if pc := nwm.PodCount; pc == -1 {
		podCount = NAValue
	}
	r.ID = client.FQN("", no.Name)
	r.Fields = model1.Fields{
		no.Name,
		join(statuses, ","),
		join(roles, ","),
		no.Status.NodeInfo.Architecture,
		strconv.Itoa(len(no.Spec.Taints)),
		no.Status.NodeInfo.KubeletVersion,
		no.Status.NodeInfo.OSImage,
		no.Status.NodeInfo.KernelVersion,
		iIP,
		eIP,
		podCount,
		toMc(c.cpu),
		toMc(a.cpu),
		client.ToPercentageStr(c.cpu, a.cpu),
		toMi(c.mem),
		toMi(a.mem),
		client.ToPercentageStr(c.mem, a.mem),
		toMu(a.gpu),
		toMu(c.gpu),
		mapToStr(no.Labels),
		AsStatus(n.diagnose(statuses)),
		ToAge(no.GetCreationTimestamp()),
	}

	return nil
}

// Healthy checks component health.
func (n Node) Healthy(_ context.Context, o any) error {
	nwm, ok := o.(*NodeWithMetrics)
	if !ok {
		slog.Error("Expected *NodeWithMetrics", slogs.Type, fmt.Sprintf("%T", o))
		return nil
	}
	var no v1.Node
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(nwm.Raw.Object, &no)
	if err != nil {
		slog.Error("Failed to convert unstructured to Node", slogs.Error, err)
		return nil
	}
	ss := make([]string, 10)
	status(no.Status.Conditions, no.Spec.Unschedulable, ss)

	return n.diagnose(ss)
}

func (Node) diagnose(ss []string) error {
	if len(ss) == 0 {
		return nil
	}

	var ready bool
	for _, s := range ss {
		if s == "" {
			continue
		}
		if s == "SchedulingDisabled" {
			return errors.New("node is cordoned")
		}
		if s == "Ready" {
			ready = true
		}
	}

	if !ready {
		return errors.New("node is not ready")
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// NodeWithMetrics represents a node with its associated metrics.
type NodeWithMetrics struct {
	Raw      *unstructured.Unstructured
	MX       *mv1beta1.NodeMetrics
	PodCount int
}

// GetObjectKind returns a schema object.
func (*NodeWithMetrics) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (n *NodeWithMetrics) DeepCopyObject() runtime.Object {
	return n
}

type metric struct {
	cpu, gpu, mem    int64
	lcpu, lgpu, lmem int64
}

func gatherNodeMX(no *v1.Node, mx *mv1beta1.NodeMetrics) (c, a metric) {
	a.cpu = no.Status.Allocatable.Cpu().MilliValue()
	a.mem = no.Status.Allocatable.Memory().Value()
	if mx != nil {
		c.cpu = mx.Usage.Cpu().MilliValue()
		c.mem = mx.Usage.Memory().Value()
	}

	a.gpu = extractGPU(no.Status.Allocatable).Value()
	c.gpu = extractGPU(no.Status.Capacity).Value()

	return
}

func nodeRoles(node *v1.Node, res []string) {
	index := 0
	for k, v := range node.Labels {
		switch {
		case strings.HasPrefix(k, labelNodeRolePrefix):
			if role := strings.TrimPrefix(k, labelNodeRolePrefix); role != "" {
				res[index] = role
				index++
			}
		case strings.HasSuffix(k, labelNodeRoleSuffix) && v != "":
			res[index] = v
			index++
		}
		if index >= len(res) {
			break
		}
	}

	if blank(res) {
		res[index] = MissingValue
	}
}

func getIPs(addrs []v1.NodeAddress) (iIP, eIP string) {
	for _, a := range addrs {
		//nolint:exhaustive
		switch a.Type {
		case v1.NodeExternalIP:
			eIP = a.Address
		case v1.NodeInternalIP:
			iIP = a.Address
		}
	}

	return
}

func status(conds []v1.NodeCondition, exempt bool, res []string) {
	var index int
	conditions := make(map[v1.NodeConditionType]*v1.NodeCondition, len(conds))
	for n := range conds {
		cond := conds[n]
		conditions[cond.Type] = &cond
	}

	validConditions := []v1.NodeConditionType{v1.NodeReady}
	for _, validCondition := range validConditions {
		condition, ok := conditions[validCondition]
		if !ok {
			continue
		}
		neg := ""
		if condition.Status != v1.ConditionTrue {
			neg = "Not"
		}
		res[index] = neg + string(condition.Type)
		index++
	}
	if len(res) == 0 {
		res[index] = "Unknown"
		index++
	}
	if exempt {
		res[index] = "SchedulingDisabled"
	}
}
