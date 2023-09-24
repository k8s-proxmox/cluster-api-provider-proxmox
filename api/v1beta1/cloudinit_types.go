package v1beta1

// CloudInit is passed to disk directly as raw yaml file
// not via Proxmox API so you can configure more detailed configs
type CloudInit struct {
	User *User `json:"user,omitempty"`
}

type User struct {
	BootCmd           []string     `yaml:"bootcmd,omitempty" json:"bootcmd,omitempty"`
	CACerts           CACert       `yaml:"ca_certs,omitempty" json:"ca_certs,omitempty"`
	ChPasswd          ChPasswd     `yaml:"chpasswd,omitempty" json:"chpasswd,omitempty"`
	HostName          string       `yaml:"hostname,omitempty" json:"-"`
	ManageEtcHosts    bool         `yaml:"manage_etc_hosts,omitempty" json:"manage_etc_hosts,omitempty"`
	NoSSHFingerprints bool         `yaml:"no_ssh_fingerprints,omitempty" json:"no_ssh_fingerprints,omitempty"`
	Packages          []string     `yaml:"packages,omitempty" json:"packages,omitempty"`
	PackageUpdate     bool         `yaml:"package_update,omitempty" json:"package_update,omitempty"`
	PackageUpgrade    bool         `yaml:"package_upgrade,omitempty" json:"package_upgrade,omitempty"`
	Password          string       `yaml:"password,omitempty" json:"password,omitempty"`
	RunCmd            []string     `yaml:"runcmd,omitempty" json:"runCmd,omitempty"`
	SSH               SSH          `yaml:"ssh,omitempty" json:"ssh,omitempty"`
	SSHAuthorizedKeys []string     `yaml:"ssh_authorized_keys,omitempty" json:"ssh_authorized_keys,omitempty"`
	SSHKeys           SSHKeys      `yaml:"ssh_keys,omitempty" json:"ssh_keys,omitempty"`
	SSHPWAuth         bool         `yaml:"ssh_pwauth,omitempty" json:"ssh_pwauth,omitempty"`
	User              string       `yaml:"user,omitempty" json:"user,omitempty"`
	Users             []string     `yaml:"users,omitempty" json:"-"`
	WriteFiles        []WriteFiles `yaml:"write_files,omitempty" json:"writeFiles,omitempty"`
}

type CACert struct {
	RemoveDefaults bool     `yaml:"remove_defaults,omitempty" json:"remove_defaults,omitempty"`
	Trusted        []string `yaml:"trusted,omitempty" json:"trusted,omitempty"`
}

type ChPasswd struct {
	Expire string `yaml:"expire,omitempty" json:"expire,omitempty"`
}

type SSH struct {
	EmitKeysToConsole bool `yaml:"emit_keys_to_console,omitempty" json:"emit_keys_to_console,omitempty"`
}

type SSHKeys struct {
	RSAPrivate   string `yaml:"rsa_private,omitempty" json:"rsa_private,omitempty"`
	RSAPublic    string `yaml:"rsa_public,omitempty" json:"rsa_public,omitempty"`
	DSAPrivate   string `yaml:"dsa_private,omitempty" json:"dsa_private,omitempty"`
	DSAPublic    string `yaml:"dsa_public,omitempty" json:"dsa_public,omitempty"`
	ECDSAPrivate string `yaml:"ecdsa_private,omitempty" json:"ecdsa_private,omitempty"`
	EDSCAPublic  string `yaml:"ecdsa_public,omitempty" json:"ecdsa_public,omitempty"`
}

type WriteFiles struct {
	Encoding    string `yaml:"encoding,omitempty" json:"encoding,omitempty"`
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	Owner       string `yaml:"owner,omitempty" json:"owner,omitempty"`
	Permissions string `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Defer       bool   `yaml:"defer,omitempty" json:"defer,omitempty"`
	Content     string `yaml:"content,omitempty" json:"content,omitempty"`
}
