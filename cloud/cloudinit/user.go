package cloudinit

import (
	"fmt"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"

	infrav1 "github.com/k8s-proxmox/cluster-api-provider-proxmox/api/v1beta1"
)

func ParseUserData(content string) (*infrav1.UserData, error) {
	var config *infrav1.UserData
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GenerateUserDataYaml(config infrav1.UserData) (string, error) {
	b, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#cloud-config\n%s", string(b)), nil
}

func MergeUserDatas(a, b *infrav1.UserData) (*infrav1.UserData, error) {
	if err := mergo.Merge(a, b, mergo.WithAppendSlice); err != nil {
		return nil, err
	}
	return a, nil
}
