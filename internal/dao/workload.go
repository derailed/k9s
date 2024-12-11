// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	DegradedStatus = "DEGRADED"
	NotAvailable   = "n/a"
)

var (
	SaGVR  = client.NewGVR("v1/serviceaccounts")
	PvcGVR = client.NewGVR("v1/persistentvolumeclaims")
	PcGVR  = client.NewGVR("scheduling.k8s.io/v1/priorityclasses")
	CmGVR  = client.NewGVR("v1/configmaps")
	SecGVR = client.NewGVR("v1/secrets")
	PodGVR = client.NewGVR("v1/pods")
	SvcGVR = client.NewGVR("v1/services")
	DsGVR  = client.NewGVR("apps/v1/daemonsets")
	StsGVR = client.NewGVR("apps/v1/statefulSets")
	DpGVR  = client.NewGVR("apps/v1/deployments")
	RsGVR  = client.NewGVR("apps/v1/replicasets")
)

// Workload tracks a select set of resources in a given namespace.
type Workload struct {
	Table
}

func (w *Workload) Delete(ctx context.Context, path string, propagation *metav1.DeletionPropagation, grace Grace) error {
	gvr, _ := ctx.Value(internal.KeyGVR).(client.GVR)
	ns, n := client.Namespaced(path)
	auth, err := w.Client().CanI(ns, gvr.String(), n, []string{client.DeleteVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to delete %s", path)
	}

	var gracePeriod *int64
	if grace != DefaultGrace {
		gracePeriod = (*int64)(&grace)
	}
	opts := metav1.DeleteOptions{
		PropagationPolicy:  propagation,
		GracePeriodSeconds: gracePeriod,
	}

	ctx, cancel := context.WithTimeout(ctx, w.Client().Config().CallTimeout())
	defer cancel()

	d, err := w.Client().DynDial()
	if err != nil {
		return err
	}
	dial := d.Resource(gvr.GVR())
	if client.IsClusterScoped(ns) {
		return dial.Delete(ctx, n, opts)
	}

	return dial.Namespace(ns).Delete(ctx, n, opts)
}

// List fetch workloads.
func (a *Workload) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	oo := make([]runtime.Object, 0, 100)

	workloadGVRs, _ := ctx.Value(internal.KeyWorkloadGVRs).([]config.WorkloadGVR)
	for i, wkgvr := range workloadGVRs {
		// Apply default values
		workloadGVRs[i].ApplyDefault()

		table, err := a.fetch(ctx, workloadGVRs[i].GetGVR(), ns)
		if err != nil {
			log.Warn().Msgf("could not fetch gvr %s: %q", workloadGVRs[i].Name, err)
			continue
		}

		var (
			ns string
			ts metav1.Time
		)
		for _, r := range table.Rows {
			if obj := r.Object.Object; obj != nil {
				if m, err := meta.Accessor(obj); err == nil {
					ns = m.GetNamespace()
					ts = m.GetCreationTimestamp()
				}
			} else {
				var m metav1.PartialObjectMetadata
				if err := json.Unmarshal(r.Object.Raw, &m); err == nil {
					ns = m.GetNamespace()
					ts = m.CreationTimestamp
				}
			}

			oo = append(oo, &render.WorkloadRes{Row: metav1.TableRow{Cells: []interface{}{
				workloadGVRs[i].GetGVR().String(),
				ns,
				r.Cells[indexOf("Name", table.ColumnDefinitions)],
				a.getStatus(wkgvr, table.ColumnDefinitions, r.Cells),
				a.getReadiness(wkgvr, table.ColumnDefinitions, r.Cells),
				a.getValidity(wkgvr, table.ColumnDefinitions, r.Cells),
				ts,
			}}})
		}
	}

	return oo, nil
}

func (a *Workload) fetch(ctx context.Context, gvr client.GVR, ns string) (*metav1.Table, error) {
	a.Table.gvr = gvr
	oo, err := a.Table.List(ctx, ns)
	if err != nil {
		return nil, err
	}
	if len(oo) == 0 {
		return nil, fmt.Errorf("no table found for gvr: %s", gvr)
	}
	tt, ok := oo[0].(*metav1.Table)
	if !ok {
		return nil, errors.New("not a metav1.Table")
	}

	return tt, nil
}

