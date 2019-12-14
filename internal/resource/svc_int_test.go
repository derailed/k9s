package resource

// BOZO!!
// import (
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func TestSvcExtIPs(t *testing.T) {
// 	i := k8sSVCLb()

// 	var s Service
// 	ips := s.getSvcExtIPS(i)

// 	assert.Equal(t, "10.0.0.0,2.2.2.2", s.toIPs(i.Spec.Type, ips))
// }

// func TestLbIngressIP(t *testing.T) {
// 	lb := v1.LoadBalancerStatus{
// 		Ingress: []v1.LoadBalancerIngress{
// 			{IP: "10.0.0.0", Hostname: "fred"},
// 			{IP: "10.0.0.1", Hostname: "blee"},
// 		},
// 	}

// 	var s Service
// 	assert.Equal(t, "10.0.0.0,10.0.0.1", s.lbIngressIP(lb))
// }

// func TestToIPs(t *testing.T) {
// 	uu := []struct {
// 		t  v1.ServiceType
// 		ii []string
// 		e  string
// 	}{
// 		{v1.ServiceTypeLoadBalancer, []string{"2.2.2.2", "1.1.1.1"}, "1.1.1.1,2.2.2.2"},
// 		{v1.ServiceTypeLoadBalancer, []string{}, "<pending>"},
// 		{v1.ServiceTypeClusterIP, []string{}, MissingValue},
// 	}

// 	var s Service
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, s.toIPs(u.t, u.ii))
// 	}
// }

// func TestToPorts(t *testing.T) {
// 	uu := []struct {
// 		pp []v1.ServicePort
// 		e  string
// 	}{
// 		{[]v1.ServicePort{
// 			{Name: "http", Port: 80, NodePort: 90, Protocol: "TCP"}},
// 			"http:80►90",
// 		},
// 		{[]v1.ServicePort{
// 			{Port: 80, NodePort: 30080, Protocol: "UDP"}},
// 			"80►30080╱UDP",
// 		},
// 	}

// 	var s Service
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, s.toPorts(u.pp))
// 	}
// }

// func BenchmarkToPorts(b *testing.B) {
// 	sp := []v1.ServicePort{
// 		{Name: "http", Port: 80, NodePort: 90, Protocol: "TCP"},
// 		{Port: 80, NodePort: 90, Protocol: "TCP"},
// 		{Name: "http", Port: 80, NodePort: 90, Protocol: "TCP"},
// 	}
// 	b.ResetTimer()
// 	b.ReportAllocs()

// 	var s Service
// 	for i := 0; i < b.N; i++ {
// 		s.toPorts(sp)
// 	}
// }

// func k8sSVCLb() *v1.Service {
// 	return &v1.Service{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:              "fred",
// 			Namespace:         "blee",
// 			CreationTimestamp: metav1.Time{Time: testTime()},
// 		},
// 		Spec: v1.ServiceSpec{
// 			Type:        v1.ServiceTypeLoadBalancer,
// 			ClusterIP:   "1.1.1.1",
// 			ExternalIPs: []string{"2.2.2.2"},
// 			Selector:    map[string]string{"fred": "blee"},
// 			Ports: []v1.ServicePort{
// 				{
// 					Name:     "http",
// 					Port:     90,
// 					Protocol: "TCP",
// 				},
// 			},
// 		},
// 		Status: v1.ServiceStatus{
// 			LoadBalancer: v1.LoadBalancerStatus{
// 				Ingress: []v1.LoadBalancerIngress{
// 					{IP: "10.0.0.0", Hostname: "fred"},
// 				},
// 			},
// 		},
// 	}
// }

// func testTime() time.Time {
// 	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
// 	if err != nil {
// 		fmt.Println("TestTime Failed", err)
// 	}
// 	return t
// }
