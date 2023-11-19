// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import metricsapi "k8s.io/metrics/pkg/apis/metrics"

// Error represents an error.
type Error string

// Error returns the error text.
func (e Error) Error() string {
	return string(e)
}

const (
	noMetricServerErr     = Error("No metrics-server detected")
	metricsUnsupportedErr = Error("No metrics api group " + metricsapi.GroupName + " found on cluster")
)
