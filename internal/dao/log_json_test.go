// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyJSON(t *testing.T) {
	uu := map[string]struct {
		msg string
		e   string
		ok  bool
	}{
		"empty": {
			msg: "",
		},
		"plain": {
			msg: "Server listening on: [::]:5000\n",
		},
		"invalid": {
			msg: `{"level":"info","msg":` + "\n",
		},
		"brace-prefix-not-json": {
			msg: "{not json}\n",
		},
		"bracket-log-level": {
			msg: "[INFO] [main] started\n",
		},
		"bracket-log-level-ends-with-bracket": {
			msg: "[INFO] started [main]\n",
		},
		"empty-array": {
			msg: "[]",
			e:   "[]\n",
			ok:  true,
		},
		"array-of-objects": {
			msg: `[ {"a":1} ]`,
			e:   "[\n  {\n    \"a\": 1\n  }\n]\n",
			ok:  true,
		},
		"text-before-json": {
			msg: `INFO {"a":1}` + "\n",
		},
		"object": {
			msg: `{"level":"info","msg":"hello"}` + "\n",
			e:   "{\n  \"level\": \"info\",\n  \"msg\": \"hello\"\n}\n",
			ok:  true,
		},
		"keeps-key-order": {
			msg: `{"z":1,"a":2}`,
			e:   "{\n  \"z\": 1,\n  \"a\": 2\n}\n",
			ok:  true,
		},
		"keeps-number-precision": {
			msg: `{"id":1234567890123456789}`,
			e:   "{\n  \"id\": 1234567890123456789\n}\n",
			ok:  true,
		},
		"array": {
			msg: `["a","b"]`,
			e:   "[\n  \"a\",\n  \"b\"\n]\n",
			ok:  true,
		},
		"nested": {
			msg: `  {"a":{"b":[1,2]}}  ` + "\r\n",
			e:   "{\n  \"a\": {\n    \"b\": [\n      1,\n      2\n    ]\n  }\n}\n",
			ok:  true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			res, ok := prettyJSON([]byte(u.msg))
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.e, string(res))
			}
		})
	}
}

func TestNewRawLogItem(t *testing.T) {
	const ts = "2018-12-14T10:36:43.326972-07:00 "
	uu := map[string]struct {
		raw, bytes, rawJSON string
	}{
		"plain": {
			raw:   ts + "[INFO] started\n",
			bytes: ts + "[INFO[] started\n",
		},
		"json": {
			raw:     ts + `{"tags":["a"]}` + "\n",
			bytes:   ts + `{"tags":["a"[]}` + "\n",
			rawJSON: `{"tags":["a"]}` + "\n",
		},
		"json-no-timestamp": {
			raw:     `{"a":1}`,
			bytes:   `{"a":1}`,
			rawJSON: `{"a":1}`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			i := NewRawLogItem([]byte(u.raw))
			assert.Equal(t, u.bytes, string(i.Bytes))
			// Only JSON looking lines keep a raw copy around.
			assert.Equal(t, u.rawJSON, string(i.rawJSON))
			assert.Equal(t, u.rawJSON == "", i.rawJSON == nil)
		})
	}
}
