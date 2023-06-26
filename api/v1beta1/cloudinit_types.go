package v1beta1

// CloudInit defines options related to the bootstrapping systems where
// CloudInit is used.
type CloudInit struct {
	User    User    `json:"user,omitempty"`
	Meta    Meta    `json:"meta,omitempty"`
	Network Network `json:"network,omitempty"`
}

type User struct {
	GrowPart       GrowPart     `yaml:"growpart,omitempty" json:"-"`
	HostName       string       `yaml:"hostname,omitempty" json:"-"`
	ManageEtcHosts bool         `yaml:"manage_etc_hosts,omitempty" json:"-"`
	User           string       `yaml:"user,omitempty" json:"user,omitempty"`
	ChPasswd       ChPasswd     `yaml:"chpasswd,omitempty" json:"-"`
	Users          []string     `yaml:"users,omitempty" json:"-"`
	Password       string       `yaml:"password,omitempty" json:"password,omitempty"`
	Packages       []string     `yaml:"packages,omitempty" json:"-"`
	PackageUpgrade bool         `yaml:"package_upgrade,omitempty" json:"-"`
	WriteFiles     []WriteFiles `yaml:"write_files,omitempty" json:"-"`
	RunCmd         []string     `yaml:"runcmd,omitempty" json:"-"`
}

type Network struct {
	Version int             `json:"version,omitempty"`
	Config  []NetworkConfig `json:"config,omitempty"`
}

type NetworkConfig struct {
	Type        string   `json:"type,omitempty"`
	Name        string   `json:"name,omitempty"`
	MacAddress  string   `json:"mac_address,omitempty"`
	Subnets     []Subnet `json:"subnets,omitempty"`
	Destination string   `json:"destination,omitempty"`
	Gateway     string   `json:"gateway,omitempty"`
}

type Subnet struct {
	Type    string `json:"type,omitempty"`
	Address string `json:"address,omitempty"`
	Gateway string `json:"gateway,omitempty"`
}

type Meta struct {
}

type GrowPart struct {
	Mode                   string   `yaml:"mode,omitempty" json:"-"`
	Devices                []string `yaml:"devices,omitempty" json:"-"`
	IgnoreGrowrootDisabled bool     `yaml:"ignore_growroot_disabled,omitempty" json:"-"`
}

type ChPasswd struct {
	Expire string `yaml:"expire,omitempty" json:"-"`
}

type WriteFiles struct {
	Path        string `yaml:"path,omitempty" json:"-"`
	Owner       string `yaml:"owner,omitempty" json:"-"`
	Permissions string `yaml:"permissions,omitempty" json:"-"`
	Content     string `yaml:"content,omitempty" json:"-"`
}
