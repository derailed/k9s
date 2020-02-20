package dao

import (
	appsv1 "k8s.io/api/apps/v1"
)

// ReplicaSet represents a replicaset K8s resource.
type ReplicaSet struct {
	Resource
}

// IsHappy check for happy deployments.
func (d *ReplicaSet) IsHappy(rs appsv1.ReplicaSet) bool {
	if rs.Status.Replicas == 0 && rs.Status.Replicas != rs.Status.ReadyReplicas {
		return false
	}

	if rs.Status.Replicas != 0 && rs.Status.Replicas != rs.Status.ReadyReplicas {
		return false
	}

	return true
}