// getStatus will retrieve the status of the resource depending of it's configuration
func (wk *Workload) getStatus(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	status := NotAvailable

	if wkgvr.Status == nil || wkgvr.Status.NA {
		return status
	}

	if statusIndex := indexOf(string(wkgvr.Status.CellName), cd); statusIndex != -1 {
		status = valueToString(cells[statusIndex])

	}

	return status
}

// getReadiness will retrieve the readiness of the resource depending of it's configuration
func (wk *Workload) getReadiness(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	ready := NotAvailable

	if wkgvr.Readiness == nil || wkgvr.Readiness.NA {
		return ready
	}

	if readyIndex := indexOf(string(wkgvr.Readiness.CellName), cd); readyIndex != -1 {
		ready = valueToString(cells[readyIndex])
	}

	if extrReadyIndex := indexOf(string(wkgvr.Readiness.CellExtraName), cd); extrReadyIndex != -1 {
		ready = fmt.Sprintf("%s/%s", ready, valueToString(cells[extrReadyIndex]))
	}

	return ready
}

// getValidity will retrieve the validity of the resource depending of it's configuration
func (wk *Workload) getValidity(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	if wkgvr.Validity == nil || wkgvr.Validity.NA {
		return ""
	}

	if validity := getMatchesValidity(wkgvr, cd, cells); validity == DegradedStatus {
		return DegradedStatus
	}

	if validity := getReplicasValidity(wkgvr, cd, cells); validity == DegradedStatus {
		return DegradedStatus
	}

	return ""
}

func getMatchesValidity(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	for _, m := range wkgvr.Validity.Matchs {
		v := ""
		if matchCellNameIndex := indexOf(string(m.CellName), cd); matchCellNameIndex != -1 {
			v = valueToString(cells[matchCellNameIndex])
		}

		if v != m.Value {
			return DegradedStatus
		}

	}

	return ""
}

func getReplicasValidity(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	if getReplicasGrouped(wkgvr, cd, cells) == DegradedStatus {
		return DegradedStatus
	}

	if getReplicasSeparated(wkgvr, cd, cells) == DegradedStatus {
		return DegradedStatus
	}

	return ""
}

func getReplicasGrouped(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	if wkgvr.Validity.Replicas.CellAllName == "" {
		return ""
	}

	allCellNameIndex := indexOf(string(wkgvr.Validity.Replicas.CellAllName), cd)
	if allCellNameIndex < 0 {
		return ""
	}

	if !isReady(valueToString(cells[allCellNameIndex])) {
		return DegradedStatus
	}

	return ""
}

func getReplicasSeparated(wkgvr config.WorkloadGVR, cd []metav1.TableColumnDefinition, cells []interface{}) string {
	if wkgvr.Validity.Replicas.CellCurrentName == "" || wkgvr.Validity.Replicas.CellDesiredName == "" {
		return ""
	}

	currentIndex := indexOf(string(wkgvr.Validity.Replicas.CellCurrentName), cd)
	desiredIndex := indexOf(string(wkgvr.Validity.Replicas.CellDesiredName), cd)

	if currentIndex < 0 || desiredIndex < 0 {
		return ""
	}

	if !isReady(fmt.Sprintf("%s/%s", valueToString(cells[desiredIndex]), valueToString(cells[currentIndex]))) {
		return DegradedStatus
	}

	return ""
}

func valueToString(v interface{}) string {
	if sv, ok := v.(string); ok {
		return sv
	}

	if iv, ok := v.(int64); ok {
		return strconv.Itoa(int(iv))
	}

	return ""
}

func isReady(s string) bool {
	tt := strings.Split(s, "/")
	if len(tt) != 2 {
		return false
	}
	r, err := strconv.Atoi(tt[0])
	if err != nil {
		log.Error().Msgf("invalid ready count: %q", tt[0])
		return false
	}
	c, err := strconv.Atoi(tt[1])
	if err != nil {
		log.Error().Msgf("invalid expected count: %q", tt[1])
		return false
	}

	if c == 0 {
		return true
	}
	return r == c
}

func indexOf(n string, defs []metav1.TableColumnDefinition) int {
	if n == "" {
		return -1
	}

	for i, d := range defs {
		if d.Name == n {
			return i
		}
	}

	return -1
}
