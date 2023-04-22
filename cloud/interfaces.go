package cloud

import (
	"context"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

type Reconciler interface {
	Reconcile(ctx context.Context) error
	Delete(ctx context.Context) error
}

type Client interface {
}

type Cluster interface {
	ClusterGetter
	ClusterSettter
}

type ClusterGetter interface {
	Client
	Name() string
	Namespace() string
}

type ClusterSettter interface {
	SetControlPlaneEndpoint(endpoint clusterv1.APIEndpoint)
}
