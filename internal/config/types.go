package config

type VMInfo struct {
	Name       string
	LineNumber int
}

type AppConfig struct {
	VMs        []VMInfo
	SSHKeyPath string
}
