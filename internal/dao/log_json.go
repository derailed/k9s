// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"bytes"
	"encoding/json"
)

const jsonIndent = "  "

// jsonValueStart lists the bytes a JSON array element (or its end) may start with.
var jsonValueStart = []byte(`{["-0123456789tfn]`)

// prettyJSON indents a raw (unescaped) log message when it holds a valid JSON
// object or array. It returns false when the message is not JSON.
// json.Indent is used over Unmarshal/MarshalIndent so that key order and
// number precision are preserved.
func prettyJSON(msg []byte) ([]byte, bool) {
	b := bytes.TrimSpace(msg)
	if !maybeJSON(b) || !json.Valid(b) {
		return nil, false
	}

	var out bytes.Buffer
	out.Grow(len(b) * 2)
	if err := json.Indent(&out, b, "", jsonIndent); err != nil {
		return nil, false
	}
	out.WriteByte('\n')

	return out.Bytes(), true
}

// maybeJSON cheaply weeds out lines that can't be a JSON object or array,
// e.g. "[INFO] ...", so that most plain log lines skip the JSON validation.
func maybeJSON(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	switch b[0] {
	case '{':
		return b[len(b)-1] == '}'
	case '[':
		if b[len(b)-1] != ']' {
			return false
		}
		v := bytes.TrimLeft(b[1:], " \t\r\n")
		return len(v) > 0 && bytes.IndexByte(jsonValueStart, v[0]) >= 0
	default:
		return false
	}
}
