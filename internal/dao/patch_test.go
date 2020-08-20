package dao

import (
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestGetTemplateJsonPatch(t *testing.T) {
	type args struct {
		podSpec v1.PodSpec
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				podSpec: v1.PodSpec{
					InitContainers: []v1.Container{v1.Container{Image: "busybox:latest", Name: "init"}},
					Containers:     []v1.Container{v1.Container{Image: "nginx:latest", Name: "nginx"}},
				},
			},
			want:    `{"spec":{"template":{"spec":{"$setElementOrder/containers":[{"name":"nginx"}],"$setElementOrder/initContainers":[{"name":"init"}],"containers":[{"image":"nginx:latest","name":"nginx"}],"initContainers":[{"image":"busybox:latest","name":"init"}]}}}}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTemplateJsonPatch(tt.args.podSpec)
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
		podSpec v1.PodSpec
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "simple",
			args: args{
				podSpec: v1.PodSpec{
					InitContainers: []v1.Container{v1.Container{Image: "busybox:latest", Name: "init"}},
					Containers:     []v1.Container{v1.Container{Image: "nginx:latest", Name: "nginx"}},
				},
			},
			want:    `{"spec":{"$setElementOrder/containers":[{"name":"nginx"}],"$setElementOrder/initContainers":[{"name":"init"}],"containers":[{"image":"nginx:latest","name":"nginx"}],"initContainers":[{"image":"busybox:latest","name":"init"}]}}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetJsonPatch(tt.args.podSpec)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.JSONEq(t, tt.want, got, "Json strings should be equal")
		})
	}
}
