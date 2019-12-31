package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
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
		assert.Equal(t, u.expected, config.InList(u.list, u.item))
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
		assert.Equal(t, u.expected, config.InNSList(u.list, u.item))
	}
}

func TestEnsurePathNone(t *testing.T) {
	var mod os.FileMode = 0744
	dir := filepath.Join("/tmp", "fred")
	os.Remove(dir)

	path := filepath.Join(dir, "duh.yml")
	config.EnsurePath(path, mod)

	p, err := os.Stat(dir)
	assert.Nil(t, err)
	assert.Equal(t, "drwxr--r--", p.Mode().String())
}

func TestEnsurePathNoOpt(t *testing.T) {
	var mod os.FileMode = 0744
	dir := filepath.Join("/tmp", "blee")
	os.Remove(dir)
	assert.Nil(t, os.Mkdir(dir, mod))

	path := filepath.Join(dir, "duh.yml")
	config.EnsurePath(path, mod)

	p, err := os.Stat(dir)
	assert.Nil(t, err)
	assert.Equal(t, "drwxr--r--", p.Mode().String())
}
