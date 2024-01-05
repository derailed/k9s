// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog/log"

	"github.com/anchore/clio"
	"github.com/anchore/grype/cmd/grype/cli/options"
	"github.com/anchore/grype/grype"
	"github.com/anchore/grype/grype/db"
	"github.com/anchore/grype/grype/matcher"
	"github.com/anchore/grype/grype/matcher/dotnet"
	"github.com/anchore/grype/grype/matcher/golang"
	"github.com/anchore/grype/grype/matcher/java"
	"github.com/anchore/grype/grype/matcher/javascript"
	"github.com/anchore/grype/grype/matcher/python"
	"github.com/anchore/grype/grype/matcher/ruby"
	"github.com/anchore/grype/grype/matcher/stock"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/store"
	"github.com/anchore/grype/grype/vex"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ImgScanner *imageScanner

type imageScanner struct {
	store       *store.Store
	dbCloser    *db.Closer
	dbStatus    *db.Status
	opts        *options.Grype
	scans       Scans
	mx          sync.RWMutex
	initialized bool
	config      *config.ImageScans
}

// NewImageScanner returns a new instance.
func NewImageScanner(cfg *config.ImageScans) *imageScanner {
	return &imageScanner{
		scans:  make(Scans),
		config: cfg,
	}
}

func (s *imageScanner) ShouldExcludes(m metav1.ObjectMeta) bool {
	return s.config.ShouldExclude(m.Namespace, m.Labels)
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
	s.mx.Lock()
	defer s.mx.Unlock()

	id := clio.Identification{Name: name, Version: version}
	s.opts = options.DefaultGrype(id)
	s.opts.GenerateMissingCPEs = true

	var err error
	s.store, s.dbStatus, s.dbCloser, err = grype.LoadVulnerabilityDB(
		s.opts.DB.ToCuratorConfig(),
		s.opts.DB.AutoUpdate,
	)
	if err != nil {
		log.Error().Err(err).Msgf("VulDb load failed")
		return
	}

	if err := validateDBLoad(err, s.dbStatus); err != nil {
		log.Error().Err(err).Msgf("VulDb validate failed")
		return
	}

	s.initialized = true
}

// Stop closes scan database.
func (s *imageScanner) Stop() {
	s.mx.RLock()
	defer s.mx.RUnlock()

	if s.dbCloser != nil {
		s.dbCloser.Close()
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

func (s *imageScanner) isInitialized() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.initialized
}

func (s *imageScanner) Enqueue(images ...string) {
	if !s.isInitialized() {
		return
	}
	for _, i := range images {
		go func(img string) {
			if _, ok := s.GetScan(img); ok {
				return
			}
			sc := newScan(img)
			s.setScan(img, sc)
			if err := s.scan(img, sc); err != nil {
				log.Warn().Err(err).Msgf("Scan failed for img %s --", img)
			}
		}(i)
	}
}

func (s *imageScanner) scan(img string, sc *Scan) error {
	defer func(t time.Time) {
		log.Debug().Msgf("Scan %s images: %v", img, time.Since(t))
	}(time.Now())

	var errs error
	packages, pkgContext, _, err := pkg.Provide(img, getProviderConfig(s.opts))
	if err != nil {
		errs = errors.Join(errs, fmt.Errorf("failed to catalog %s: %w", img, err))
	}

	v := grype.VulnerabilityMatcher{
		Store:          *s.store,
		IgnoreRules:    s.opts.Ignore,
		NormalizeByCVE: s.opts.ByCVE,
		FailSeverity:   s.opts.FailOnServerity(),
		Matchers:       getMatchers(s.opts),
		VexProcessor: vex.NewProcessor(vex.ProcessorOptions{
			Documents:   s.opts.VexDocuments,
			IgnoreRules: s.opts.Ignore,
		}),
	}

	mm, _, err := v.FindMatches(packages, pkgContext)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	if err := sc.run(mm, s.store); err != nil {
		errs = errors.Join(errs, err)
	}

	return errs
}

func getProviderConfig(opts *options.Grype) pkg.ProviderConfig {
	return pkg.ProviderConfig{
		SyftProviderConfig: pkg.SyftProviderConfig{
			RegistryOptions:        opts.Registry.ToOptions(),
			Exclusions:             opts.Exclusions,
			CatalogingOptions:      opts.Search.ToConfig(),
			Platform:               opts.Platform,
			Name:                   opts.Name,
			DefaultImagePullSource: opts.DefaultImagePullSource,
		},
		SynthesisConfig: pkg.SynthesisConfig{
			GenerateMissingCPEs: opts.GenerateMissingCPEs,
		},
	}
}

func getMatchers(opts *options.Grype) []matcher.Matcher {
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
				UseCPEs:               opts.Match.Golang.UseCPEs,
				AlwaysUseCPEForStdlib: opts.Match.Golang.AlwaysUseCPEForStdlib,
			},
			Stock: stock.MatcherConfig(opts.Match.Stock),
		},
	)
}

func validateDBLoad(loadErr error, status *db.Status) error {
	if loadErr != nil {
		return fmt.Errorf("failed to load vulnerability db: %w", loadErr)
	}
	if status == nil {
		return fmt.Errorf("unable to determine the status of the vulnerability db")
	}
	if status.Err != nil {
		return fmt.Errorf("db could not be loaded: %w", status.Err)
	}

	return nil
}
