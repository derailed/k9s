// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

type Options struct {
	MetricsCacheExpiry int
}

func NewOptions(metricsCacheExpiry int) Options {
	return Options{MetricsCacheExpiry: metricsCacheExpiry}
}
