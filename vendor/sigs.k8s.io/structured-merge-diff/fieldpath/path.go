/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fieldpath

import (
	"fmt"
	"strings"

	"sigs.k8s.io/structured-merge-diff/value"
)

// Path describes how to select a potentially deeply-nested child field given a
// containing object.
type Path []PathElement

func (fp Path) String() string {
	strs := make([]string, len(fp))
	for i := range fp {
		strs[i] = fp[i].String()
	}
	return strings.Join(strs, "")
}

func (fp Path) Copy() Path {
	new := make(Path, len(fp))
	copy(new, fp)
	return new
}

// MakePath constructs a Path. The parts may be PathElements, ints, strings.
func MakePath(parts ...interface{}) (Path, error) {
	var fp Path
	for _, p := range parts {
		switch t := p.(type) {
		case PathElement:
			fp = append(fp, t)
		case int:
			// TODO: Understand schema and object and convert this to the
			// FieldSpecifier below if appropriate.
			fp = append(fp, PathElement{Index: &t})
		case string:
			fp = append(fp, PathElement{FieldName: &t})
		case []value.Field:
			if len(t) == 0 {
				return nil, fmt.Errorf("associative list key type path elements must have at least one key (got zero)")
			}
			fp = append(fp, PathElement{Key: t})
		case value.Value:
			// TODO: understand schema and verify that this is a set type
			// TODO: make a copy of t
			fp = append(fp, PathElement{Value: &t})
		default:
			return nil, fmt.Errorf("unable to make %#v into a path element", p)
		}
	}
	return fp, nil
}

// MakePathOrDie panics if parts can't be turned into a path. Good for things
// that are known at complie time.
func MakePathOrDie(parts ...interface{}) Path {
	fp, err := MakePath(parts...)
	if err != nil {
		panic(err)
	}
	return fp
}
