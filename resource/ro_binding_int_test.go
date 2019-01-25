package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestToSubjectAlias(t *testing.T) {
	r := RoleBinding{}

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
		assert.Equal(t, u.e, r.toSubjectAlias(u.i))
	}
}

func TestToSubjects(t *testing.T) {
	r := RoleBinding{}

	uu := []struct {
		i []rbacv1.Subject
		e string
	}{
		{
			[]rbacv1.Subject{
				{Name: "blee", Kind: rbacv1.UserKind},
			},
			"blee/USR",
		},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, r.toSubjects(u.i))
	}
}

func BenchmarkToSubjects(b *testing.B) {
	var r RoleBinding
	ss := []rbacv1.Subject{
		{Name: "blee", Kind: rbacv1.UserKind},
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		r.toSubjects(ss)
	}
}
