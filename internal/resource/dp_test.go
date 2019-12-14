package resource_test

// BOZO!!
// import (
// 	"testing"

// 	"github.com/derailed/k9s/internal/k8s"
// 	"github.com/derailed/k9s/internal/resource"
// 	m "github.com/petergtz/pegomock"
// 	"github.com/stretchr/testify/assert"
// 	appsv1 "k8s.io/api/apps/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func NewDeploymentListWithArgs(ns string, r *resource.Deployment) resource.List {
// 	return resource.NewList(ns, "deploy", r, resource.AllVerbsAccess|resource.DescribeAccess)
// }

// func NewDeploymentWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Deployment {
// 	r := &resource.Deployment{Base: resource.NewBase(conn, res)}
// 	r.Factory = r
// 	return r
// }

// func TestDeploymentListAccess(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()

// 	ns := "blee"
// 	l := NewDeploymentListWithArgs(resource.AllNamespaces, NewDeploymentWithArgs(mc, mr))
// 	l.SetNamespace(ns)

// 	assert.Equal(t, "blee", l.GetNamespace())
// 	assert.Equal(t, "deploy", l.GetName())
// 	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
// 		assert.True(t, l.Access(a))
// 	}
// }

// func TestDeploymentFields(t *testing.T) {
// 	r := newDeployment().Fields("blee")
// 	assert.Equal(t, "fred", r[0])
// }

// func TestDeploymentMarshal(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.Get("blee", "fred")).ThenReturn(k8sDeployment(), nil)

// 	cm := NewDeploymentWithArgs(mc, mr)
// 	ma, err := cm.Marshal("blee/fred")

// 	mr.VerifyWasCalledOnce().Get("blee", "fred")
// 	assert.Nil(t, err)
// 	assert.Equal(t, dpYaml(), ma)
// }

// // BOZO!!
// // func TestDeploymentListData(t *testing.T) {
// // 	mc := NewMockConnection()
// // 	mr := NewMockCruder()
// // 	m.When(mr.List(resource.NotNamespaced, metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sDeployment()}, nil)

// // 	l := NewDeploymentListWithArgs("-", NewDeploymentWithArgs(mc, mr))
// // 	// Make sure we can get deltas!
// // 	for i := 0; i < 2; i++ {
// // 		err := l.Reconcile(nil, "", "")
// // 		assert.Nil(t, err)
// // 	}

// // 	mr.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced, metav1.ListOptions{})
// // 	td := l.Data()
// // 	assert.Equal(t, 1, len(td.Rows))
// // 	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
// // 	row := td.Rows["blee/fred"]
// // 	assert.Equal(t, 6, len(row.Deltas))
// // 	for _, d := range row.Deltas {
// // 		assert.Equal(t, "", d)
// // 	}
// // 	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
// // }

// // Helpers...

// func k8sDeployment() *appsv1.Deployment {
// 	var i int32 = 1
// 	return &appsv1.Deployment{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Namespace:         "blee",
// 			Name:              "fred",
// 			CreationTimestamp: metav1.Time{Time: testTime()},
// 		},
// 		Spec: appsv1.DeploymentSpec{
// 			Replicas: &i,
// 		},
// 	}
// }

// func newDeployment() resource.Columnar {
// 	mc := NewMockConnection()
// 	c, _ := resource.NewDeployment(mc).New(k8sDeployment())
// 	return c
// }

// func dpYaml() string {
// 	return `apiVersion: apps/v1
// kind: Deployment
// metadata:
//   creationTimestamp: "2018-12-14T17:36:43Z"
//   name: fred
//   namespace: blee
// spec:
//   replicas: 1
//   selector: null
//   strategy: {}
//   template:
//     metadata:
//       creationTimestamp: null
//     spec:
//       containers: null
// status: {}
// `
// }
