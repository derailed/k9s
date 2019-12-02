package render

// BOZO!!
// type (
// 	colorerUC struct {
// 		ns string
// 		r  RowEvent
// 		e  tcell.Color
// 	}
// 	colorerUCs []colorerUC
// )

// func TestNSColorer(t *testing.T) {
// 	var (
// 		ns   = Row{Fields: Fields{"blee", "Active"}}
// 		term = Row{Fields: Fields{"blee", Terminating}}
// 		dead = Row{Fields: Fields{"blee", "Inactive"}}
// 	)

// 	uu := colorerUCs{
// 		// Add AllNS
// 		{"", RowEvent{
// 			Kind: EventAdd,
// 			Row:  ns,
// 		},
// 			AddColor},
// 		// Mod AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: ns}, ModColor},
// 		// MoChange AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: ns}, StdColor},
// 		// Bust NS
// 		{"", RowEvent{Kind: EventUnchanged, Row: term}, ErrColor},
// 		// Bust NS
// 		{"", RowEvent{Kind: EventUnchanged, Row: dead}, ErrColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, nsColorer(u.ns, u.r))
// 	}
// }

// func TestEvColorer(t *testing.T) {
// 	var (
// 		ns       = Row{Fields: Fields{"", "blee", "fred", "Normal"}}
// 		nonNS    = Row{Fields: Fields{"", "fred", "Normal"}}
// 		failNS   = Row{Fields: Fields{"", "blee", "fred", "Failed"}}
// 		failNoNS = Row{Fields: Fields{"", "fred", "Failed"}}
// 		killNS   = Row{Fields: Fields{"", "blee", "fred", "Killing"}}
// 		killNoNS = Row{Fields: Fields{"", "fred", "Killing"}}
// 	)

// 	uu := colorerUCs{
// 		// Add AllNS
// 		{"", RowEvent{Kind: EventAdd, Row: ns}, AddColor},
// 		// Add NS
// 		{"blee", RowEvent{Kind: EventAdd, Row: nonNS}, AddColor},
// 		// Mod AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: ns}, ModColor},
// 		// Mod NS
// 		{"blee", RowEvent{Kind: EventUpdate, Row: nonNS}, ModColor},
// 		// Bust AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: failNS}, ErrColor},
// 		// Bust NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: failNoNS}, ErrColor},
// 		// Bust AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: killNS}, KillColor},
// 		// Bust NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: killNoNS}, KillColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, evColorer(u.ns, u.r))
// 	}
// }

// func TestRSColorer(t *testing.T) {
// 	var (
// 		ns       = Row{Fields: Fields{"blee", "fred", "1", "1"}}
// 		noNs     = Row{Fields: Fields{"fred", "1", "1"}}
// 		bustNS   = Row{Fields: Fields{"blee", "fred", "1", "0"}}
// 		bustNoNS = Row{Fields: Fields{"fred", "1", "0"}}
// 	)

// 	uu := colorerUCs{
// 		// Add AllNS
// 		{"", RowEvent{Kind: EventAdd, Row: ns}, AddColor},
// 		// Add NS
// 		{"blee", RowEvent{Kind: EventAdd, Row: noNs}, AddColor},
// 		// Bust AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: bustNS}, ErrColor},
// 		// Bust NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: bustNoNS}, ErrColor},
// 		// Nochange AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: ns}, StdColor},
// 		// Nochange NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: noNs}, StdColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, rsColorer(u.ns, u.r))
// 	}
// }

// func TestStsColorer(t *testing.T) {
// 	var (
// 		ns       = Row{Fields: Fields{"blee", "fred", "1", "1"}}
// 		nonNS    = Row{Fields: Fields{"fred", "1", "1"}}
// 		bustNS   = Row{Fields: Fields{"blee", "fred", "2", "1"}}
// 		bustNoNS = Row{Fields: Fields{"fred", "2", "1"}}
// 	)

// 	uu := colorerUCs{
// 		// Add AllNS
// 		{"", RowEvent{Kind: EventAdd, Row: ns}, AddColor},
// 		// Add NS
// 		{"blee", RowEvent{Kind: EventAdd, Row: nonNS}, AddColor},
// 		// Mod AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: ns}, ModColor},
// 		// Mod NS
// 		{"blee", RowEvent{Kind: EventUpdate, Row: nonNS}, ModColor},
// 		// Bust AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: bustNS}, ErrColor},
// 		// Bust NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: bustNoNS}, ErrColor},
// 		// Unchanged cool AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: ns}, StdColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, stsColorer(u.ns, u.r))
// 	}
// }

// func TestDpColorer(t *testing.T) {
// 	var (
// 		ns       = Row{Fields: Fields{"blee", "fred", "1", "1"}}
// 		nonNS    = Row{Fields: Fields{"fred", "1", "1"}}
// 		bustNS   = Row{Fields: Fields{"blee", "fred", "2", "1"}}
// 		bustNoNS = Row{Fields: Fields{"fred", "2", "1"}}
// 	)

// 	uu := colorerUCs{
// 		// Add AllNS
// 		{"", RowEvent{Kind: EventAdd, Row: ns}, AddColor},
// 		// Add NS
// 		{"blee", RowEvent{Kind: EventAdd, Row: nonNS}, AddColor},
// 		// Mod AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: ns}, ModColor},
// 		// Mod NS
// 		{"blee", RowEvent{Kind: EventUpdate, Row: nonNS}, ModColor},
// 		// Unchanged cool
// 		{"", RowEvent{Kind: EventUnchanged, Row: ns}, StdColor},
// 		// Bust AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: bustNS}, ErrColor},
// 		// Bust NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: bustNoNS}, ErrColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, dpColorer(u.ns, u.r))
// 	}
// }

