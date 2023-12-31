// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

const (
	// NonResource represents a custom resource.
	NonResource = "*"
)

const (
	// Terminating represents a pod terminating status.
	Terminating = "Terminating"

	// Running represents a pod running status.
	Running = "Running"

	// Initialized represents a pod initialized status.
	Initialized = "Initialized"

	// Completed represents a pod completed status.
	Completed = "Completed"

	// ContainerCreating represents a pod container status.
	ContainerCreating = "ContainerCreating"

	// PodInitializing represents a pod initializing status.
	PodInitializing = "PodInitializing"

	// Pending represents a pod pending status.
	Pending = "Pending"

	// Blank represents no value.
	Blank = ""
)

const (
	// MissingValue indicates an unset value.
	MissingValue = "<none>"

	// NAValue indicates a value that does not pertain.
	NAValue = "n/a"

	// UnknownValue represents an unknown.
	UnknownValue = "<unknown>"

	// UnsetValue represent an unset value.
	UnsetValue = ""

	// ZeroValue represents a zero value.
	ZeroValue = "0"
)
