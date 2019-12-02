package render

const (
	// AllNamespaces represents all namespaces.
	AllNamespaces = ""

	// NamespaceAll represent the all namespace.
	NamespaceAll = "all"

	// ClusterWide represents a cluster resources.
	ClusterWide = "-"

	// NonResource represents a custom resource.
	NonResource = "*"
)

const (
	// Terminating represents a pod terminating status.
	Terminating = "Terminating"

	// Running represents a pod running status.
	Running = "Running"

	// Initialized represents a pod intialized status.
	Initialized = "Initialized"

	// Completed represents a pod completed status.
	Completed = "Completed"

	// ContainerCreating represents a pod container status.
	ContainerCreating = "ContainerCreating"

	// PodInitializing represents a pod initializing status.
	PodInitializing = "PodInitializing"
)
