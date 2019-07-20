package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionFromAnnotation(t *testing.T) {
	ann := map[string]string{
		"kubectl.kubernetes.io/last-applied-configuration": `{"apiVersion":"autoscaling/v1","kind":"HorizontalPodAutoscaler","metadata":{"annotations":{},"name":"nginx","namespace":"default"},"spec":{"maxReplicas":10,"minReplicas":1,"scaleTargetRef":{"apiVersion":"apps/v1","kind":"Deployment","name":"nginx"},"targetCPUUtilizationPercentage":10}}`,
	}

	assert.Equal(t, "autoscaling/v1", extractVersion(ann))
}
