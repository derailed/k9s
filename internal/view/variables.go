// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui/dialog"
)

type ExtractedEnvVar struct {
	Key   string
	Value string
}

func extractVariables(r Runner, vars *[]config.Variable, env *Env, index int, done func()) {
	v := (*vars)[index]
	extractVariableAction(r, &v, env, func(variable *ExtractedEnvVar, err error) {
		if err != nil {
			slog.Error("Variable resolution failed", slogs.Error, err)
			r.App().Flash().Infof("Variable resolution failed for '%q': %q", v.Name, err.Error())
			return
		}

		(*env)[variable.Key] = variable.Value

		if index < len(*vars)-1 {
			go extractVariables(r, vars, env, index+1, done)
		} else {
			done()
		}
	})
}

func extractVariableAction(r Runner, v *config.Variable, e *Env, done func(envVar *ExtractedEnvVar, err error)) {
	datas, err := extractVariableDatas(r, v, e)
	if err != nil {
		done(nil, err)
		return
	}

	if len(datas) == 0 && v.Display != config.VariableDisplayText {
		done(nil, fmt.Errorf("variable source didn't provide any data"))
		return
	}

	d := r.App().Styles.Dialog()

	switch v.Display {
	case config.VariableDisplayNone:
		if len(datas) > 1 {
			done(nil, fmt.Errorf("variable source provide too many data"))
			return
		}
		envVar := ExtractedEnvVar{
			Key:   v.Name,
			Value: datas[0],
		}
		done(&envVar, nil)

	case config.VariableDisplayText:
		if len(datas) > 1 {
			done(nil, fmt.Errorf("variable source provide too many data"))
			return
		}

		textOpts := dialog.TextDialogOpts{
			Title:           "Variable '" + v.Name + "'",
			Text:            "Type a value for the variable '" + v.Name + "'",
			Label:           "Value: ",
			Placeholder:     "Expected value for '" + v.Name + "'...",
			InitialValue:    datas[0],
			MaxSize:         0,
			AllowEmptyValue: false,
			Selected: func(value *string) {
				envVar := ExtractedEnvVar{
					Key:   v.Name,
					Value: *value,
				}
				done(&envVar, nil)
			},
			Cancel: func() {
				done(nil, fmt.Errorf("no value typed for %q", v.Name))
			},
		}
		dialog.ShowText(&d, r.App().Content.Pages, &textOpts)

	case config.VariableDisplaySelect:
		listOpts := dialog.ListSelectionDialogOpts{
			Title:                "Variable '" + v.Name + "'",
			Text:                 "Select a value for the variable '" + v.Name + "'",
			Label:                "Value: ",
			Options:              datas,
			InitialValueIndex:    0,
			AllowUnSelectedValue: false,
			Selected: func(_ int, value string) {
				envVar := ExtractedEnvVar{
					Key:   v.Name,
					Value: value,
				}
				done(&envVar, nil)
			},
			Cancel: func() {
				done(nil, fmt.Errorf("no value selected for %q", v.Name))
			},
		}
		dialog.ShowListSelection(&d, r.App().Content.Pages, &listOpts)
	default:
		done(nil, fmt.Errorf("unmanaged display options"))
	}
}

func extractVariableDatas(r Runner, v *config.Variable, e *Env) ([]string, error) {
	args := make([]string, len(v.Args))
	for i, a := range v.Args {
		arg, err := e.Substitute(a)
		if err != nil {
			slog.Error("Variables Args env var match failed", slogs.Error, err)
			return nil, fmt.Errorf("variables args env var match failed")
		}
		args[i] = arg
	}
	v.Args = args

	switch v.Source {
	case config.VariableSourceScript:
		pipes := make([]string, len(v.Pipes))
		for i, p := range v.Pipes {
			pipe, err := e.Substitute(p)
			if err != nil {
				slog.Error("Variables Pipes env var match failed", slogs.Error, err)
				return nil, fmt.Errorf("variables pipes env var match failed")
			}
			pipes[i] = pipe
		}
		v.Pipes = pipes

		res, err := executeVariableScript(r, v)
		if err != nil {
			return nil, err
		}
		return res, nil

	case config.VariableSourceStatic:
		return args, nil

	default:
		return nil, fmt.Errorf("unmanaged Variable origin")
	}
}

func executeVariableScript(r Runner, v *config.Variable) ([]string, error) {
	opts := shellOpts{
		binary:     v.Command,
		background: true,
		pipes:      v.Pipes,
		args:       v.Args,
	}
	suspend, errChan, outChan := run(r.App(), &opts)
	if !suspend {
		r.App().Flash().Infof("Variable command failed: %q", v.Name)
		return nil, fmt.Errorf("variable command failed: %q", v.Name)
	}
	var errs error
	for e := range errChan {
		errs = errors.Join(errs, e)
	}
	if errs != nil {
		if !strings.Contains(errs.Error(), "signal: interrupt") {
			slog.Error("Variable command failed", slogs.Error, errs)
			r.App().cowCmd(errs.Error())
			return nil, errs
		}
	}

	var results []string
	for out := range outChan {
		if strings.Contains(out, outputPrefix) {
			results = append(results, strings.TrimPrefix(out, outputPrefix+" "))
		}
	}

	return results, nil
}
