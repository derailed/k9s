// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Fields represents a collection of row fields.
type Fields []string

// Customize returns a subset of fields.
func (f Fields) Customize(cols []int, out Fields, extractionInfoBag ExtractionInfoBag) {
	for i, c := range cols {
		if c < 0 {

			// If current index can retrieve an extractionInfo from extractionInfoBag,
			// meaning this column has to retrieve the actual value from other field.
			// For example: `LABELS[kubernetes.io/hostname]` needs to extract the value from column `LABELS`
			if extractionInfo, ok := extractionInfoBag[i]; ok {
				idxInFields := extractionInfo.IdxInFields
				key := extractionInfo.Key

				// Escape dots from the key
				// For example: `kubernetes.io/hostname` needs to be escaped to `kubernetes\.io/hostname`
				escapedKey := strings.ReplaceAll(key, ".", "\\.")

				// Extract the value by using regex
				pattern := fmt.Sprintf(`%s=([^ ]+)`, escapedKey)
				regex := regexp.MustCompile(pattern)

				// Find the value in the field that store original values
				matches := regex.FindStringSubmatch(f[idxInFields])
				if len(matches) > 1 {
					out[i] = matches[1]
					continue
				}
			}

			out[i] = NAValue
			continue
		}
		if c < len(f) {
			out[i] = f[c]
		}
	}
}

// Diff returns true if fields differ or false otherwise.
func (f Fields) Diff(ff Fields, ageCol int) bool {
	if ageCol < 0 {
		return !reflect.DeepEqual(f[:len(f)-1], ff[:len(ff)-1])
	}
	if !reflect.DeepEqual(f[:ageCol], ff[:ageCol]) {
		return true
	}
	return !reflect.DeepEqual(f[ageCol+1:], ff[ageCol+1:])
}

// Clone returns a copy of the fields.
func (f Fields) Clone() Fields {
	cp := make(Fields, len(f))
	copy(cp, f)

	return cp
}
