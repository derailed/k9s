package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestToSubjectAlias(t *testing.T) {
	uu := []struct {
		i string
		e string
	}{
		{rbacv1.UserKind, "USR"},
		{rbacv1.GroupKind, "GRP"},
		{rbacv1.ServiceAccountKind, "SA"},
		{"fred", "FRED"},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, toSubjectAlias(u.i))
	}
}

func TestRenderSubjects(t *testing.T) {
	uu := []struct {
		ss []rbacv1.Subject
		ek string
		e  string
	}{
		{
			[]rbacv1.Subject{
				{Name: "blee", Kind: rbacv1.UserKind},
			},
			"USR",
			"blee",
		},
		{
			[]rbacv1.Subject{},
			NAValue,
			"",
		},
	}
	for _, u := range uu {
		kind, ss := renderSubjects(u.ss)
		assert.Equal(t, u.e, ss)
		assert.Equal(t, u.ek, kind)
	}
}

func BenchmarkToSubjects(b *testing.B) {
	ss := []rbacv1.Subject{
		{Name: "blee", Kind: rbacv1.UserKind},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		renderSubjects(ss)
	}
}
