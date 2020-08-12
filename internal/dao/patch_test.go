package dao

import (
	"reflect"
	"testing"
)

func TestSetImageJsonPatch(t *testing.T) {
	type args struct {
		images map[string]string
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
				images: map[string]string{"nginx": "nginx:latest"},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetImageJsonPatch(tt.args.images)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetImageJsonPatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetImageJsonPatch() got = %v, want %v", got, tt.want)
			}
		})
	}
}
