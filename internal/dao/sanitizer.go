package dao

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	cfg "github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/popeye/pkg"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Sanitizer)(nil)

// Sanitizer tracks cluster sanitization.
type Sanitizer struct {
	NonResource
}

// NewSanitizer returns a new set of aliases.
func NewSanitizer(f Factory) *Sanitizer {
	s := Sanitizer{}
	s.Init(f, client.NewGVR("report"))

	return &s
}

// List returns a collection of aliases.
func (s *Sanitizer) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	report, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, fmt.Errorf("no sanitizer report path")
	}
	sections := []string{report}
	js := "json"
	flags := config.NewFlags()
	flags.Sections = &sections
	spinach := filepath.Join(cfg.K9sHome, "spinach.yml")
	flags.Spinach = &spinach
	flags.Output = &js

	popeye, err := pkg.NewPopeye(flags, &log.Logger)
	if err != nil {
		return nil, err
	}
	popeye.SetFactory(newPopFactory(s.Factory))
	if err = popeye.Init(); err != nil {
		return nil, err
	}
	buff := readWriteCloser{Buffer: bytes.NewBufferString("")}
	popeye.SetOutputTarget(buff)
	if err = popeye.Sanitize(); err != nil {
		return nil, err
	}

	var b render.Builder
	if err = json.Unmarshal(buff.Bytes(), &b); err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(b.Report.Sections))
	for i, s := range b.Report.Sections {
		oo[i] = s
	}

	return oo, nil
}

// Get fetch a resource.
func (*Sanitizer) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("NYI!!")
}
