package instance

import (
	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

func MergeUserDatas(a, b, c *infrav1.UserData) (*infrav1.UserData, error) {
	return mergeUserDatas(a, b, c)
}
