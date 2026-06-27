// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var gatewayResList = []*client.GVR{
	client.GtwGVR,
	client.GtwGtwClassGVR,
	client.GtwHTTPRouteGVR,
	client.GtwGRPCRouteGVR,
	client.GtwTCPRouteGVR,
	client.GtwUDPRouteGVR,
	client.GtwTLSRouteGVR,
}

// GatewayAPI tracks Gateway API resources.
type GatewayAPI struct {
	Table
}

// List fetches Gateway API resources.
func (g *GatewayAPI) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	oo := make([]runtime.Object, 0, 100)
	for _, gvr := range gatewayResList {
		table, err := g.fetch(ctx, gvr, ns)
		if err != nil {
			slog.Warn("Gateway API resource fetch failed", slogs.GVR, gvr, slogs.Error, err)
			continue
		}

		for _, row := range table.Rows {
			var (
				resNs    string
				resName  string
				resTs    metav1.Time
				resObj   interface{}
			)

			if obj := row.Object.Object; obj != nil {
				m, err := meta.Accessor(obj)
				if err == nil {
					resNs, resName, resTs = m.GetNamespace(), m.GetName(), m.GetCreationTimestamp()
					// Convert empty namespace to ClusterScope for cluster-scoped resources
					if resNs == "" {
						resNs = client.ClusterScope
					}
					resObj = obj
				}
			} else {
				var m metav1.PartialObjectMetadata
				if err := json.Unmarshal(row.Object.Raw, &m); err == nil {
					resNs, resName, resTs = m.GetNamespace(), m.GetName(), m.GetCreationTimestamp()
					// Convert empty namespace to ClusterScope for cluster-scoped resources
					if resNs == "" {
						resNs = client.ClusterScope
					}
					resObj = m.DeepCopyObject()
				}
			}

			status := getGatewayStatus(gvr, &row, table.ColumnDefinitions, resObj)
		oo = append(oo, &render.GatewayAPIRes{TableRow: metav1.TableRow{Cells: []any{
			gvr.String(),
			resNs,
			resName,
			status,
			resTs,
		}}})
		}
	}

	return oo, nil
}

func (g *GatewayAPI) fetch(ctx context.Context, gvr *client.GVR, ns string) (*metav1.Table, error) {
	g.Init(g.Factory, gvr)

	oo, err := g.Table.List(ctx, ns)
	if err != nil {
		return nil, err
	}

	if len(oo) == 0 {
		return &metav1.Table{}, nil
	}

	if ta, ok := oo[0].(*metav1.Table); ok {
		return ta, nil
	}

	return nil, fmt.Errorf("unexpected type: %T", oo[0])
}

func getGatewayStatus(gvr *client.GVR, row *metav1.TableRow, h []metav1.TableColumnDefinition, obj interface{}) string {
	if obj == nil {
		return render.NAValue
	}

	switch gvr {
	case client.GtwGVR:
		return getGatewayResourceStatus(obj)
	case client.GtwGtwClassGVR:
		return getGatewayClassStatus(obj)
	case client.GtwHTTPRouteGVR:
		return getHTTPRouteStatus(obj)
	case client.GtwGRPCRouteGVR:
		return getGRPCRouteStatus(obj)
	case client.GtwTCPRouteGVR:
		return getTCPRouteStatus(obj)
	case client.GtwUDPRouteGVR:
		return getUDPRouteStatus(obj)
	case client.GtwTLSRouteGVR:
		return getTLSRouteStatus(obj)
	}

	return render.NAValue
}

func getGatewayResourceStatus(obj interface{}) string {
	if objMap, ok := obj.(map[string]interface{}); ok {
		if status, ok := objMap["status"].(map[string]interface{}); ok {
			if conditions, ok := status["conditions"].([]interface{}); ok {
				for _, cond := range conditions {
					if condMap, ok := cond.(map[string]interface{}); ok {
						if typ, ok := condMap["type"].(string); ok && typ == "Ready" {
							if statusCond, ok := condMap["status"].(string); ok {
								if statusCond == "True" {
									return StatusOK
								}
								return DegradedStatus
							}
						}
					}
				}
			}
		}
	}
	return DegradedStatus
}

func getGatewayClassStatus(obj interface{}) string {
	if objMap, ok := obj.(map[string]interface{}); ok {
		if status, ok := objMap["status"].(map[string]interface{}); ok {
			if conditions, ok := status["conditions"].([]interface{}); ok {
				for _, cond := range conditions {
					if condMap, ok := cond.(map[string]interface{}); ok {
						if typ, ok := condMap["type"].(string); ok && typ == "Ready" {
							if statusCond, ok := condMap["status"].(string); ok {
								if statusCond == "True" {
									return StatusOK
								}
								return DegradedStatus
							}
						}
					}
				}
			}
		}
	}
	return DegradedStatus
}

func getHTTPRouteStatus(obj interface{}) string {
	if objMap, ok := obj.(map[string]interface{}); ok {
		if status, ok := objMap["status"].(map[string]interface{}); ok {
			if parents, ok := status["parents"].([]interface{}); ok {
				for _, parent := range parents {
					if parentMap, ok := parent.(map[string]interface{}); ok {
						if conditions, ok := parentMap["conditions"].([]interface{}); ok {
							for _, cond := range conditions {
								if condMap, ok := cond.(map[string]interface{}); ok {
									if typ, ok := condMap["type"].(string); ok && typ == "Accepted" {
										if statusCond, ok := condMap["status"].(string); ok {
											if statusCond == "True" {
												return StatusOK
											}
											return DegradedStatus
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return DegradedStatus
}

func getGRPCRouteStatus(obj interface{}) string {
	return getHTTPRouteStatus(obj)
}

func getTCPRouteStatus(obj interface{}) string {
	if objMap, ok := obj.(map[string]interface{}); ok {
		if status, ok := objMap["status"].(map[string]interface{}); ok {
			if parents, ok := status["parents"].([]interface{}); ok {
				for _, parent := range parents {
					if parentMap, ok := parent.(map[string]interface{}); ok {
						if conditions, ok := parentMap["conditions"].([]interface{}); ok {
							for _, cond := range conditions {
								if condMap, ok := cond.(map[string]interface{}); ok {
									if typ, ok := condMap["type"].(string); ok && typ == "Accepted" {
										if statusCond, ok := condMap["status"].(string); ok {
											if statusCond == "True" {
												return StatusOK
											}
											return DegradedStatus
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return DegradedStatus
}

func getUDPRouteStatus(obj interface{}) string {
	return getTCPRouteStatus(obj)
}

func getTLSRouteStatus(obj interface{}) string {
	if objMap, ok := obj.(map[string]interface{}); ok {
		if status, ok := objMap["status"].(map[string]interface{}); ok {
			if parents, ok := status["parents"].([]interface{}); ok {
				for _, parent := range parents {
					if parentMap, ok := parent.(map[string]interface{}); ok {
						if conditions, ok := parentMap["conditions"].([]interface{}); ok {
							for _, cond := range conditions {
								if condMap, ok := cond.(map[string]interface{}); ok {
									if typ, ok := condMap["type"].(string); ok && typ == "Accepted" {
										if statusCond, ok := condMap["status"].(string); ok {
											if statusCond == "True" {
												return StatusOK
											}
											return DegradedStatus
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return DegradedStatus
}