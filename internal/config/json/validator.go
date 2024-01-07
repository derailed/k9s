// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package json

import (
	"cmp"
	_ "embed"
	"errors"
	"fmt"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

const (
	// PluginsSchema describes plugins schema.
	PluginsSchema = "plugins.json"

	// AliasesSchema describes aliases schema.
	AliasesSchema = "aliases.json"

	// ViewsSchema describes views schema.
	ViewsSchema = "views.json"

	// HotkeysSchema describes hotkeys schema.
	HotkeysSchema = "hotkeys.json"

	// K9sSchema describes k9s config schema.
	K9sSchema = "k9s.json"

	// ContextSchema describes context config schema.
	ContextSchema = "context.json"

	// SkinSchema describes skin config schema.
	SkinSchema = "skin.json"
)

var (
	//go:embed schemas/plugins.json
	pluginSchema string

	//go:embed schemas/aliases.json
	aliasSchema string

	//go:embed schemas/views.json
	viewsSchema string

	//go:embed schemas/k9s.json
	k9sSchema string

	//go:embed schemas/context.json
	contextSchema string

	//go:embed schemas/hotkeys.json
	hotkeysSchema string

	//go:embed schemas/skin.json
	skinSchema string
)

// Validator tracks schemas validation.
type Validator struct {
	schemas map[string]gojsonschema.JSONLoader
	loader  *gojsonschema.SchemaLoader
}

// NewValidator returns a new instance.
func NewValidator() *Validator {
	v := Validator{
		schemas: map[string]gojsonschema.JSONLoader{
			K9sSchema:     gojsonschema.NewStringLoader(k9sSchema),
			ContextSchema: gojsonschema.NewStringLoader(contextSchema),
			AliasesSchema: gojsonschema.NewStringLoader(aliasSchema),
			ViewsSchema:   gojsonschema.NewStringLoader(viewsSchema),
			PluginsSchema: gojsonschema.NewStringLoader(pluginSchema),
			HotkeysSchema: gojsonschema.NewStringLoader(hotkeysSchema),
			SkinSchema:    gojsonschema.NewStringLoader(skinSchema),
		},
	}
	v.register()

	return &v
}

// Init initializes the schemas.
func (v *Validator) register() {
	v.loader = gojsonschema.NewSchemaLoader()
	v.loader.Validate = true
	for k, s := range v.schemas {
		if err := v.loader.AddSchema(k, s); err != nil {
			log.Error().Err(err).Msgf("schema initialization failed: %q", k)
		}
	}
}

// Validate runs document thru given schema validation.
func (v *Validator) Validate(k string, bb []byte) error {
	var m interface{}
	err := yaml.Unmarshal(bb, &m)
	if err != nil {
		return err
	}

	s, ok := v.schemas[k]
	if !ok {
		return fmt.Errorf("no schema found for: %q", k)
	}
	result, err := gojsonschema.Validate(s, gojsonschema.NewGoLoader(m))
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}

	slices.SortFunc(result.Errors(), func(a, b gojsonschema.ResultError) int {
		return cmp.Compare(a.Description(), b.Description())
	})
	var errs error
	for _, re := range result.Errors() {
		errs = errors.Join(errs, errors.New(re.Description()))
	}

	return errs
}

func (v *Validator) ValidateObj(k string, o any) error {
	s, ok := v.schemas[k]
	if !ok {
		return fmt.Errorf("no schema found for: %q", k)
	}
	result, err := gojsonschema.Validate(s, gojsonschema.NewGoLoader(o))
	if err != nil {
		return err
	}
	if result.Valid() {
		return nil
	}

	slices.SortFunc(result.Errors(), func(a, b gojsonschema.ResultError) int {
		return cmp.Compare(a.Description(), b.Description())
	})
	var errs error
	for _, re := range result.Errors() {
		errs = errors.Join(errs, errors.New(re.Description()))
	}

	return errs
}
