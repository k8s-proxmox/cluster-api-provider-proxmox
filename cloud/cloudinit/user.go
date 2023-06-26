package cloudinit

import (
	"fmt"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

func ParseUser(content string) (*infrav1.User, error) {
	var config *infrav1.User
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func GenerateUserYaml(config infrav1.User) (string, error) {
	b, err := yaml.Marshal(&config)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("#cloud-config\n%s", string(b)), nil
}

func MergeUsers(a, b infrav1.User) (*infrav1.User, error) {
	if err := mergo.Merge(&a, b, mergo.WithAppendSlice); err != nil {
		return nil, err
	}
	return &a, nil
}
