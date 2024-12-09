// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2024 Robert Bosch Manufacturing GmbH
package prettyjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

// ColorRule defines a coloring rule for string fields
type ColorRule struct {
	KeyPattern *regexp.Regexp
	Pattern    *regexp.Regexp
	Color      *color.Color
}

// PrimitiveColor set coloring for primitive types

type PrimitiveColor struct {
	BooleanColor *color.Color
	NumberColor  *color.Color
	NullColor    *color.Color
	StringColor  *color.Color
	KeyColor     *color.Color
}

// ColorEncoder wraps the standard JSON encoder with color formatting
type ColorEncoder struct {
	typeColors  PrimitiveColor
	stringRules []ColorRule
}

// NewColorEncoder creates a new color encoder writing to the given writer
func NewColorEncoder() *ColorEncoder {
	return &ColorEncoder{
		typeColors: PrimitiveColor{
			BooleanColor: color.New(color.FgYellow),
			NumberColor:  color.New(color.FgCyan),
			NullColor:    color.New(color.FgMagenta),
			StringColor:  color.New(color.FgWhite),
			KeyColor:     color.New(color.FgWhite),
		},
		stringRules: []ColorRule{},
	}
}

// AddStringRule adds a new rule for coloring string fields based on regex
func (c *ColorEncoder) AddStringRule(keypattern string, pattern string, color *color.Color) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}
	keyregex, err := regexp.Compile(keypattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}
	c.stringRules = append(c.stringRules, ColorRule{KeyPattern: keyregex, Pattern: regex, Color: color})
	return nil
}

// Encode colorizes and encodes the value
func (c *ColorEncoder) Encode(v []byte) ([]byte, error) {
	// Now process the JSON tokens and add colors
	var processed bytes.Buffer
	decoder := json.NewDecoder(bytes.NewReader(v))
	decoder.UseNumber()
	err := c.colorizeTokens(decoder, &processed)
	return processed.Bytes(), err
}

// Main function used to colorize JSON tokens
func (c *ColorEncoder) colorizeTokens(dec *json.Decoder, out *bytes.Buffer) error {
	token, err := dec.Token()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	switch v := token.(type) {
	case json.Delim:
		switch v {
		case '{':
			err := c.processMap(dec, out, 1)
			if err != nil {
				return err
			}
			out.WriteString("}")
		case '[':
			err := c.processArray(dec, out, 1)
			if err != nil {
				return err
			}
			out.WriteString("]")
		}
	}
	return nil
}

// process an array
func (c *ColorEncoder) processArray(dec *json.Decoder, out *bytes.Buffer, depth int) error {
	indent := strings.Repeat("  ", depth)
	out.WriteString("[\n")
	for {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		switch v := token.(type) {
		case json.Delim:
			switch v {
			case '{':
				out.WriteString(indent)
				err := c.processMap(dec, out, depth+1)
				if err != nil {
					return err
				}
				out.WriteString(indent)
				out.WriteString("}")
			case '[':
				out.WriteString(indent)
				err := c.processArray(dec, out, depth+1)
				if err != nil {
					return err
				}
				out.WriteString(indent)
				out.WriteString("]")
			case ']', '}':
				if dec.More() && depth == 1 {
					continue
				} else {
					return nil
				}

			}
		default:
			out.WriteString(indent)
			c.writeValue(out, v, "")
		}
		if dec.More() {
			out.WriteString(",")
		}
		out.WriteString("\n")
	}
}

// process a map
func (c *ColorEncoder) processMap(dec *json.Decoder, out *bytes.Buffer, depth int) error {
	indent := strings.Repeat("  ", depth)
	out.WriteString("{\n")
	for {
		key, err := dec.Token()
		if err != nil {
			return err
		}
		switch k := key.(type) {
		case json.Delim:
			switch k {
			case '}', ']':
				if dec.More() && depth == 1 {
					continue
				} else {
					return nil
				}
			}
		case string:
			color := c.typeColors.KeyColor
			out.WriteString(indent)
			out.WriteString(color.Sprintf("\"%v\": ", key))
		}
		shouldReturn, returnValue := c.processValue(dec, out, depth, key.(string))
		if shouldReturn {
			return returnValue
		}
		if dec.More() {
			out.WriteString(",")
		}
		out.WriteString("\n")

	}
}

// process a map value
func (c *ColorEncoder) processValue(dec *json.Decoder, out *bytes.Buffer, depth int, key string) (bool, error) {
	indent := strings.Repeat("  ", depth)
	value, err := dec.Token()
	if err != nil {
		return true, err
	}
	switch v := value.(type) {
	case json.Delim:
		switch v {
		case '{':
			err := c.processMap(dec, out, depth+1)
			if err != nil {
				return true, err
			}
			out.WriteString(indent)
			out.WriteString("}")
		case '[':
			err := c.processArray(dec, out, depth+1)
			if err != nil {
				return true, err
			}
			out.WriteString(indent)
			out.WriteString("]")
		}
	default:
		c.writeValue(out, v, key)

	}
	return false, nil
}

// Write a basic type map value or array item
func (c *ColorEncoder) writeValue(out *bytes.Buffer, v interface{}, key string) {
	switch val := v.(type) {
	case string:
		// Check string rules first
		color := c.typeColors.StringColor
		for _, rule := range c.stringRules {
			if key != "" && rule.KeyPattern.MatchString(key) && rule.Pattern.MatchString(val) {
				color = rule.Color
				break
			}
		}
		out.WriteString(color.Sprintf("%q", val))
	case json.Number:
		out.WriteString(c.typeColors.NumberColor.Sprintf("%v",val))
	case bool:
		out.WriteString(c.typeColors.BooleanColor.Sprintf("%v", val))
	case nil:
		out.WriteString(c.typeColors.NullColor.Sprint("null"))
	default:
		out.WriteString(c.typeColors.StringColor.Sprintf("%v", val))
	}
}
