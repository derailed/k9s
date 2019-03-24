package k8s

import (
	"math"
)

const megaByte = 1024 * 1024

func asMi(v int64) float64 {
	return float64(v) / megaByte
}

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return math.Round((v1 / v2) * 100)
}
