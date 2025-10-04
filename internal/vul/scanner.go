// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/anchore/clio"
	"github.com/anchore/grype/cmd/grype/cli/options"
	"github.com/anchore/grype/grype"
	"github.com/anchore/grype/grype/match"
	"github.com/anchore/grype/grype/matcher"
	"github.com/anchore/grype/grype/matcher/dotnet"
	"github.com/anchore/grype/grype/matcher/golang"
	"github.com/anchore/grype/grype/matcher/java"
	"github.com/anchore/grype/grype/matcher/javascript"
	"github.com/anchore/grype/grype/matcher/python"
	"github.com/anchore/grype/grype/matcher/ruby"
	"github.com/anchore/grype/grype/matcher/stock"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/vex"
	"github.com/anchore/grype/grype/vulnerability"
	"github.com/anchore/syft/syft"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/slogs"
)

var ImgScanner *imageScanner

const (
	imgChanSize     = 3
	imgScanTimeout  = 2 * time.Second
	scanConcurrency = 2
)

type imageScanner struct {
	provider    vulnerability.Provider
	status      *vulnerability.ProviderStatus
	opts        *options.Grype
	scans       Scans
	mx          sync.RWMutex
	initialized bool
	config      config.ImageScans
	log         *slog.Logger
}

// NewImageScanner returns a new instance.
func NewImageScanner(cfg config.ImageScans, l *slog.Logger) *imageScanner {
	return &imageScanner{
		scans:  make(Scans),
		config: cfg,
		log:    l.With(slogs.Subsys, "vul"),
	}
}

func (s *imageScanner) ShouldExcludes(ns string, lbls map[string]string) bool {
	return s.config.ShouldExclude(ns, lbls)
}

// GetScan fetch scan for a given image. Returns ok=false when not found.
func (s *imageScanner) GetScan(img string) (*Scan, bool) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	scan, ok := s.scans[img]

	return scan, ok
}

func (s *imageScanner) setScan(img string, sc *Scan) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.scans[img] = sc
}

// Init initializes image vulnerability database.
func (s *imageScanner) Init(name, version string) {
	defer func(t time.Time) {
		slog.Debug("VulDb initialization complete",
			slogs.Elapsed, time.Since(t),
		)
	}(time.Now())

	opts := options.DefaultGrype(clio.Identification{Name: name, Version: version})
	opts.GenerateMissingCPEs = true

	provider, status, err := grype.LoadVulnerabilityDB(
		opts.ToClientConfig(),
		opts.ToCuratorConfig(),
		opts.DB.AutoUpdate,
	)
	if err != nil {
		s.log.Error("VulDb load failed", slogs.Error, err)
		return
	}
	s.mx.Lock()
	s.opts, s.provider, s.status = opts, provider, status
	s.mx.Unlock()

	if e := validateDBLoad(err, status); e != nil {
		s.log.Error("VulDb validate failed", slogs.Error, e)
		return
	}

	s.mx.Lock()
	s.initialized = true
	s.mx.Unlock()
	slog.Debug("VulDB initialized")
}

// Stop closes scan database.
func (s *imageScanner) Stop() {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if s.provider != nil {
		_ = s.provider.Close()
		s.provider = nil
	}
}

func (s *imageScanner) Score(ii ...string) string {
	var sc scorer
	for _, i := range ii {
		if scan, ok := s.GetScan(i); ok {
			sc = sc.Add(newScorer(scan.Tally))
		}
	}

	return sc.String()
}

func (s *imageScanner) IsInitialized() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.initialized
}

func (s *imageScanner) Enqueue(ctx context.Context, images ...string) {
	ctx, cancel := context.WithTimeout(ctx, imgScanTimeout)
	defer cancel()

	for _, img := range images {
		if _, ok := s.GetScan(img); ok {
			continue
		}
		go s.scanWorker(ctx, img)
	}
}

func (s *imageScanner) scanWorker(ctx context.Context, img string) {
	defer s.log.Debug("ScanWorker bailing out!")

	s.log.Debug("ScanWorker processing image", slogs.Image, img)
	sc := newScan(img)
	s.setScan(img, sc)
	if err := s.scan(ctx, img, sc); err != nil {
		s.log.Warn("Scan failed for image",
			slogs.Image, img,
			slogs.Error, err,
		)
	}
}

func (s *imageScanner) scan(_ context.Context, img string, sc *Scan) error {
	defer func(t time.Time) {
		s.log.Debug("[Vulscan] perf",
			slogs.Image, img,
			slogs.Elapsed, time.Since(t),
		)
	}(time.Now())

	var errs error
	packages, pkgContext, _, err := pkg.Provide(img, getProviderConfig(s.opts))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to catalog %s: %w", img, err))
	}

	v := grype.VulnerabilityMatcher{
		VulnerabilityProvider: s.provider,
		IgnoreRules:           s.opts.Ignore,
		NormalizeByCVE:        s.opts.ByCVE,
		FailSeverity:          s.opts.FailOnSeverity(),
		Matchers:              getMatchers(s.opts),
		VexProcessor: vex.NewProcessor(vex.ProcessorOptions{
			Documents:   s.opts.VexDocuments,
			IgnoreRules: s.opts.Ignore,
		}),
	}

	mm, _, err := v.FindMatches(packages, pkgContext)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	if err := sc.run(mm, s.provider); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

func getProviderConfig(opts *options.Grype) pkg.ProviderConfig {
	return pkg.ProviderConfig{
		SyftProviderConfig: pkg.SyftProviderConfig{
			SBOMOptions:            syft.DefaultCreateSBOMConfig(),
			RegistryOptions:        opts.Registry.ToOptions(),
			Exclusions:             opts.Exclusions,
			Platform:               opts.Platform,
			Name:                   opts.Name,
			DefaultImagePullSource: opts.DefaultImagePullSource,
		},
		SynthesisConfig: pkg.SynthesisConfig{
			GenerateMissingCPEs: opts.GenerateMissingCPEs,
		},
	}
}

func getMatchers(opts *options.Grype) []match.Matcher {
	return matcher.NewDefaultMatchers(
		matcher.Config{
			Java: java.MatcherConfig{
				ExternalSearchConfig: opts.ExternalSources.ToJavaMatcherConfig(),
				UseCPEs:              opts.Match.Java.UseCPEs,
			},
			Ruby:       ruby.MatcherConfig(opts.Match.Ruby),
			Python:     python.MatcherConfig(opts.Match.Python),
			Dotnet:     dotnet.MatcherConfig(opts.Match.Dotnet),
			Javascript: javascript.MatcherConfig(opts.Match.Javascript),
			Golang: golang.MatcherConfig{
				UseCPEs:                                opts.Match.Golang.UseCPEs,
				AlwaysUseCPEForStdlib:                  opts.Match.Golang.AlwaysUseCPEForStdlib,
				AllowMainModulePseudoVersionComparison: opts.Match.Golang.AllowMainModulePseudoVersionComparison,
			},
			Stock: stock.MatcherConfig(opts.Match.Stock),
		},
	)
}
func validateDBLoad(loadErr error, status *vulnerability.ProviderStatus) error {
	if loadErr != nil {
		return fmt.Errorf("failed to load vulnerability db: %w", loadErr)
	}
	if status == nil {
		return fmt.Errorf("unable to determine the status of the vulnerability db")
	}
	if status.Error != nil {
		return fmt.Errorf("db could not be loaded: %w", status.Error)
	}

	return nil
}
