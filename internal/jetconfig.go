/*
Copyright (c) 2019 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Jettison main config parser.
 Reads configuration and serialize everything to appConfig struct.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package internal

import (
	"errors"
	"fmt"
	"github.com/spyroot/jettison/jettypes"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/vmware/govmomi/object"
	"gopkg.in/yaml.v2"
)

const (
	ConfigFile        = "/config.yml"
	DefaultConfigPath = "/Users/spyroot/go/src/github.com/spyroot/jettison/src"
)

type SshGlobalEnvironments struct {
	SshUsername   string `yaml:"Username"`
	SshPassword   string `yaml:"Password"`
	SshPublicKey  string `yaml:"Publickey"`
	SshPrivateKey string `yaml:"Privatekey"`
	SshpassPath   string `yaml:"sshpassPath"`
	SshCopyIdPath string `yaml:"sshCopyIdPath"`
	SshPort       int    `yaml:"Port"`
}

func (s *SshGlobalEnvironments) Username() string {
	return s.SshUsername
}

func (s *SshGlobalEnvironments) Password() string {
	return s.SshPassword
}
func (s *SshGlobalEnvironments) PublicKey() string {
	return s.SshPublicKey
}

func (s *SshGlobalEnvironments) PrivateKey() string {
	return s.SshPrivateKey
}
func (s *SshGlobalEnvironments) SshpassTool() string {
	return s.SshpassPath
}

func (s *SshGlobalEnvironments) SshCopyIdTool() string {
	return s.SshCopyIdPath
}

func (s *SshGlobalEnvironments) Port() int {
	return s.SshPort
}

type AnsibleEnvironments struct {
	AnsibleInventory string `yaml:"ansibleInventory"`
	AnsibleConfig    string `yaml:"ansibleConfig"`
	AnsibleTemplates string `yaml:"ansibleTemplate"`
	AnsiblePath      string `yaml:"ansiblePath"`
}

// K8S configuration controller struct
type Controller struct {
	DesiredAddress string `yaml:"desiredAddress"`
	Gateway        string `yaml:"gateway"`
	VM             string `yaml:"vmName"`
	UUID           string `yaml:"uuid"`
	Mac            string
	Vimname        string
	DhcpStatus     jettypes.Status
}

type ComputeConnector struct {
	Hostname   string `yaml:"hostname"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Datacenter string `yaml:"datacenter"`
	Dc         *object.Datacenter
}

//
//func (c *ComputeConnector) Datacenter() string {
//	if c != nil {
//		return c.Datacenter
//	}
//	return ""
//}

func (c *ComputeConnector) Endpoint() string {
	if c != nil {
		return c.Hostname
	}
	return ""
}

func (c *ComputeConnector) VimUsername() string {
	if c != nil {
		return c.Username
	}
	return ""
}

func (c *ComputeConnector) VimPassword() string {
	if c != nil {
		return c.Password
	}
	return ""
}

func (c *ComputeConnector) VimDatacenter() string {
	if c != nil {
		return c.Datacenter
	}
	return ""
}

type KubernetesCluster struct {
	ClusterCidr  string `yaml:"cluster-cidr"`
	ServiceCidr  string `yaml:"service-cidr"`
	ClusterDns   string `yaml:"cluster-dns"`
	AllocateSize int    `yaml:"allocate-size"`
}

type AppConfig struct {
	Infra struct {
		Vcenter ComputeConnector `yaml:"vcenter"`
		//	Nsxt             NsxtConfig       `yaml:"nsxt"`
		ParallelJobs     int    `yaml:"parallelJobs"`
		CleanupOnFailure bool   `yaml:"cleanupOnFailure"`
		DeploymentName   string `yaml:"deploymentName"`

		Cluster         KubernetesCluster                 `yaml:"cluster"`
		Scenario        map[string]*jettypes.NodeTemplate `yaml:"deployment"`
		Controllers     jettypes.NodeTemplate             `yaml:"controllers1"`
		WorkersTemplate jettypes.NodeTemplate             `yaml:"workers1"`
		IngressTemplate jettypes.NodeTemplate             `yaml:"ingress1"`
		SshDefaults     SshGlobalEnvironments             `yaml:"sshGlobal"`
		AnsibleDefaults AnsibleEnvironments               `yaml:"ansible"`
	}
}

func (a *AppConfig) GetComputeVim() *ComputeConnector {
	if a != nil {
		return &a.Infra.Vcenter
	}
	return &ComputeConnector{}
}

func (a *AppConfig) GetCluster() KubernetesCluster {
	return a.Infra.Cluster
}

func (a *AppConfig) GetMaxThreads() int {
	return a.Infra.ParallelJobs
}

func (a *AppConfig) GetAnsible() AnsibleEnvironments {
	return a.Infra.AnsibleDefaults
}

func (a *AppConfig) GetDeploymentName() string {
	return a.Infra.DeploymentName
}

// Returns data center name.
func (a *AppConfig) GetControllersTemplate() *jettypes.NodeTemplate {
	return &a.Infra.Controllers
}

// Returns data center name.
func (a *AppConfig) GetWorkersTemplate() *jettypes.NodeTemplate {
	return &a.Infra.WorkersTemplate
}

// Returns data center name.
func (a *AppConfig) GetIngresTemplate() *jettypes.NodeTemplate {
	return &a.Infra.IngressTemplate
}

// Returns data center name.
func (a AppConfig) GetDcName() string {
	return a.Infra.Vcenter.Datacenter
}

// Returns data center object.
func (a AppConfig) GetDc() *object.Datacenter {
	return a.Infra.Vcenter.Dc
}

func (a *AppConfig) GetSshUsername() string {
	return a.Infra.SshDefaults.Username()
}

func (a *AppConfig) GetSshPassword() string {
	return a.Infra.SshDefaults.Password()
}

func (a *AppConfig) GetSshPort() int {
	return a.Infra.SshDefaults.Port()
}

func (a *AppConfig) GetSshPublicKey() string {
	return a.Infra.SshDefaults.PublicKey()
}

func (a *AppConfig) GetSshCopyIdPath() string {
	return a.Infra.SshDefaults.SshCopyIdPath
}

func (a *AppConfig) GetSshPassTool() string {
	return a.Infra.SshDefaults.SshpassPath
}

func (a *AppConfig) GetSshDefault() *SshGlobalEnvironments {
	return &a.Infra.SshDefaults
}

func StatusString(status jettypes.Status) string {
	if status == jettypes.Created {
		return "create"
	}
	if status == jettypes.Error {
		return "error"
	}
	if status == jettypes.Deleted {
		return "deleted"
	}

	return "undefined"
}

func setDefaults(appConfig *AppConfig) {

	for _, v := range appConfig.Infra.Scenario {
		v.SetTemplate(true)

		// set default domain
		if len(v.DomainSuffix) == 0 {
			v.DomainSuffix = "cluster.local"
		}

		if v.DesiredCount == 0 {
			v.DesiredCount = 1
		}
	}
}

// Internal function validates mandatory configuration element.
// A Mandatory element are vCenter/ESZi hostname, username, password etc
func validate(appConfig *AppConfig) (bool, error) {

	if appConfig.Infra.DeploymentName == "" {
		return false, errors.New("missing deployment name")
	}

	if appConfig.Infra.Vcenter.Hostname == "" {
		return false, errors.New("missing vCenter/ESXi hostname entry")
	}

	if appConfig.Infra.Vcenter.Username == "" {
		return false, errors.New("missing vCenter/ESXi username entry")
	}

	if appConfig.Infra.Vcenter.Password == "" {
		return false, errors.New("missing vCenter/ESXi password entry")
	}

	for k, v := range appConfig.Infra.Scenario {
		ipv4Addr, ipv4Net, err := net.ParseCIDR(v.DesiredAddress)
		if err != nil {
			return false, fmt.Errorf("failed parse controllers desired address pool : %s", err)
		}
		v.IPv4Addr = ipv4Addr
		v.IPv4Net = ipv4Net

		if v.GenericSwitch() == nil || v.GenericSwitch().Name() == "" {
			v.SetExistingNetwork(true)
		}

		if v.DesiredCount == 0 {
			v.DesiredCount = 1
		}

		if len(v.Prefix) == 0 {
			v.Prefix = k
		}

		if len(v.DomainSuffix) == 0 {
			v.DomainSuffix = "cluster.local"
		}

		if len(v.Gateway) == 0 {
			return false, fmt.Errorf("gateway is mandatory configuration")
		}

		if !v.IPv4Net.Contains(net.ParseIP(v.Gateway)) {
			return false, fmt.Errorf("gateway outside of network range")
		}
	}

	// ansible checks
	if appConfig.GetAnsible().AnsibleConfig == "" {
		return false, errors.New("ansible config is empty")
	}
	if appConfig.GetAnsible().AnsibleInventory == "" {
		return false, errors.New("ansible inventory is empty")
	}
	if _, err := os.Stat(appConfig.GetAnsible().AnsibleConfig); os.IsNotExist(err) {
		return false, errors.New("ansible directory doesn't exists. check configuration")
	}
	if _, err := os.Stat(appConfig.GetAnsible().AnsibleInventory); os.IsNotExist(err) {
		return false, errors.New("ansible inventory directory doesn't exists. check configuration")
	}
	if _, err := os.Stat(appConfig.GetAnsible().AnsiblePath); os.IsNotExist(err) {
		return false, errors.New("ansible path invalid, check configuration")
	}

	//		ansibleInventory: /usr/local/etc/ansible/
	//		sshGlobal:
	//sshpassPath:   "/usr/local/bin/sshpass"
	//sshCopyIdPath: "/usr/bin/ssh-copy-id"
	//Username:   "vmware"
	//Password:   "VMware1!"
	//Privatekey:   "/Users/spyroot/.ssh/id_rsa"
	//Publickey:   "/Users/spyroot/.ssh/id_rsa.pub"
	//Port:  22

	return true, nil
}

/**
  Reads config.yml file and serialize everything in JetConfig struct.
*/
func ReadConfig() (AppConfig, error) {

	var appConfig AppConfig

	pwd, _ := os.Getwd()
	log.Println(" Reading config ", pwd+ConfigFile)
	p := filepath.Join(pwd, ConfigFile)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		log.Println(" Reading default config ", DefaultConfigPath+ConfigFile)
		p := filepath.Join(DefaultConfigPath, ConfigFile)
		data, err = ioutil.ReadFile(p)
		if err != nil {
			return appConfig, err
		}
	}

	log.Println(" Parsing yaml file.")
	err = yaml.Unmarshal(data, &appConfig)
	if err != nil {
		return appConfig, err
	}

	var r bool
	r, err = validate(&appConfig)
	if r == false {
		return appConfig, err
	}
	// sets a default
	setDefaults(&appConfig)

	//
	if appConfig.Infra.Vcenter.Password == "*" {
		for ok := true; ok; ok = !(len(appConfig.Infra.Vcenter.Password) > 1) {
			fmt.Print("vCenter password: ")
			_, err = fmt.Scanln(&appConfig.Infra.Vcenter.Password)
			if err != nil {
				ok = false
			}
		}
	}

	return appConfig, nil
}
