package client

import metricsapi "k8s.io/metrics/pkg/apis/metrics"

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	noMetricServerErr     = Error("No metrics-server detected")
	metricsUnsupportedErr = Error("No metrics api group " + metricsapi.GroupName + " found on cluster")
)
