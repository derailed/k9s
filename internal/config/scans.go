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

// Blacklist tracks vul scan exclusions.
type BlackList struct {
	Namespaces []string `yaml:"namespaces"`
	Labels     Labels   `yaml:"labels"`
}

func newBlackList() BlackList {
	return BlackList{
		Labels: make(Labels),
	}
}

func (b BlackList) exclude(ns string, ll map[string]string) bool {
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
	Enable    bool      `yaml:"enable"`
	BlackList BlackList `yaml:"blackList"`
}

// NewImageScans returns a new instance.
func NewImageScans() *ImageScans {
	return &ImageScans{
		BlackList: newBlackList(),
	}
}

// ShouldExclude checks if scan should be excluder given ns/labels
func (i *ImageScans) ShouldExclude(ns string, ll map[string]string) bool {
	if !i.Enable {
		return false
	}

	return i.BlackList.exclude(ns, ll)
}
