package render

import (
	"fmt"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSortLabels(t *testing.T) {
	uu := map[string]struct {
		labels string
		e      [][]string
	}{
		"simple": {
			labels: "a=b,c=d",
			e: [][]string{
				{"a", "c"},
				{"b", "d"},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			hh, vv := sortLabels(labelize(u.labels))
			assert.Equal(t, u.e[0], hh)
			assert.Equal(t, u.e[1], vv)
		})
	}
}

func TestLabelize(t *testing.T) {
	uu := map[string]struct {
		labels string
		e      map[string]string
	}{
		"simple": {
			labels: "a=b,c=d",
			e:      map[string]string{"a": "b", "c": "d"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, labelize(u.labels))
		})
	}
}

func TestDurationToSecond(t *testing.T) {
	uu := map[string]struct {
		s string
		e int64
	}{
		"seconds":                 {s: "22s", e: 22},
		"minutes":                 {s: "22m", e: 1320},
		"hours":                   {s: "12h", e: 43200},
		"days":                    {s: "3d", e: 259200},
		"day_hour":                {s: "3d9h", e: 291600},
		"day_hour_minute":         {s: "2d22h3m", e: 252180},
		"day_hour_minute_seconds": {s: "2d22h3m50s", e: 252230},
		"year":                    {s: "3y", e: 94608000},
		"year_day":                {s: "1y2d", e: 31708800},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, durationToSeconds(u.s))
		})
	}
}

func BenchmarkDurationToSecond(b *testing.B) {
	t := "2d22h3m50s"

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		durationToSeconds(t)
	}
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
			assert.Equal(t, uc.e, toAge(metav1.Time{Time: uc.t}))
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

// Helpers

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		fmt.Println("TestTime Failed", err)
	}
	return t
}
