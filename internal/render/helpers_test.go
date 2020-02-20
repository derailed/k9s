package render

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestToMB(t *testing.T) {
	uu := []struct {
		v int64
		e float64
	}{
		{0, 0},
		{2 * megaByte, 2},
		{10 * megaByte, 10},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, ToMB(u.v))
	}
}

func TestToPerc(t *testing.T) {
	uu := []struct {
		v1, v2, e float64
	}{
		{0, 0, 0},
		{100, 200, 50},
		{200, 100, 200},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, toPerc(u.v1, u.v2))
	}
}

func TestToAge(t *testing.T) {
	uu := map[string]struct {
		t time.Time
		e string
	}{
		"good": {
			t: time.Now().Add(-10 * time.Second),
			e: "10",
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, toAge(metav1.Time{Time: uc.t})[:2])
		})
	}
}

func TestToAgeHuma(t *testing.T) {
	uu := map[string]struct {
		t time.Time
		e string
	}{
		"good": {
			t: time.Now().Add(-10 * time.Second),
			e: "10",
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			ti := toAge(metav1.Time{Time: uc.t})
			assert.Equal(t, uc.e, toAgeHuman(ti)[:2])
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

func TestIn(t *testing.T) {
	uu := map[string]struct {
		a []string
		v string
		e bool
	}{
		"in": {
			a: []string{"fred", "blee"},
			v: "blee",
			e: true,
		},
		"empty": {
			v: "blee",
		},
		"missing": {
			a: []string{"fred", "blee"},
			v: "duh",
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, in(uc.a, uc.v))
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