// func TestPdbColorer(t *testing.T) {
// 	var (
// 		ns       = Row{Fields: Fields{"blee", "fred", "1", "1", "1", "1", "1"}}
// 		nonNS    = Row{Fields: Fields{"fred", "1", "1", "1", "1", "1"}}
// 		bustNS   = Row{Fields: Fields{"blee", "fred", "1", "1", "1", "1", "2"}}
// 		bustNoNS = Row{Fields: Fields{"fred", "1", "1", "1", "1", "2"}}
// 	)

// 	uu := colorerUCs{
// 		// Add AllNS
// 		{"", RowEvent{Kind: EventAdd, Row: ns}, AddColor},
// 		// Add NS
// 		{"blee", RowEvent{Kind: EventAdd, Row: nonNS}, AddColor},
// 		// Mod AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: ns}, ModColor},
// 		// Mod NS
// 		{"blee", RowEvent{Kind: EventUpdate, Row: nonNS}, ModColor},
// 		// Unchanged cool
// 		{"", RowEvent{Kind: EventUnchanged, Row: ns}, StdColor},
// 		// Bust AllNS
// 		{"", RowEvent{Kind: EventUnchanged, Row: bustNS}, ErrColor},
// 		// Bust NS
// 		{"blee", RowEvent{Kind: EventUnchanged, Row: bustNoNS}, ErrColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, pdbColorer(u.ns, u.r))
// 	}
// }

// func TestPVColorer(t *testing.T) {
// 	var (
// 		pv     = Row{Fields: Fields{"blee", "1G", "RO", "Duh", "Bound"}}
// 		bustPv = Row{Fields: Fields{"blee", "1G", "RO", "Duh", "UnBound"}}
// 	)

// 	uu := colorerUCs{
// 		// Add Normal
// 		{"", RowEvent{Kind: EventAdd, Row: pv}, AddColor},
// 		// Unchanged Bound
// 		{"", RowEvent{Kind: EventUnchanged, Row: pv}, StdColor},
// 		// Unchanged Bound
// 		{"", RowEvent{Kind: EventUnchanged, Row: bustPv}, ErrColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, pvColorer(u.ns, u.r))
// 	}
// }

// func TestPVCColorer(t *testing.T) {
// 	var (
// 		pvc     = Row{Fields: Fields{"blee", "fred", "Bound"}}
// 		bustPvc = Row{Fields: Fields{"blee", "fred", "UnBound"}}
// 	)

// 	uu := colorerUCs{
// 		// Add Normal
// 		{"", RowEvent{Kind: EventAdd, Row: pvc}, AddColor},
// 		// Add Bound
// 		{"", RowEvent{Kind: EventUnchanged, Row: bustPvc}, ErrColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, pvcColorer(u.ns, u.r))
// 	}
// }

// func TestCtxColorer(t *testing.T) {
// 	var (
// 		ctx    = Row{Fields: Fields{"blee"}}
// 		defCtx = Row{Fields: Fields{"blee*"}}
// 	)

// 	uu := colorerUCs{
// 		// Add Normal
// 		{"", RowEvent{Kind: EventAdd, Row: ctx}, AddColor},
// 		// Add Default
// 		{"", RowEvent{Kind: EventAdd, Row: defCtx}, AddColor},
// 		// Mod Normal
// 		{"", RowEvent{Kind: EventUpdate, Row: ctx}, ModColor},
// 		// Mod Default
// 		{"", RowEvent{Kind: EventUpdate, Row: defCtx}, ModColor},
// 		// Unchanged Normal
// 		{"", RowEvent{Kind: EventUnchanged, Row: ctx}, StdColor},
// 		// Unchanged Default
// 		{"", RowEvent{Kind: EventUnchanged, Row: defCtx}, HighlightColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, ctxColorer(u.ns, u.r))
// 	}
// }

// func TestPodColorer(t *testing.T) {
// 	var (
// 		nsRow      = Row{Fields: Fields{"blee", "fred", "1/1", "Running"}}
// 		toastNS    = Row{Fields: Fields{"blee", "fred", "1/1", "Boom"}}
// 		notReadyNS = Row{Fields: Fields{"blee", "fred", "0/1", "Boom"}}
// 		row        = Row{Fields: Fields{"fred", "1/1", "Running"}}
// 		toast      = Row{Fields: Fields{"fred", "1/1", "Boom"}}
// 		notReady   = Row{Fields: Fields{"fred", "0/1", "Boom"}}
// 	)

// 	uu := colorerUCs{
// 		// Add allNS
// 		{"", RowEvent{Kind: EventAdd, Row: nsRow}, AddColor},
// 		// Add Namespaced
// 		{"blee", RowEvent{Kind: EventAdd, Row: row}, AddColor},
// 		// Mod AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: nsRow}, ModColor},
// 		// Mod Namespaced
// 		{"blee", RowEvent{Kind: EventUpdate, Row: row}, ModColor},
// 		// Mod Busted AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: toastNS}, ErrColor},
// 		// Mod Busted Namespaced
// 		{"blee", RowEvent{Kind: EventUpdate, Row: toast}, ErrColor},
// 		// NotReady AllNS
// 		{"", RowEvent{Kind: EventUpdate, Row: notReadyNS}, ErrColor},
// 		// NotReady Namespaced
// 		{"blee", RowEvent{Kind: EventUpdate, Row: notReady}, ErrColor},
// 	}
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, podColorer(u.ns, u.r))
// 	}
// }
