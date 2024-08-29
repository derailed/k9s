// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
)

// HorizontalPodAutoscaler renders a K8s HorizontalPodAutoscaler to screen.
type HorizontalPodAutoscaler struct {
	Generic
}

// ColorerFunc colors a resource row.
func (hpa HorizontalPodAutoscaler) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)

		maxPodsIndex, ok := h.IndexOf("MAXPODS", true)
		if !ok || maxPodsIndex >= len(re.Row.Fields) {
			return c
		}

		replicasIndex, ok := h.IndexOf("REPLICAS", true)
		if !ok || replicasIndex >= len(re.Row.Fields) {
			return c
		}

		maxPodsS := strings.TrimSpace(re.Row.Fields[maxPodsIndex])
		currentReplicasS := strings.TrimSpace(re.Row.Fields[replicasIndex])

		maxPods, err := strconv.Atoi(maxPodsS)
		if err != nil {
			return c
		}
		currentReplicas, err := strconv.Atoi(currentReplicasS)
		if err != nil {
			return c
		}
		if currentReplicas >= maxPods {
			c = model1.ErrColor
		}
		return c
	}
}
