// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

// Labels tracks a collection of labels.
type Labels map[string][]string

func (l Labels) exclude(k, val string) bool {
	vv, ok := l[k]
	if !ok {
		return false
	}

	for _, v := range vv {
		if v == val {
			return true
		}
	}

	return false
}

// ScanExcludes tracks vul scan exclusions.
type ScanExcludes struct {
	Namespaces []string `json:"namespaces" yaml:"namespaces"`
	Labels     Labels   `json:"labels" yaml:"labels"`
}

func newScanExcludes() ScanExcludes {
	return ScanExcludes{
		Labels: make(Labels),
	}
}

func (b ScanExcludes) exclude(ns string, ll map[string]string) bool {
	for _, nss := range b.Namespaces {
		if nss == ns {
			return true
		}
	}
	for k, v := range ll {
		if b.Labels.exclude(k, v) {
			return true
		}
	}

	return false
}

// ImageScans tracks vul scans options.
type ImageScans struct {
	Enable     bool         `json:"enable" yaml:"enable"`
	Exclusions ScanExcludes `json:"exclusions" yaml:"exclusions"`
}

// NewImageScans returns a new instance.
func NewImageScans() ImageScans {
	return ImageScans{
		Exclusions: newScanExcludes(),
	}
}

// ShouldExclude checks if scan should be excluded given ns/labels
func (i ImageScans) ShouldExclude(ns string, ll map[string]string) bool {
	if !i.Enable {
		return false
	}

	return i.Exclusions.exclude(ns, ll)
}
