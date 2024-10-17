package dao

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

type obj struct {
	Value string `json:"value,omitempty"`
	Child struct {
		Nested string `json:"nested,omitempty"`
	} `json:"child"`
}

func (o obj) GetObjectKind() schema.ObjectKind {
	return nil
}

func (o obj) DeepCopyObject() runtime.Object {
	return nil
}

func TestExtractPathFromObject(t *testing.T) {
	tests := []struct {
		name string
		obj  runtime.Object
		path string
		want []string
	}{
		{
			name: "empty path",
			obj: &obj{
				Value: "",
				Child: struct {
					Nested string `json:"nested,omitempty"`
				}{},
			},
			path: ".nonexistent",
			want: []string{},
		},
		{
			name: "object with path",
			obj: &obj{
				Value: "foo",
				Child: struct {
					Nested string `json:"nested,omitempty"`
				}{},
			},
			path: ".value",
			want: []string{"foo"},
		},
		{
			name: "nested with path",
			obj: &obj{
				Value: "foo",
				Child: struct {
					Nested string `json:"nested,omitempty"`
				}{
					Nested: "bar",
				},
			},
			path: ".child.nested",
			want: []string{"bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractPathFromObject(tt.obj, tt.path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
