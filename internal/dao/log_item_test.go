package dao_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestLogItemsFilter(t *testing.T) {
	uu := map[string]struct {
		q    string
		opts dao.LogOptions
		e    []int
		err  error
	}{
		"empty": {
			opts: dao.LogOptions{},
		},
		"pod-name": {
			q: "blee",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{0, 1, 2},
		},
		"container-name": {
			q: "c1",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{0, 1, 2},
		},
		"message": {
			q: "zorg",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{2},
		},
		"fuzzy": {
			q: "-f zorg",
			opts: dao.LogOptions{
				Path:      "fred/blee",
				Container: "c1",
			},
			e: []int{2},
		},
	}

	for k := range uu {
		u := uu[k]
		ii := dao.LogItems{
			dao.NewLogItem([]byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))),
			dao.NewLogItemFromString("Bumble bee tuna"),
			dao.NewLogItemFromString("Jean Batiste Emmanuel Zorg"),
		}
		t.Run(k, func(t *testing.T) {
			_, n := client.Namespaced(u.opts.Path)
			for _, i := range ii {
				i.Pod, i.Container = n, u.opts.Container
			}
			res, _, err := ii.Filter(u.q, false)
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, u.e, res)
			}
		})
	}
}

func TestLogItemsRender(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		e    string
	}{
		"empty": {
			opts: dao.LogOptions{},
			e:    "Testing 1,2,3...",
		},
		"container": {
			opts: dao.LogOptions{
				Container: "fred",
			},
			e: "\x1b[38;5;161mfred\x1b[0m Testing 1,2,3...",
		},
		"pod": {
			opts: dao.LogOptions{
				Path:      "blee/fred",
				Container: "blee",
			},
			e: "\x1b[38;5;161mfred\x1b[0m:\x1b[38;5;161mblee\x1b[0m Testing 1,2,3...",
		},
		"full": {
			opts: dao.LogOptions{
				Path:          "blee/fred",
				Container:     "blee",
				ShowTimestamp: true,
			},
			e: "\x1b[38;5;106m2018-12-14T10:36:43.326972-07:00\x1b[0m \x1b[38;5;161mfred\x1b[0m:\x1b[38;5;161mblee\x1b[0m Testing 1,2,3...",
		},
	}

	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	ii := dao.LogItems{dao.NewLogItem(s)}
	for k := range uu {
		u := uu[k]
		_, n := client.Namespaced(u.opts.Path)
		ii[0].Pod, ii[0].Container = n, u.opts.Container
		t.Run(k, func(t *testing.T) {
			res := make([][]byte, 1)
			ii.Render(u.opts.ShowTimestamp, res)
			assert.Equal(t, u.e, string(res[0]))
		})
	}
}

func TestLogItemEmpty(t *testing.T) {
	uu := map[string]struct {
		s string
		e bool
	}{
		"empty": {s: "", e: true},
		"full":  {s: "Testing 1,2,3..."},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			i := dao.NewLogItemFromString(u.s)
			assert.Equal(t, u.e, i.IsEmpty())
		})
	}
}

func TestLogItemRender(t *testing.T) {
	uu := map[string]struct {
		opts dao.LogOptions
		e    string
	}{
		"empty": {
			opts: dao.LogOptions{},
			e:    "Testing 1,2,3...",
		},
		"container": {
			opts: dao.LogOptions{
				Container: "fred",
			},
			e: "\x1b[38;5;0mfred\x1b[0m Testing 1,2,3...",
		},
		"pod": {
			opts: dao.LogOptions{
				Path:      "blee/fred",
				Container: "blee",
			},
			e: "\x1b[38;5;0mfred\x1b[0m:\x1b[38;5;0mblee\x1b[0m Testing 1,2,3...",
		},
		"full": {
			opts: dao.LogOptions{
				Path:          "blee/fred",
				Container:     "blee",
				ShowTimestamp: true,
			},
			e: "\x1b[38;5;106m2018-12-14T10:36:43.326972-07:00\x1b[0m \x1b[38;5;0mfred\x1b[0m:\x1b[38;5;0mblee\x1b[0m Testing 1,2,3...",
		},
	}

	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			i := dao.NewLogItem(s)
			_, n := client.Namespaced(u.opts.Path)
			i.Pod, i.Container = n, u.opts.Container

			assert.Equal(t, u.e, string(i.Render(0, u.opts.ShowTimestamp)))
		})
	}
}

func BenchmarkLogItemRender(b *testing.B) {
	s := []byte(fmt.Sprintf("%s %s\n", "2018-12-14T10:36:43.326972-07:00", "Testing 1,2,3..."))
	i := dao.NewLogItem(s)
	i.Pod, i.Container = "fred", "blee"
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		i.Render(0, true)
	}
}
