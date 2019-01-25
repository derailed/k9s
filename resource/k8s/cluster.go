package k8s

// Cluster manages a Kubernetes ClusterRole.
type Cluster struct{}

// NewCluster instantiates a new ClusterRole.
func NewCluster() *Cluster {
	return &Cluster{}
}

// Version retrieves cluster git version.
func (c *Cluster) Version() (string, error) {
	rev, err := conn.dialOrDie().Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return rev.GitVersion, nil
}

// ClusterName retrieves cluster name.
func (c *Cluster) ClusterName() string {
	return conn.getClusterName()
}
