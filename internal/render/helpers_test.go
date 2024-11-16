// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTableGenericHydrate(t *testing.T) {
	raw := raw(t, "p1")
	tt := metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "c1"},
			{Name: "c2"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Cells:  []interface{}{"fred", 10},
				Object: runtime.RawExtension{Raw: raw},
			},
			{
				Cells:  []interface{}{"blee", 20},
				Object: runtime.RawExtension{Raw: raw},
			},
		},
	}
	rr := make([]model1.Row, 2)
	var re Generic
	re.SetTable("blee", &tt)

	assert.Nil(t, model1.GenericHydrate("blee", &tt, rr, &re))
	assert.Equal(t, 2, len(rr))
	assert.Equal(t, 3, len(rr[0].Fields))
}

func TestTableHydrate(t *testing.T) {
	oo := []runtime.Object{
		&PodWithMetrics{Raw: load(t, "p1")},
	}
	rr := make([]model1.Row, 1)

	assert.Nil(t, model1.Hydrate("blee", oo, rr, Pod{}))
	assert.Equal(t, 1, len(rr))
	assert.Equal(t, 25, len(rr[0].Fields))
}

func TestToAge(t *testing.T) {
	uu := map[string]struct {
		t time.Time
		e string
	}{
		"zero": {
			t: time.Time{},
			e: UnknownValue,
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, ToAge(metav1.Time{Time: uc.t}))
		})
	}
}

func TestToAgeHuman(t *testing.T) {
	uu := map[string]struct {
		t, e string
	}{
		"blank": {
			t: "",
			e: UnknownValue,
		},
		"good": {
			t: time.Now().Add(-10 * time.Second).Format(time.RFC3339Nano),
			e: "10s",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, toAgeHuman(u.t))
		})
	}
}

func TestJoin(t *testing.T) {
	uu := map[string]struct {
		i []string
		e string
	}{
		"zero":      {[]string{}, ""},
		"std":       {[]string{"a", "b", "c"}, "a,b,c"},
		"blank":     {[]string{"", "", ""}, ""},
		"sparse":    {[]string{"a", "", "c"}, "a,c"},
		"withBlank": {[]string{"", "a", "c"}, "a,c"},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, join(uc.i, ","))
		})
	}
}

func TestBoolPtrToStr(t *testing.T) {
	tv, fv := true, false

	uu := []struct {
		p *bool
		e string
	}{
		{nil, "false"},
		{&tv, "true"},
		{&fv, "false"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, boolPtrToStr(u.p))
	}
}

func TestNamespaced(t *testing.T) {
	uu := []struct {
		p, ns, n string
	}{
		{"fred/blee", "fred", "blee"},
	}

	for _, u := range uu {
		ns, n := client.Namespaced(u.p)
		assert.Equal(t, u.ns, ns)
		assert.Equal(t, u.n, n)
	}
}

func TestMissing(t *testing.T) {
	uu := []struct {
		i, e string
	}{
		{"fred", "fred"},
		{"", MissingValue},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, missing(u.i))
	}
}

func TestBoolToStr(t *testing.T) {
	uu := []struct {
		i bool
		e string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, boolToStr(u.i))
	}
}

func TestNa(t *testing.T) {
	uu := []struct {
		i, e string
	}{
		{"fred", "fred"},
		{"", NAValue},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, na(u.i))
	}
}

func TestTruncate(t *testing.T) {
	uu := map[string]struct {
		data string
		size int
		e    string
	}{
		"same": {
			data: "fred",
			size: 4,
			e:    "fred",
		},
		"small": {
			data: "fred",
			size: 10,
			e:    "fred",
		},
		"larger": {
			data: "fred",
			size: 3,
			e:    "fr…",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, Truncate(u.data, u.size))
		})
	}
}

func TestToSelector(t *testing.T) {
	uu := map[string]struct {
		m map[string]string
		e []string
	}{
		"cool": {
			map[string]string{"app": "fred", "env": "test"},
			[]string{"app=fred,env=test", "env=test,app=fred"},
		},
		"empty": {
			map[string]string{},
			[]string{""},
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			s := toSelector(uc.m)
			var match bool
			for _, e := range uc.e {
				if e == s {
					match = true
				}
			}
			assert.True(t, match)
		})
	}
}

func TestBlank(t *testing.T) {
	uu := map[string]struct {
		a []string
		e bool
	}{
		"full": {
			a: []string{"fred", "blee"},
		},
		"empty": {
			e: true,
		},
		"blank": {
			a: []string{"fred", ""},
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, blank(uc.a))
		})
	}
}

func TestMetaFQN(t *testing.T) {
	uu := map[string]struct {
		m metav1.ObjectMeta
		e string
	}{
		"full": {metav1.ObjectMeta{Namespace: "fred", Name: "blee"}, "fred/blee"},
		"nons": {metav1.ObjectMeta{Name: "blee"}, "-/blee"},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, client.MetaFQN(uc.m))
		})
	}
}

func TestFQN(t *testing.T) {
	uu := map[string]struct {
		ns, n string
		e     string
	}{
		"full": {ns: "fred", n: "blee", e: "fred/blee"},
		"nons": {n: "blee", e: "blee"},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, client.FQN(uc.ns, uc.n))
		})
	}
}

func TestMapToStr(t *testing.T) {
	uu := []struct {
		i map[string]string
		e string
	}{
		{map[string]string{"blee": "duh", "aa": "bb"}, "aa=bb blee=duh"},
		{map[string]string{}, ""},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, mapToStr(u.i))
	}
}

func BenchmarkMapToStr(b *testing.B) {
	ll := map[string]string{
		"blee": "duh",
		"aa":   "bb",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mapToStr(ll)
	}
}

func TestRunesToNum(t *testing.T) {
	uu := map[string]struct {
		rr []rune
		e  int64
	}{
		"0": {
			rr: []rune(""),
			e:  0,
		},
		"100": {
			rr: []rune("100"),
			e:  100,
		},
		"64": {
			rr: []rune("64"),
			e:  64,
		},
		"52640": {
			rr: []rune("52640"),
			e:  52640,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, runesToNum(u.rr))
		})
	}
}

func BenchmarkRunesToNum(b *testing.B) {
	rr := []rune("5465")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runesToNum(rr)
	}
}

func TestToMc(t *testing.T) {
	uu := []struct {
		v int64
		e string
	}{
		{0, "0"},
		{2, "2"},
		{1_000, "1000"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, toMc(u.v))
	}
}

func TestToMi(t *testing.T) {
	uu := []struct {
		v int64
		e string
	}{
		{0, "0"},
		{2 * client.MegaByte, "2"},
		{1_000 * client.MegaByte, "1000"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, toMi(u.v))
	}
}

func TestIntToStr(t *testing.T) {
	uu := []struct {
		v int
		e string
	}{
		{0, "0"},
		{10, "10"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, IntToStr(u.v))
	}
}

func BenchmarkIntToStr(b *testing.B) {
	v := 10
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		IntToStr(v)
	}
}

// Helpers...

func load(t *testing.T, n string) *unstructured.Unstructured {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)
	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)
	return &o
}

func raw(t *testing.T, n string) []byte {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)
	return raw
}
