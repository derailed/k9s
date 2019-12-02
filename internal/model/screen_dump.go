package model

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ScreenDump represents a container model.
type ScreenDump struct {
	Resource

	pod *v1.Pod
}

// List returns a collection of containers
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
		oo[i] = FileRes{file: f, dir: dir}
	}

	return oo, nil
}

// Hydrate returns a pod as container rows.
func (c *ScreenDump) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		res, ok := o.(FileRes)
		if !ok {
			return fmt.Errorf("expecting a file resource but got %T", o)
		}

		if err := re.Render(res, render.NonResource, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

// ----------------------------------------------------------------------------

// FileRes represents a file resource.
type FileRes struct {
	file os.FileInfo
	dir  string
}

func (c FileRes) GetFile() os.FileInfo { return c.file }
func (c FileRes) GetDir() string       { return c.dir }

// GetObjectKind returns a schema object.
func (c FileRes) GetObjectKind() schema.ObjectKind {

	return nil
}

// DeepCopyObject returns a container copy.
func (c FileRes) DeepCopyObject() runtime.Object {

	return c
}
