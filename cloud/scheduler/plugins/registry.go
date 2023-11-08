package plugins

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/idrange"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/nodename"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/noderesource"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/overcommit"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/regex"
)

type PluginConfigs struct {
	FilterPlugins map[string]PluginConfig `yaml:"filters,omitempty"`
	ScorePlugins  map[string]PluginConfig `yaml:"scores,omitempty"`
	VMIDPlugins   map[string]PluginConfig `yaml:"vmids,omitempty"`
}

type PluginConfig struct {
	Enable bool                   `yaml:"enable,omitempty"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

type PluginRegistry struct {
	filterPlugins []framework.NodeFilterPlugin
	scorePlugins  []framework.NodeScorePlugin
	vmidPlugins   []framework.VMIDPlugin
}

func (r *PluginRegistry) FilterPlugins() []framework.NodeFilterPlugin {
	return r.filterPlugins
}

func (r *PluginRegistry) ScorePlugins() []framework.NodeScorePlugin {
	return r.scorePlugins
}

func (r *PluginRegistry) VMIDPlugins() []framework.VMIDPlugin {
	return r.vmidPlugins
}

func NewRegistry(configs PluginConfigs) PluginRegistry {
	r := PluginRegistry{
		filterPlugins: NewNodeFilterPlugins(configs.FilterPlugins),
		scorePlugins:  NewNodeScorePlugins(configs.ScorePlugins),
		vmidPlugins:   NewVMIDPlugins(configs.VMIDPlugins),
	}
	return r
}

func NewNodeFilterPlugins(config map[string]PluginConfig) []framework.NodeFilterPlugin {
	pls := []framework.NodeFilterPlugin{
		&nodename.NodeName{},
		&overcommit.CPUOvercommit{},
		&overcommit.MemoryOvercommit{},
		&regex.NodeRegex{},
	}
	plugins := []framework.NodeFilterPlugin{}
	for _, pl := range pls {
		c, ok := config[pl.Name()]
		if ok && !c.Enable {
			continue
		}
		plugins = append(plugins, pl)
	}
	return plugins
}

func NewNodeScorePlugins(config map[string]PluginConfig) []framework.NodeScorePlugin {
	pls := []framework.NodeScorePlugin{
		&noderesource.NodeResource{},
	}
	plugins := []framework.NodeScorePlugin{}
	for _, pl := range pls {
		c, ok := config[pl.Name()]
		if ok && !c.Enable {
			continue
		}
		plugins = append(plugins, pl)
	}
	return plugins
}

func NewVMIDPlugins(config map[string]PluginConfig) []framework.VMIDPlugin {
	pls := []framework.VMIDPlugin{
		&idrange.Range{},
		&regex.Regex{},
	}
	plugins := []framework.VMIDPlugin{}
	for _, pl := range pls {
		c, ok := config[pl.Name()]
		if ok && !c.Enable {
			continue
		}
		plugins = append(plugins, pl)
	}
	return plugins
}

// Read config file and unmarshal it to PluginConfig type
func GetPluginConfigFromFile(path string) (PluginConfigs, error) {
	var config PluginConfigs
	if path == "" {
		return config, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	if err := yaml.Unmarshal(b, &config); err != nil {
		return config, err
	}
	return config, nil
}
