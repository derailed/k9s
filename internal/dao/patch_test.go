package dao

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetTemplateJsonPatch(t *testing.T) {
	type args struct {
		imageSpecs ImageSpecs
	}
	tests := map[string]struct {
		args    args
		want    string
		wantErr bool
	}{
		"simple": {
			args: args{
				imageSpecs: ImageSpecs{
					ImageSpec{
						Index:       0,
						Name:        "init",
						DockerImage: "busybox:latest",
						Init:        true,
					},
					ImageSpec{
						Index:       0,
						Name:        "nginx",
						DockerImage: "nginx:latest",
						Init:        false,
					},
				},
			},
			want:    `{"spec":{"template":{"spec":{"$setElementOrder/initContainers":[{"name":"init"}],"$setElementOrder/containers":[{"name":"nginx"}],"initContainers":[{"image":"busybox:latest","name":"init"}],"containers":[{"image":"nginx:latest","name":"nginx"}]}}}}`,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetTemplateJsonPatch(tt.args.imageSpecs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.JSONEq(t, tt.want, string(got), "Json strings should be equal")
		})
	}
}

func TestGetJsonPatch(t *testing.T) {
	type args struct {
		imageSpecs ImageSpecs
	}
	tests := map[string]struct {
		args    args
		want    string
		wantErr bool
	}{
		"simple": {
			args: args{
				imageSpecs: ImageSpecs{
					ImageSpec{
						Index:       0,
						Name:        "init",
						DockerImage: "busybox:latest",
						Init:        true,
					},
					ImageSpec{
						Index:       0,
						Name:        "nginx",
						DockerImage: "nginx:latest",
						Init:        false,
					},
				},
			},
			want:    `{"spec":{"$setElementOrder/initContainers":[{"name":"init"}],"initContainers":[{"image":"busybox:latest","name":"init"}],"$setElementOrder/containers":[{"name":"nginx"}],"containers":[{"image":"nginx:latest","name":"nginx"}]}}`,
			wantErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := GetJsonPatch(tt.args.imageSpecs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.JSONEq(t, tt.want, string(got), "Json strings should be equal")
		})
	}
}
