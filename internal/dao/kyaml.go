// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
)

// plainKeyRX matches mapping keys that are safe to emit unquoted.
var plainKeyRX = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)

// ToKYAML converts a resource to its KYAML representation.
//
// KYAML is a safer, less ambiguous subset of YAML designed for Kubernetes.
// It requires explicit flow style collections ("{}" and "[]") and
// double-quotes all string scalars, eliminating YAML's whitespace
// sensitivity and implicit type coercion pitfalls.
// See: https://kubernetes.io/docs/reference/encodings/kyaml/
func ToKYAML(o runtime.Object, showManaged bool) (string, error) {
	raw, err := ToYAML(o, showManaged)
	if err != nil {
		return "", err
	}

	return YAMLToKYAML(raw)
}

// YAMLToKYAML converts a YAML document (potentially multi-document, as
// produced by Helm manifests) into its KYAML representation.
func YAMLToKYAML(raw string) (string, error) {
	dec := yaml.NewDecoder(strings.NewReader(raw))

	var docs []string
	for {
		var node yaml.Node
		if err := dec.Decode(&node); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		if len(node.Content) == 0 {
			continue
		}

		var buff bytes.Buffer
		writeKYAML(&buff, &node, 0)
		docs = append(docs, buff.String())
	}
	if len(docs) == 0 {
		return "", nil
	}

	return "---\n" + strings.Join(docs, "\n---\n") + "\n", nil
}

func writeKYAML(buff *bytes.Buffer, n *yaml.Node, indent int) {
	switch n.Kind {
	case yaml.DocumentNode:
		writeKYAML(buff, n.Content[0], indent)
	case yaml.MappingNode:
		writeMapping(buff, n, indent)
	case yaml.SequenceNode:
		writeSequence(buff, n, indent)
	case yaml.AliasNode:
		writeKYAML(buff, n.Alias, indent)
	default:
		buff.WriteString(scalarString(n))
	}
}

func writeMapping(buff *bytes.Buffer, n *yaml.Node, indent int) {
	if len(n.Content) == 0 {
		buff.WriteString("{}")
		return
	}

	buff.WriteString("{\n")
	childIndent := strings.Repeat("  ", indent+1)
	for i := 0; i < len(n.Content); i += 2 {
		key, val := n.Content[i], n.Content[i+1]
		buff.WriteString(childIndent)
		buff.WriteString(keyString(key))
		buff.WriteString(": ")
		writeKYAML(buff, val, indent+1)
		if i+2 < len(n.Content) {
			buff.WriteString(",")
		}
		buff.WriteString("\n")
	}
	buff.WriteString(strings.Repeat("  ", indent))
	buff.WriteString("}")
}

func writeSequence(buff *bytes.Buffer, n *yaml.Node, indent int) {
	if len(n.Content) == 0 {
		buff.WriteString("[]")
		return
	}

	buff.WriteString("[\n")
	childIndent := strings.Repeat("  ", indent+1)
	for i, c := range n.Content {
		buff.WriteString(childIndent)
		writeKYAML(buff, c, indent+1)
		if i+1 < len(n.Content) {
			buff.WriteString(",")
		}
		buff.WriteString("\n")
	}
	buff.WriteString(strings.Repeat("  ", indent))
	buff.WriteString("]")
}

// keyString renders a mapping key, quoting it only when required.
func keyString(n *yaml.Node) string {
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" && plainKeyRX.MatchString(n.Value) {
		return n.Value
	}

	var buff bytes.Buffer
	writeKYAML(&buff, n, 0)
	return buff.String()
}

// scalarString renders a scalar value: strings are always double-quoted,
// every other type (numbers, bools, null) is emitted as-is.
func scalarString(n *yaml.Node) string {
	switch n.Tag {
	case "!!null":
		return "null"
	case "!!str":
		return strconv.Quote(n.Value)
	default:
		if n.Value == "" {
			return "null"
		}
		return n.Value
	}
}
