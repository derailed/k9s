package dao

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetTemplateJsonPatch(t *testing.T) {
	type args struct {
		containers     map[string]string
		initContainers map[string]string
	}
	tests := map[string]struct {
		args    args
		want    string
		wantErr bool
	}{
		"simple": {
			args: args{
				initContainers: map[string]string{"init": "busybox:latest"},
				containers:     map[string]string{"nginx": "nginx:latest"},
			},
			want:    `{"spec":{"template":{"spec":{"$setElementOrder/containers":[{"name":"nginx"}],"$setElementOrder/initContainers":[{"name":"init"}],"containers":[{"image":"nginx:latest","name":"nginx"}],"initContainers":[{"image":"busybox:latest","name":"init"}]}}}}`,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateJsonPatch(tt.args.containers, tt.args.initContainers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.JSONEq(t, tt.want, got, "Json strings should be equal")
		})
	}
}

func TestGetJsonPatch(t *testing.T) {
	type args struct {
		containers     map[string]string
		initContainers map[string]string
	}
	tests := map[string]struct {
		args    args
		want    string
		wantErr bool
	}{
		"simple": {
			args: args{
				initContainers: map[string]string{"init": "busybox:latest"},
				containers:     map[string]string{"nginx": "nginx:latest"},
			},
			want:    `{"spec":{"$setElementOrder/containers":[{"name":"nginx"}],"$setElementOrder/initContainers":[{"name":"init"}],"containers":[{"image":"nginx:latest","name":"nginx"}],"initContainers":[{"image":"busybox:latest","name":"init"}]}}`,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetJsonPatch(tt.args.containers, tt.args.initContainers)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.JSONEq(t, tt.want, got, "Json strings should be equal")
		})
	}
}
