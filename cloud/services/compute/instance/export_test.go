package instance

import (
	infrav1 "github.com/k8s-proxmox/cluster-api-provider-proxmox/api/v1beta1"
)

func MergeUserDatas(a, b, c *infrav1.UserData) (*infrav1.UserData, error) {
	return mergeUserDatas(a, b, c)
}
