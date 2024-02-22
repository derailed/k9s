// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"runtime"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

const (
	imgScanTitle = "Scans"
	browseOSX    = "open"
	browseLinux  = "sensible-browser"
	cveGovURL    = "https://nvd.nist.gov/vuln/detail/"
	ghsaURL      = "https://github.com/advisories/"
)

// ImageScan represents an image vulnerability scan view.
type ImageScan struct {
	ResourceViewer
}

// NewImageScan returns a new scans view.
func NewImageScan(gvr client.GVR) ResourceViewer {
	v := ImageScan{}
	v.ResourceViewer = NewBrowser(gvr)
	v.AddBindKeysFn(v.bindKeys)
	v.GetTable().SetEnterFn(v.viewCVE)
	v.GetTable().SetSortCol("SEVERITY", true)

	return &v
}

// Name returns the component name.
func (s *ImageScan) Name() string { return imgScanTitle }

func (c *ImageScan) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlZ, tcell.KeyCtrlW)

	aa.Bulk(ui.KeyMap{
		ui.KeyShiftL: ui.NewKeyAction("Sort Lib", c.GetTable().SortColCmd("LIBRARY", false), true),
		ui.KeyShiftS: ui.NewKeyAction("Sort Severity", c.GetTable().SortColCmd("SEVERITY", false), true),
		ui.KeyShiftF: ui.NewKeyAction("Sort Fixed-in", c.GetTable().SortColCmd("FIXED-IN", false), true),
		ui.KeyShiftV: ui.NewKeyAction("Sort Vulnerability", c.GetTable().SortColCmd("VULNERABILITY", false), true),
	})
}

func (s *ImageScan) viewCVE(app *App, _ ui.Tabular, _ client.GVR, path string) {
	bin := browseLinux
	if runtime.GOOS == "darwin" {
		bin = browseOSX
	}

	tt := strings.Split(path, "|")
	if len(tt) < 7 {
		app.Flash().Errf("parse path failed: %s", path)
	}
	cve := tt[render.CVEParseIdx]
	site := cveGovURL
	if strings.Index(cve, "GHSA") == 0 {
		site = ghsaURL
	}
	site += cve

	ok, errChan, _ := run(app, shellOpts{
		background: true,
		binary:     bin,
		args:       []string{site},
	})
	if !ok {
		app.Flash().Errf("unable to run browser command")
		return
	}
	var errs error
	for e := range errChan {
		errs = errors.Join(e)
	}
	if errs != nil {
		app.Flash().Err(errs)
	}
}
