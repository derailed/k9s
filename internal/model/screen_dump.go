package model

import (
	"context"
	"errors"
	"io/ioutil"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// ScreenDump represents a collections of screendumps.
type ScreenDump struct {
	Resource
}

// List returns a collection of screen dumps.
func (c *ScreenDump) List(ctx context.Context) ([]runtime.Object, error) {
	dir, ok := ctx.Value(internal.KeyDir).(string)
	if !ok {
		return nil, errors.New("no screendump dir found in context")
	}

	ff, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(ff))
	for i, f := range ff {
		oo[i] = render.FileRes{File: f, Dir: dir}
	}

	return oo, nil
}

// Hydrate returns a pod as container rows.
func (c *ScreenDump) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		if err := re.Render(o, render.NonResource, &rr[i]); err != nil {
			return err
		}
	}
	return nil
}
