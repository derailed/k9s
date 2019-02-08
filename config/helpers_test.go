package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHelperInList(t *testing.T) {
	uu := []struct {
		item     string
		list     []string
		expected bool
	}{
		{"a", []string{}, false},
		{"", []string{}, false},
		{"", []string{""}, true},
		{"a", []string{"a", "b", "c", "d"}, true},
		{"z", []string{"a", "b", "c", "d"}, false},
	}

	for _, u := range uu {
		assert.Equal(t, u.expected, InList(u.list, u.item))
	}
}

func TestHelperInNSList(t *testing.T) {
	uu := []struct {
		item     string
		list     []interface{}
		expected bool
	}{
		{
			"fred",
			[]interface{}{
				v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fred"}},
			},
			true,
		},
		{
			"blee",
			[]interface{}{
				v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "fred"}},
			},
			false,
		},
	}

	for _, u := range uu {
		assert.Equal(t, u.expected, InNSList(u.list, u.item))
	}
}
