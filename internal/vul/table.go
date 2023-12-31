// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
)

const (
	nameIdx = iota
	verIdx
	fixIdx
	typeIdx
	vulIdx
	sevIdx
)

type Row []string

func newRow(ss ...string) Row {
	r := make(Row, 0, len(ss))
	for i, s := range ss {
		if i == sevIdx {
			s = toSev(s)
		}
		r = append(r, s)
	}
	return r
}

func toSev(s string) string {
	switch s {
	case "Critical":
		return Sev1
	case "High":
		return Sev2
	case "Medium":
		return Sev3
	case "Low":
		return Sev4
	case "Negligible":
		return Sev5
	default:
		return SevU
	}
}

func (r Row) Name() string          { return r[nameIdx] }
func (r Row) Version() string       { return r[verIdx] }
func (r Row) Fix() string           { return r[fixIdx] }
func (r Row) Type() string          { return r[typeIdx] }
func (r Row) Vulnerability() string { return r[vulIdx] }
func (r Row) Severity() string      { return r[sevIdx] }

func sevColor(s string) string {
	switch strings.ToLower(s) {
	case "critical":
		return fmt.Sprintf("[red::b]%s[-::-]", s)
	case "high":
		return fmt.Sprintf("[orange::b]%s[-::-]", s)
	case "medium":
		return fmt.Sprintf("[yellow::b]%s[-::-]", s)
	case "low":
		return fmt.Sprintf("[blue::b]%s[-::-]", s)
	default:
		return fmt.Sprintf("[gray::b]%s[-::-]", s)
	}
}

type table struct {
	Rows []Row
}

func newTable() *table {
	return &table{}
}

func (t *table) dedup() {
	var (
		seen = make(map[string]struct{}, len(t.Rows))
		rr   = make([]Row, 0, len(t.Rows))
	)
	for _, v := range t.Rows {
		key := strings.Join(v, "|")
		if _, ok := seen[key]; ok {
			continue
		}
		rr, seen[key] = append(rr, v), struct{}{}
	}
	t.Rows = rr
}

func (t *table) addRow(r Row) {
	t.Rows = append(t.Rows, r)
}

func (t *table) dump(w io.Writer) {
	columns := []string{"Name", "Installed", "Fixed-In", "Type", "Vulnerability", "Severity"}

	table := tablewriter.NewWriter(w)
	table.SetHeader(columns)
	table.SetAutoWrapText(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)

	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(true)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(true)

	for _, row := range t.Rows {
		table.Append(colorize(row))
	}
	table.Render()
}

func (t *table) sort() {
	t.dedup()

	sort.SliceStable(t.Rows, func(i, j int) bool {
		if t.Rows[i][nameIdx] != t.Rows[j][nameIdx] {
			return t.Rows[i][nameIdx] < t.Rows[j][nameIdx]
		}
		if t.Rows[i][verIdx] != t.Rows[j][verIdx] {
			return t.Rows[i][verIdx] < t.Rows[j][verIdx]
		}
		if t.Rows[i][typeIdx] != t.Rows[j][typeIdx] {
			return t.Rows[i][typeIdx] < t.Rows[j][typeIdx]
		}

		if t.Rows[i][sevIdx] == t.Rows[j][sevIdx] {
			return t.Rows[i][vulIdx] < t.Rows[j][vulIdx]
		}
		return sevToScore(t.Rows[i][sevIdx]) < sevToScore(t.Rows[j][sevIdx])
	})
}

func (t *table) sortSev() {
	t.dedup()

	sort.SliceStable(t.Rows, func(i, j int) bool {
		if s1, s2 := sevToScore(t.Rows[i][sevIdx]), sevToScore(t.Rows[j][sevIdx]); s1 != s2 {
			return s1 < s2
		}
		if t.Rows[i][nameIdx] != t.Rows[j][nameIdx] {
			return t.Rows[i][nameIdx] < t.Rows[j][nameIdx]
		}
		if t.Rows[i][verIdx] != t.Rows[j][verIdx] {
			return t.Rows[i][verIdx] < t.Rows[j][verIdx]
		}
		if t.Rows[i][typeIdx] != t.Rows[j][typeIdx] {
			return t.Rows[i][typeIdx] < t.Rows[j][typeIdx]
		}

		return t.Rows[i][vulIdx] < t.Rows[j][vulIdx]
	})
}

func sevToScore(s string) int {
	switch s {
	case Sev1:
		return 1
	case Sev2:
		return 2
	case Sev3:
		return 3
	case Sev4:
		return 4
	case Sev5:
		return 5
	default:
		return 6
	}
}
