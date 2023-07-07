package cloudinit_test

import (
	"testing"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/cloudinit"
)

func TestIsValidType(t *testing.T) {
	testTypes := []string{"user", "users", "meta", "memetah", "network", "networkig"}
	expectedValidity := []bool{true, false, true, false, true, false}
	for i, tt := range testTypes {
		if cloudinit.IsValidType(tt) != expectedValidity[i] {
			t.Errorf("expected validity for %s is %v", tt, expectedValidity[i])
		}
	}
}
