package config

type AppConfig struct {
	Ecs struct {
		Vcenter struct {
			Hostname string `yaml:"hostname"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"vcenter"`
		Nsxt struct {
			Hostname string `yaml:"hostname"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"nsxt"`

		Controllers []struct {
			Name           string `yaml:"name"`
			TaskDefinition string `yaml:"desiredAddress"`
			VM             string `yaml:"vmName"`
		} `yaml:"controllers"`
	}
}
