package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestToIPs(t *testing.T) {
	s := Service{}

	uu := []struct {
		t  v1.ServiceType
		ii []string
		e  string
	}{
		{v1.ServiceTypeLoadBalancer, []string{"2.2.2.2", "1.1.1.1"}, "1.1.1.1,2.2.2.2"},
		{v1.ServiceTypeLoadBalancer, []string{}, "<pending>"},
		{v1.ServiceTypeClusterIP, []string{}, MissingValue},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, s.toIPs(u.t, u.ii))
	}
}

func TestToPorts(t *testing.T) {
	var s Service

	uu := []struct {
		pp []v1.ServicePort
		e  string
	}{
		{[]v1.ServicePort{
			{Name: "http", Port: 80, NodePort: 90, Protocol: "TCP"}},
			"http:80►90",
		},
		{[]v1.ServicePort{
			{Port: 80, NodePort: 30080, Protocol: "UDP"}},
			"80►30080╱UDP",
		},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, s.toPorts(u.pp))
	}
}

func BenchmarkToPorts(b *testing.B) {
	var s Service
	sp := []v1.ServicePort{
		{Name: "http", Port: 80, NodePort: 90, Protocol: "TCP"},
		{Port: 80, NodePort: 90, Protocol: "TCP"},
		{Name: "http", Port: 80, NodePort: 90, Protocol: "TCP"},
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		s.toPorts(sp)
	}
}
