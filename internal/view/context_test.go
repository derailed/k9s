package view_test

// import (
// 	"testing"

// 	"github.com/derailed/k9s/internal/config"
// 	"github.com/derailed/k9s/internal/resource"
// 	"github.com/derailed/k9s/internal/view"
// 	"github.com/stretchr/testify/assert"
// )

// func TestContext(t *testing.T) {
// 	l := resource.NewContextList(nil, "fred")
// 	v := view.NewContext("blee", "", NewApp(config.NewConfig(ks{})), l).(*contextView)

// 	assert.Equal(t, 10, len(v.Hints()))
// }

// func TestCleaner(t *testing.T) {
// 	uu := map[string]struct {
// 		s, e string
// 	}{
// 		"normal":  {"fred", "fred"},
// 		"default": {"fred*", "fred"},
// 		"delta":   {"fred(𝜟)", "fred"},
// 	}

// 	v := contextView{}
// 	for k, u := range uu {
// 		t.Run(k, func(t *testing.T) {
// 			assert.Equal(t, u.e, v.cleanser(u.s))
// 		})
// 	}
// }
