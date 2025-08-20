package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRefreshRateBackwardCompatibility(t *testing.T) {
	tests := map[string]struct {
		yamlContent string
		expected    float64
	}{
		"integer_value": {
			yamlContent: `refreshRate: 2`,
			expected:    2.0,
		},
		"float_value": {
			yamlContent: `refreshRate: 2.5`,
			expected:    2.5,
		},
		"sub_second": {
			yamlContent: `refreshRate: 0.5`,
			expected:    0.5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			var k K9s
			err := yaml.Unmarshal([]byte(test.yamlContent), &k)
			require.NoError(t, err)
			assert.InDelta(t, test.expected, k.RefreshRate, 0.001)
		})
	}
}
