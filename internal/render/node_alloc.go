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

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// NodeAllocHeader defines the header for nodes-alloc view without ROLE, TAINTS, VERSION, and AGE columns.
var nodeAllocHeader = model1.Header{
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "ARCH", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "OS-IMAGE", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "KERNEL", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "INTERNAL-IP", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "EXTERNAL-IP", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "PODS", Attrs: model1.Attrs{Align: tview.AlignRight}},
	model1.HeaderColumn{Name: "CPU", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "CPU/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "CPU/R", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%CPU", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%CPU/R", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "MEM", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "MEM/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "MEM/R", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%MEM", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "%MEM/R", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "GPU/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "GPU/C", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "SH-GPU/A", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "SH-GPU/C", Attrs: model1.Attrs{Align: tview.AlignRight, MX: true}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
}

// NodeAlloc renders a K8s Node for the nodes-alloc view.
type NodeAlloc struct {
	Base
}

// Header returns a header row without ROLE, TAINTS, VERSION, and AGE columns.
func (n NodeAlloc) Header(_ string) model1.Header {
	return n.doHeader(nodeAllocHeader)
}

// Render renders a K8s resource to screen.
func (n NodeAlloc) Render(o any, _ string, row *model1.Row) error {
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

	cols, err := n.specs.realize(nwm.Raw, nodeAllocHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

// defaultRow renders a node row without ROLE, TAINTS, VERSION, and AGE fields.
func (n NodeAlloc) defaultRow(nwm *NodeWithMetrics, r *model1.Row) error {
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

	podCount := strconv.Itoa(nwm.PodCount)
	if pc := nwm.PodCount; pc == -1 {
		podCount = NAValue
	}

	var cpuReq, memReq string
	if nwm.RequestedCPU >= 0 {
		cpuReq = toMc(nwm.RequestedCPU)
	} else {
		cpuReq = NAValue
	}
	if nwm.RequestedMemory >= 0 {
		memReq = toMi(nwm.RequestedMemory)
	} else {
		memReq = NAValue
	}

	var cpuReqPct, memReqPct string
	if nwm.RequestedCPU >= 0 && a.cpu > 0 {
		cpuReqPct = client.ToPercentageStr(nwm.RequestedCPU, a.cpu)
	} else {
		cpuReqPct = NAValue
	}
	if nwm.RequestedMemory >= 0 && a.mem > 0 {
		memReqPct = client.ToPercentageStr(nwm.RequestedMemory, a.mem)
	} else {
		memReqPct = NAValue
	}

	// Fields without ROLE (index 2), TAINTS (index 4), VERSION (index 5), and AGE (index 20)
	r.ID = client.FQN("", no.Name)
	r.Fields = model1.Fields{
		no.Name,                              // NAME
		join(statuses, ","),                  // STATUS
		no.Status.NodeInfo.Architecture,      // ARCH
		no.Status.NodeInfo.OSImage,           // OS-IMAGE
		no.Status.NodeInfo.KernelVersion,     // KERNEL
		iIP,                                  // INTERNAL-IP
		eIP,                                  // EXTERNAL-IP
		podCount,                             // PODS
		toMc(c.cpu),                          // CPU
		toMc(a.cpu),                          // CPU/A
		cpuReq,                               // CPU/R
		client.ToPercentageStr(c.cpu, a.cpu), // %CPU
		cpuReqPct,                            // %CPU/R
		toMi(c.mem),                          // MEM
		toMi(a.mem),                          // MEM/A
		memReq,                               // MEM/R
		client.ToPercentageStr(c.mem, a.mem), // %MEM
		memReqPct,                            // %MEM/R
		toMu(a.gpu),                          // GPU/A
		toMu(c.gpu),                          // GPU/C
		toMu(a.gpuShared),                    // SH-GPU/A
		toMu(c.gpuShared),                    // SH-GPU/C
		mapToStr(no.Labels),                  // LABELS
		AsStatus(n.diagnose(statuses)),       // VALID
	}

	return nil
}

// Healthy checks component health.
func (n NodeAlloc) Healthy(_ context.Context, o any) error {
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

func (NodeAlloc) diagnose(ss []string) error {
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

// ColorerFunc colors a resource row.
func (NodeAlloc) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}
