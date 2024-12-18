// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"reflect"
)

// Fields represents a collection of row fields.
type Fields []string

// Customize returns a subset of fields.
func (f Fields) Customize(cols []int, out Fields, extractionInfoBag ExtractionInfoBag) {
	for i, c := range cols {

		if c < 0 {
			out[i] = getValueOfInvalidColumn(f, i, extractionInfoBag)
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

func getValueOfInvalidColumn(f Fields, i int, extractionInfoBag ExtractionInfoBag) string {

	extractionInfo, ok := extractionInfoBag[i]

	// If the extractionInfo is existed in extractionInfoBag,
	// meaning this column has to retrieve the actual value from other field.
	// For example: `LABELS[kubernetes.io/hostname]` needs to extract the value from column `LABELS`
	if ok {
		idxInFields := extractionInfo.IdxInFields
		escapedKey := escapeDots(extractionInfo.Key)
		return extractValueFromField(escapedKey, f[idxInFields])
	}

	return NAValue
}
