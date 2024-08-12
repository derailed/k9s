// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/vul"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*ImageScan)(nil)
)

// ImageScan represents vulnerability scans.
type ImageScan struct {
	NonResource
}

func (is *ImageScan) listImages(ctx context.Context, gvr client.GVR, path string) ([]string, error) {
	res, err := AccessorFor(is.Factory, gvr)
	if err != nil {
		return nil, err
	}
	s, ok := res.(ImageLister)
	if !ok {
		return nil, fmt.Errorf("resource %s is not image lister: %T", gvr, res)
	}

	return s.ListImages(ctx, path)
}

// List returns a collection of scans.
func (is *ImageScan) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, fmt.Errorf("no context path for %q", is.gvr)
	}
	gvr, ok := ctx.Value(internal.KeyGVR).(client.GVR)
	if !ok {
		return nil, fmt.Errorf("no context gvr for %q", is.gvr)
	}

	ii, err := is.listImages(ctx, gvr, fqn)
	if err != nil {
		return nil, err
	}

	res := make([]runtime.Object, 0, len(ii))
	for _, img := range ii {
		s, ok := vul.ImgScanner.GetScan(img)
		if !ok {
			continue
		}
		for _, r := range s.Table.Rows {
			res = append(res, render.ImageScanRes{Image: img, Row: r})
		}
	}

	return res, nil
}
