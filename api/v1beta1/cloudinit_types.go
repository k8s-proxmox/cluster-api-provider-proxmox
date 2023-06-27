package v1beta1

// CloudInit is passed through raw yaml file not Proxmox API
// so you can configure more detailed configs
type CloudInit struct {
	User *User `json:"user,omitempty"`
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
