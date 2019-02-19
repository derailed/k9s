package k8s

// Cluster manages a Kubernetes ClusterRole.
type Cluster struct{}

// NewCluster instantiates a new ClusterRole.
func NewCluster() *Cluster {
	return &Cluster{}
}

// Version returns the current cluster git version.
func (c *Cluster) Version() (string, error) {
	rev, err := conn.dialOrDie().Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return rev.GitVersion, nil
}

// ContextName returns the currently active context.
func (c *Cluster) ContextName() string {
	ctx, err := conn.config.CurrentContextName()
	if err != nil {
		return "N/A"
	}
	return ctx
}

// ClusterName return the currently active cluster name.
func (c *Cluster) ClusterName() string {
	ctx, err := conn.config.CurrentClusterName()
	if err != nil {
		return "N/A"
	}
	return ctx
}

// UserName returns the currently active user.
func (c *Cluster) UserName() string {
	usr, err := conn.config.CurrentUserName()
	if err != nil {
		return "N/A"
	}
	return usr
}
