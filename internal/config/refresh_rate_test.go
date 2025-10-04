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
		expected    float32
	}{
		"integer_value": {
			yamlContent: `refreshRate: 2`,
			expected:    2.0,
		},
		"float_value": {
			yamlContent: `refreshRate: 2.5`,
			expected:    2.5,
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

func TestGetRefreshRateMinimum(t *testing.T) {
	tests := map[string]struct {
		refreshRate       float32
		manualRefreshRate float32
		expected          float32
	}{
		"below_minimum": {
			refreshRate: 0.5,
			expected:    2.0,
		},
		"at_minimum": {
			refreshRate: 2.0,
			expected:    2.0,
		},
		"above_minimum": {
			refreshRate: 3.5,
			expected:    3.5,
		},
		"manual_below_minimum": {
			refreshRate:       3.0,
			manualRefreshRate: 0.5,
			expected:          2.0,
		},
		"manual_above_minimum": {
			refreshRate:       2.0,
			manualRefreshRate: 4.0,
			expected:          4.0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := K9s{
				RefreshRate:       test.refreshRate,
				manualRefreshRate: test.manualRefreshRate,
			}
			assert.InDelta(t, test.expected, k.GetRefreshRate(), 0.001)
		})
	}
}
