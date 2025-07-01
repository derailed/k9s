// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvReplace(t *testing.T) {
	uu := map[string]struct {
		arg string
		err error
		e   string
	}{
		"no-args":   {arg: "blee blah", e: "blee blah"},
		"simple":    {arg: "$A", e: "10"},
		"substring": {arg: "$A and $AA", e: "10 and 20"},
		"with-text": {arg: "Something $A", e: "Something 10"},
		"noMatch":   {arg: "blah blah and $BLEE", e: "blah blah and $BLEE"},
		"lower":     {arg: "And then $b happened", e: "And then blee happened"},
		"dash":      {arg: "$col0", e: "fred"},
		"underline": {arg: "$RESOURCE_GROUP", e: "foo"},
		"mix":       {arg: "$col0 and then $a but $B", e: "fred and then 10 but blee"},
		"subs":      {arg: `{"spec" : {"suspend" : $COL0 }}`, e: `{"spec" : {"suspend" : fred }}`},
		"boolean":   {arg: "$COL-BOOL", e: "false"},
		"invert":    {arg: "$!COL-BOOL", e: "true"},

		"simple_braces":    {arg: "${A}", e: "10"},
		"embed_braces":     {arg: "blabla${A}blabla", e: "blabla10blabla"},
		"open_braces":      {arg: "${A", e: "${A"},
		"closed_braces":    {arg: "$A}", e: "10}"},
		"substring_braces": {arg: "${A} and ${AA}", e: "10 and 20"},
		"with-text_braces": {arg: "Something ${A}", e: "Something 10"},
		"noMatch_braces":   {arg: "blah blah and ${BLEE}", e: "blah blah and ${BLEE}"},
		"lower_braces":     {arg: "And then ${b} happened", e: "And then blee happened"},
		"dash_braces":      {arg: "${col0}", e: "fred"},
		"underline_braces": {arg: "${RESOURCE_GROUP}", e: "foo"},
		"mix_braces":       {arg: "${col0} and then ${a} but ${B}", e: "fred and then 10 but blee"},
		"subs_braces":      {arg: `{"spec" : {"suspend" : ${COL0} }}`, e: `{"spec" : {"suspend" : fred }}`},
		"boolean_braces":   {arg: "${COL-BOOL}", e: "false"},
		"invert_braces":    {arg: "${!COL-BOOL}", e: "true"},
		"special_braces":   {arg: "${COL-%CPU/L}/${COL-MEM/R:L}", e: "10/32:32"},
		"space_braces":     {arg: "${READINESS GATES}", e: "bar"},
	}

	e := Env{
		"A":               "10",
		"AA":              "20",
		"B":               "blee",
		"COL0":            "fred",
		"FRED":            "fred",
		"COL-NAME":        "zorg",
		"COL-BOOL":        "false",
		"COL-%CPU/L":      "10",
		"COL-MEM/R:L":     "32:32",
		"RESOURCE_GROUP":  "foo",
		"READINESS GATES": "bar",
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			a, err := e.Substitute(u.arg)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, a)
		})
	}
}
