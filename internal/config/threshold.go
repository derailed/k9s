package config

import (
	"github.com/derailed/k9s/internal/client"
)

const (
	defaultCPU  = 80
	defaultMEM  = 80
	defaultDisk = 70
)

// Threshold tracks threshold to alert user when excided.
type Threshold struct {
	CPU    int `yaml:"cpu"`
	Memory int `yaml:"memory"`
	Disk   int `yaml:"disk"`
}

func newThreshold() *Threshold {
	return &Threshold{
		CPU:    defaultCPU,
		Memory: defaultMEM,
		Disk:   defaultMEM,
	}
}

// Validate a namespace is setup correctly
func (t *Threshold) Validate(c client.Connection, ks KubeSettings) {
	if t.CPU == 0 || t.CPU > 100 {
		t.CPU = defaultCPU
	}
	if t.Memory == 0 || t.Memory > 100 {
		t.Memory = defaultMEM
	}
	if t.Disk == 0 || t.Disk > 100 {
		t.Disk = defaultDisk
	}
}

// ExceedsCPUPerc returns true if current metrics exceeds threshold or false otherwise.
func (t *Threshold) ExceedsCPUPerc(p int) bool {
	return p >= t.CPU
}

// ExceedsMemoryPerc returns true if current metrics exceeds threshold or false otherwise.
func (t *Threshold) ExceedsMemoryPerc(p int) bool {
	return p >= t.Memory
}

// ExceedsDiskPerc returns true if current metrics exceeds threshold or false otherwise.
func (t *Threshold) ExceedsDiskPerc(p int) bool {
	return p >= t.Disk
}
