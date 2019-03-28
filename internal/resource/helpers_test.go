package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestPad(t *testing.T) {
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 10, "fred      "},
		{"fred", 6, "fred  "},
		{"fred", 4, "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, Pad(u.s, u.l))
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

func TestSizeCol(t *testing.T) {
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 3, "fr…"},
		{"01234567890", 10, "012345678…"},
		{"fred", 10, "fred      "},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, FixCol(u.s, u.l))
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
		{0, "0m"},
		{2, "2m"},
		{1000, "1000m"},
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
		{0, "0Mi"},
		{2, "2Mi"},
		{1000, "1000Mi"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, ToMi(u.v))
	}
}
