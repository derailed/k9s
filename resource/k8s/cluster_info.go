package k8s

type ClusterInfo struct{}

func (ClusterInfo) ActiveClusterOrDie() string {
	return ActiveClusterOrDie()
}

func (ClusterInfo) AllClustersOrDie() []string {
	return AllClustersOrDie()
}

func (ClusterInfo) AllNamespacesOrDie() []string {
	return AllNamespacesOrDie()
}