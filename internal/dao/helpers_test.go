// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestToPerc(t *testing.T) {
	uu := []struct {
		v1, v2, e float64
	}{
		{0, 0, 0},
		{100, 200, 50},
		{200, 100, 200},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, toPerc(u.v1, u.v2))
	}
}

func TestServiceAccountMatches(t *testing.T) {
	uu := []struct {
		podTemplate *v1.PodSpec
		saName      string
		expect      bool
	}{
		{podTemplate: &v1.PodSpec{
			ServiceAccountName: "",
		},
			saName: defaultServiceAccount,
			expect: true,
		},
		{podTemplate: &v1.PodSpec{
			ServiceAccountName: "",
		},
			saName: "foo",
			expect: false,
		},
		{podTemplate: &v1.PodSpec{
			ServiceAccountName: "foo",
		},
			saName: "foo",
			expect: true,
		},
		{podTemplate: &v1.PodSpec{
			ServiceAccountName: "foo",
		},
			saName: "bar",
			expect: false,
		},
	}

	for _, u := range uu {
		assert.Equal(t, u.expect, serviceAccountMatches(u.podTemplate.ServiceAccountName, u.saName))
	}
}
