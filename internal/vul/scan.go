// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"fmt"
	"io"
	"strings"

	grypeDB "github.com/anchore/grype/grype/db/v5"
	"github.com/anchore/grype/grype/match"
	"github.com/anchore/grype/grype/vulnerability"
)

const (
	wontFix = "(won't fix)"
	naValue = ""
)

// Scans tracks scans per image.
type Scans map[string]*Scan

// Dump dump reports to stdout.
func (s Scans) Dump(w io.Writer) {
	for k, v := range s {
		fmt.Fprintf(w, "Image: %s -- ", k)
		v.Tally.Dump(w)
		fmt.Fprintln(w)
		v.Dump(w)
	}
}

// Scan tracks image vulnerability scan.
type Scan struct {
	ID    string
	Table *table
	Tally tally
}

func newScan(img string) *Scan {
	return &Scan{ID: img, Table: newTable()}
}

// Dump dump report to stdout.
func (s *Scan) Dump(w io.Writer) {
	s.Table.dump(w)
}

func (s *Scan) run(mm *match.Matches, store vulnerability.MetadataProvider) error {
	for m := range mm.Enumerate() {
		meta, err := store.GetMetadata(m.Vulnerability.ID, m.Vulnerability.Namespace)
		if err != nil {
			return err
		}
		var severity string
		if meta != nil {
			severity = meta.Severity
		}
		fixVersion := strings.Join(m.Vulnerability.Fix.Versions, ", ")
		switch m.Vulnerability.Fix.State {
		case grypeDB.WontFixState:
			fixVersion = wontFix
		case grypeDB.UnknownFixState:
			fixVersion = naValue
		}
		s.Table.addRow(newRow(m.Package.Name, m.Package.Version, fixVersion, string(m.Package.Type), m.Vulnerability.ID, severity))
	}
	s.Table.dedup()
	s.Tally = newTally(s.Table)

	return nil
}

func colorize(rr []string) []string {
	crr := make([]string, len(rr))
	copy(crr, rr)

	crr[len(crr)-1] = sevColor(crr[len(crr)-1])
	return crr
}
