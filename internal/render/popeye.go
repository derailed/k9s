// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import "github.com/derailed/popeye/pkg/config"

// !!BOZO!! Popeye

// // Popeye renders a sanitizer to screen.
// type Popeye struct {
// 	Base
// }

// // ColorerFunc colors a resource row.
// func (Popeye) ColorerFunc() ColorerFunc {
// 	return func(ns string, h Header, re *model1.RowEvent) tcell.Color {
// 		c := DefaultColorer(ns, h, re)

// 		warnCol := h.IndexOf("WARNING", true)
// 		status, _ := strconv.Atoi(strings.TrimSpace(re.Row.Fields[warnCol]))
// 		if status > 0 {
// 			c = tcell.ColorOrange
// 		}
// 		errCol := h.IndexOf("ERROR", true)
// 		status, _ = strconv.Atoi(strings.TrimSpace(re.Row.Fields[errCol]))
// 		if status > 0 {
// 			c = ErrColor
// 		}
// 		return c
// 	}
// }

// // Header returns a header row.
// func (Popeye) Header(ns string) model1.Header {
// 	return model1.Header{
// 		model1.HeaderColumn{Name: "RESOURCE"},
// 		model1.HeaderColumn{Name: "SCORE%", Align: tview.AlignRight},
// 		model1.HeaderColumn{Name: "SCANNED", Align: tview.AlignRight},
// 		model1.HeaderColumn{Name: "ERROR", Align: tview.AlignRight},
// 		model1.HeaderColumn{Name: "WARNING", Align: tview.AlignRight},
// 		model1.HeaderColumn{Name: "INFO", Align: tview.AlignRight},
// 		model1.HeaderColumn{Name: "OK", Align: tview.AlignRight},
// 	}
// }

// // Render renders a K8s resource to screen.
// func (Popeye) Render(o interface{}, ns string, r *model1.Row) error {
// 	s, ok := o.(Section)
// 	if !ok {
// 		return fmt.Errorf("expected Section, but got %T", o)
// 	}

// 	r.ID = client.FQN(ns, s.Title)
// 	r.Fields = append(r.Fields,
// 		s.Title,
// 		strconv.Itoa(s.Tally.Score()),
// 		strconv.Itoa(s.Tally.OK+s.Tally.Info+s.Tally.Warning+s.Tally.Error),
// 		strconv.Itoa(s.Tally.Error),
// 		strconv.Itoa(s.Tally.Warning),
// 		strconv.Itoa(s.Tally.Info),
// 		strconv.Itoa(s.Tally.OK),
// 	)
// 	return nil
// }

// // ----------------------------------------------------------------------------
// // Helpers...

type (
	// 	// Builder represents a popeye report.
	// 	Builder struct {
	// 		Report Report `json:"popeye" yaml:"popeye"`
	// 	}

	// 	// Report represents the output of a sanitization pass.
	// 	Report struct {
	// 		Score    int      `json:"score" yaml:"score"`
	// 		Grade    string   `json:"grade" yaml:"grade"`
	// 		Sections Sections `json:"sanitizers,omitempty" yaml:"sanitizers,omitempty"`
	// 	}

	// Sections represents a collection of sections.
	Sections []Section

	// Section represents a sanitizer pass.
	Section struct {
		Title   string  `json:"sanitizer" yaml:"sanitizer"`
		GVR     string  `yaml:"gvr" json:"gvr"`
		Tally   *Tally  `json:"tally" yaml:"tally"`
		Outcome Outcome `json:"issues,omitempty" yaml:"issues,omitempty"`
	}

	// Outcome represents a classification of reports outcome.
	Outcome map[string]Issues

	// Issues represents a collection of issues.
	Issues []Issue

	// Issue represents a sanitization issue.
	Issue struct {
		Group   string       `yaml:"group" json:"group"`
		GVR     string       `yaml:"gvr" json:"gvr"`
		Level   config.Level `yaml:"level" json:"level"`
		Message string       `yaml:"message" json:"message"`
	}

	// Tally tracks a section scores.

	Tally struct {
		OK, Info, Warning, Error int
		Count                    int
	}
)

// // Sum sums up tally counts.
// func (t *Tally) Sum() int {
// 	return t.OK + t.Info + t.Warning + t.Error
// }

// // Score returns the overall sections score in percent.
// func (t *Tally) Score() int {
// 	oks := t.OK + t.Info
// 	return toPerc(float64(oks), float64(oks+t.Warning+t.Error))
// }

// func toPerc(v1, v2 float64) int {
// 	if v2 == 0 {
// 		return 0
// 	}
// 	return int(math.Floor((v1 / v2) * 100))
// }

// // Len returns a section length.
// func (s Sections) Len() int {
// 	return len(s)
// }

// // Swap swaps values.
// func (s Sections) Swap(i, j int) {
// 	s[i], s[j] = s[j], s[i]
// }

// // Less compares section scores.
// func (s Sections) Less(i, j int) bool {
// 	t1, t2 := s[i].Tally, s[j].Tally
// 	return t1.Score() < t2.Score()
// }

// // GetObjectKind returns a schema object.
// func (Section) GetObjectKind() schema.ObjectKind {
// 	return nil
// }

// // DeepCopyObject returns a container copy.
// func (s Section) DeepCopyObject() runtime.Object {
// 	return s
// }

// // MaxSeverity gather the max severity in a collection of issues.
// func (s Section) MaxSeverity() config.Level {
// 	max := config.OkLevel
// 	for _, issues := range s.Outcome {
// 		m := issues.MaxSeverity()
// 		if m > max {
// 			max = m
// 		}
// 	}

// 	return max
// }

// // MaxSeverity gather the max severity in a collection of issues.
// func (i Issues) MaxSeverity() config.Level {
// 	max := config.OkLevel
// 	for _, is := range i {
// 		if is.Level > max {
// 			max = is.Level
// 		}
// 	}

// 	return max
// }

// // CountSeverity counts severity level instances.
// func (i Issues) CountSeverity(l config.Level) int {
// 	var count int
// 	for _, is := range i {
// 		if is.Level == l {
// 			count++
// 		}
// 	}

// 	return count
// }
