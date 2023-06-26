package cloudinit

import (
	"gopkg.in/yaml.v3"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

func ParseNetwork(content string) (*infrav1.Network, error) {
	var config *infrav1.Network
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GenerateNetworkYaml(config infrav1.Network) (string, error) {
	b, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
