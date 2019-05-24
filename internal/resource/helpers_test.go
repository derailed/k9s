package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	uu := map[string]struct {
		i []string
		e string
	}{
		"zero":   {[]string{}, ""},
		"std":    {[]string{"a", "b", "c"}, "a,b,c"},
		"blank":  {[]string{"", "", ""}, ""},
		"sparse": {[]string{"a", "", "c"}, "a,c"},
	}

	for k, v := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, v.e, join(v.i, ","))
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
		ns, n := namespaced(u.p)
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
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 3, "fr…"},
		{"fred", 2, "f…"},
		{"fred", 10, "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, Truncate(u.s, u.l))
	}
}

func TestMapToStr(t *testing.T) {
	uu := []struct {
		i map[string]string
		e string
	}{
		{map[string]string{"blee": "duh", "aa": "bb"}, "aa=bb,blee=duh"},
		{map[string]string{}, MissingValue},
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
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mapToStr(ll)
	}
}

func TestToMillicore(t *testing.T) {
	uu := []struct {
		v int64
		e string
	}{
		{0, "0"},
		{2, "2"},
		{1000, "1000"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, ToMillicore(u.v))
	}
}

func TestToMi(t *testing.T) {
	uu := []struct {
		v float64
		e string
	}{
		{0, "0"},
		{2, "2"},
		{1000, "1000"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, ToMi(u.v))
	}
}

func TestAsPerc(t *testing.T) {
	uu := []struct {
		v float64
		e string
	}{
		{0, "0"},
		{10.5, "10"},
		{10, "10"},
		{0.05, "0"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, AsPerc(u.v))
	}
}

func BenchmarkAsPerc(b *testing.B) {
	v := 10.5
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		AsPerc(v)
	}
}
