// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestIsSameSubject(t *testing.T) {
	uu := map[string]struct {
		kind      string
		namespace string
		name      string
		subject   rbacv1.Subject
		want      bool
	}{
		"kind-name-match": {
			kind: rbacv1.UserKind,
			name: "foo",
			subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "foo",
			},
			want: true,
		},
		"name-does-not-match": {
			kind: rbacv1.UserKind,
			name: "foo",
			subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "bar",
			},
			want: false,
		},
		"kind-does-not-match": {
			kind: rbacv1.GroupKind,
			name: "foo",
			subject: rbacv1.Subject{
				Kind: rbacv1.UserKind,
				Name: "foo",
			},
			want: false,
		},
		"serviceAccount-all-match": {
			kind:      rbacv1.ServiceAccountKind,
			name:      "foo",
			namespace: "bar",
			subject: rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "foo",
				Namespace: "bar",
			},
			want: true,
		},
		"serviceAccount-namespace-no-match": {
			kind:      rbacv1.ServiceAccountKind,
			name:      "foo",
			namespace: "bar",
			subject: rbacv1.Subject{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      "foo",
				Namespace: "bazz",
			},
			want: false,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			same := isSameSubject(u.kind, u.namespace, u.name, &u.subject)
			assert.Equal(t, u.want, same)
		})
	}
}
