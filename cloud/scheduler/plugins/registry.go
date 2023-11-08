package plugins

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/framework"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/idrange"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/nodename"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/noderesource"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/overcommit"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/random"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/scheduler/plugins/regex"
)

type PluginConfigs struct {
	filterPlugins map[string]pluginConfig `yaml:"filters"`
	scorePlugins  map[string]pluginConfig `yaml:"scores"`
	vmidPlugins   map[string]pluginConfig `yaml:"vmids"`
}

type pluginConfig struct {
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
		filterPlugins: NewNodeFilterPlugins(configs.filterPlugins),
		scorePlugins:  NewNodeScorePlugins(configs.scorePlugins),
		vmidPlugins:   NewVMIDPlugins(configs.vmidPlugins),
	}
	return r
}

func NewNodeFilterPlugins(config map[string]pluginConfig) []framework.NodeFilterPlugin {
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

func NewNodeScorePlugins(config map[string]pluginConfig) []framework.NodeScorePlugin {
	pls := []framework.NodeScorePlugin{
		&random.Random{},
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

func NewVMIDPlugins(config map[string]pluginConfig) []framework.VMIDPlugin {
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
	config := PluginConfigs{}
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
