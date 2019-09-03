/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

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
package config

import (
	"errors"
	"fmt"
	"github.com/vmware/govmomi/object"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	ConfigFile        = "/config.yml"
	DefaultConfigPath = "/Users/spyroot/go/src/github.com/spyroot/jettison/src"
)

type Status int

const (
	Undefined Status = iota
	Created
	Deleted
	Error
)

type SshGlobalEnvironments struct {
	Username      string `yaml:"Username"`
	Password      string `yaml:"Password"`
	PublicKey     string `yaml:"Publickey"`
	SshpassPath   string `yaml:"sshpassPath"`
	SshCopyIdPath string `yaml:"sshCopyIdPath"`
	Port          string `yaml:"Port"`
}

// K8S configuration controller struct
type Controller struct {
	Name           string `yaml:"name"`
	DesiredAddress string `yaml:"desiredAddress"`
	VM             string `yaml:"vmName"`
	UUID           string `yaml:"uuid"`
	Mac            string
	Vimname        string
	DhcpStatus     Status
}

//
type ComputeConnector struct {
	Hostname   string `yaml:"hostname"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	Datacenter string `yaml:"datacenter"`
	Dc         *object.Datacenter
}

type AppConfig struct {
	Infra struct {
		Vcenter ComputeConnector `yaml:"vcenter"`
		Nsxt    struct {
			Hostname      string `yaml:"hostname"`
			Username      string `yaml:"username"`
			Password      string `yaml:"password"`
			LogicalSwitch string `yaml:"logicalSwitch"`
			DhcpServerID  string
		} `yaml:"nsxt"`

		Controllers     []*Controller         `yaml:"controllers"`
		SshDefaults     SshGlobalEnvironments `yaml:"sshGlobal"`
		WorkersTemplate K8SNodes              `yaml:"workers"`
	}
}

type K8SNodes struct {
	Prefix         string `yaml:"prefix"`
	DomainSuffix   string `yaml:"domainSufix"`
	DesiredCount   string `yaml:"desiredCount"`
	FirstNode      string `yaml:"startIP"`
	LastNode       string `yaml:"stopIP"`
	VmTemplateName string `yaml:"vmTemplateName"`
	UUID           string `yaml:"uuid"`
	vimCluster     string `yaml:"clusterName"`
	mac            string
	vimName        string
}

// Returns data center name.
func (a AppConfig) GetDcName() string {
	return a.Infra.Vcenter.Datacenter
}

// Returns data center object.
func (a AppConfig) GetDc() *object.Datacenter {
	return a.Infra.Vcenter.Dc
}

// Returns data center name.
func (a AppConfig) GetLogicalSwitch() string {
	return a.Infra.Nsxt.LogicalSwitch
}

func StatusString(status Status) string {
	if status == Created {
		return "create"
	}
	if status == Error {
		return "error"
	}
	if status == Deleted {
		return "deleted"
	}

	return "undefined"
}

// Return controller related variable (VM name, Desired IP Address etc)
func (a AppConfig) GetController(uuid string) (*Controller, error) {

	for _, val := range a.Infra.Controllers {
		if val.UUID == uuid {
			return val, nil
		}
	}
	return nil, fmt.Errorf("VM reference not found: %s", uuid)
}

//
//// Return controller related variable (VM name, Desired IP Address etc)
//func (a AppConfig) GetController(uuid string) ( *struct {
//	Name           string `yaml:"name"`
//	DesiredAddress string `yaml:"desiredAddress"`
//	VM             string `yaml:"vmName"`
//	UUID           string `yaml:"uuid"`
//	Mac            string
//	Vimname        string
//}, error) {
//
//	for _, val := range a.Infra.Controllers {
//		if val.UUID == uuid {
//			return &val, nil
//		}
//	}
//	return nil, fmt.Errorf("VM reference not found: %s", uuid)
//}

// Internal function validates mandatory configuration element.
// A Mandatory element are vCenter/ESZi hostname, username, password etc
func validate(appConfig *AppConfig) (bool, error) {

	if appConfig.Infra.Vcenter.Hostname == "" {
		return false, errors.New("missing vCenter/ESXi hostname entry")
	}

	if appConfig.Infra.Vcenter.Username == "" {
		return false, errors.New("missing vCenter/ESXi username entry")
	}

	if appConfig.Infra.Vcenter.Password == "" {
		return false, errors.New("missing vCenter/ESXi password entry")
	}

	if appConfig.Infra.Nsxt.Hostname == "" {
		return false, errors.New("missing NSX-T Manager hostname")
	}

	if appConfig.Infra.Nsxt.Username == "" {
		return false, errors.New("missing NSX-T Manager username")
	}

	if appConfig.Infra.Nsxt.Password == "" {
		return false, errors.New("missing NSX-T Manager password")
	}

	return true, nil
}

// Read config.yml file and serialize everything in AppConfig struct.
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

	if appConfig.Infra.Nsxt.Password == "*" {
		for ok := true; ok; ok = !(len(appConfig.Infra.Nsxt.Password) > 1) {
			fmt.Print("NSX-T password: ")
			_, err = fmt.Scanln(&appConfig.Infra.Nsxt.Password)
			if err != nil {
				ok = false
			}
		}
	}

	return appConfig, nil
}