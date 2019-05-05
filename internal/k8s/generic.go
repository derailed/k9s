package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Get(ns, n, string, kind schema.GroupVersionKind, opts metav1.GetOptions) (runtime.Object, error) {
	return nil, nil
}

func List(ns string, kind schema.GroupVersionKind, opts metav1.ListOptions) (runtime.Object, error) {
	return nil, nil
}

func registrar() map[string]func() {
	return map[string]func(){}
}
