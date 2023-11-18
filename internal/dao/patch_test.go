// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTemplateJsonPatch(t *testing.T) {
	type args struct {
		imageSpecs ImageSpecs
	}
	uu := map[string]struct {
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
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			got, err := GetTemplateJsonPatch(u.args.imageSpecs)
			if (err != nil) != u.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, u.wantErr)
				return
			}
			require.JSONEq(t, u.want, string(got), "Json strings should be equal")
		})
	}
}

func TestGetJsonPatch(t *testing.T) {
	type args struct {
		imageSpecs ImageSpecs
	}
	uu := map[string]struct {
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
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			got, err := GetJsonPatch(u.args.imageSpecs)
			if (err != nil) != u.wantErr {
				t.Errorf("GetTemplateJsonPatch() error = %v, wantErr %v", err, u.wantErr)
				return
			}
			require.JSONEq(t, u.want, string(got), "Json strings should be equal")
		})
	}
}
