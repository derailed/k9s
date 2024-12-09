// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Robert Bosch Manufacturing GmbH
package prettyjson

import (
	"testing"

	// "bytes"
	"encoding/json"
	// "fmt"
	// "io"
	// "regexp"
	// "strings"

	// "github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestRenderingForBasicTypes(t *testing.T) {
	testCases := []struct {
		name             string
		input            interface{}
		expectedContains string
	}{
		{
			name:  "Simple map with string key and value",
			input: map[string]string{"key": "value"},
			expectedContains: `{
  "key": "value"
}`,
		},
		{
			name:  "Simple array with string values",
			input: []string{"1", "2"},
			expectedContains: `[
  "1",
  "2"
]`,
		},
		{
			name:  "Simple map with string key and int value",
			input: map[string]int{"key": 1},
			expectedContains: `{
  "key": 1
}`,
		},
		{
			name:  "Simple array with int values",
			input: []int{1, 2},
			expectedContains: `[
  1,
  2
]`,
		},
		{
			name: "Nested map",
			input: map[string]map[string]int{"one": {"two": 3}},
			expectedContains: `{
  "one": {
    "two": 3
  }
}`,
		},
		{
			name: "Array of maps",
			input: []map[string]int{{"one": 1}, {"two": 2}},
			expectedContains: `[
  {
    "one": 1
  },
  {
    "two": 2
  }
]`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encoder := NewColorEncoder()
			input, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			buf, err := encoder.Encode([]byte(input))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			output := string(buf)
			assert.Equal(t, tc.expectedContains, output)
		})
	}
}

func TestTokenRenderingForComplexDataStructure(t *testing.T) {
	testJsonString := []byte(`{"one":1, "two":[["2",{"foo": "bar"}],[3]], "three":{"four":4}, "five": "5"}`)
	c := NewColorEncoder()
	p, _ := c.Encode(testJsonString)
	assert.Equal(t, `{
  "one": 1,
  "two": [
    [
      "2",
      {
        "foo": "bar"
      }
    ],
    [
      3
    ]
  ],
  "three": {
    "four": 4
  },
  "five": "5"
}`, string(p))
}
