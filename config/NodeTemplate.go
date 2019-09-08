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

Jettison main config parser populate slices of node template.
A node template a data structure that hold a date related to a template vm
deployed in the VIM.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package config

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net"
	"strings"
)

/* a node status */
type Status int

const (
	Undefined Status = iota
	Created
	Deleted
	Error
)

/* node type */
type NodeType int

const (
	//Template
	ControlType NodeType = iota

	//Template

	WorkerType
	//Template

	IngressType

	//Template
	TemplateType

	//
	Unknown
)

/**
Function return string representation of node type
*/
func GetNodeType(nodeType string) NodeType {

	switch nodeType {
	case "Controller":
		return ControlType
	case "Worker":
		return WorkerType
	case "Ingress":
		return IngressType
	case "Template":
		return TemplateType
	default:
	}

	return Unknown
}

func (n NodeType) String() string {

	names := [...]string{
		"Controller",
		"Worker",
		"IngressType",
		"TemplateType",
	}
	return names[n]
}

type NodeTemplate struct {
	Name           string
	Prefix         string `yaml:"prefix"`
	DomainSuffix   string `yaml:"domainSuffix"`
	DesiredCount   int    `yaml:"desiredCount"`
	DesiredAddress string `yaml:"desiredAddress"`
	IPv4AddrStr    string `yaml:"IPv4address"`
	IPv4Addr       net.IP
	IPv4Net        *net.IPNet
	Gateway        string `yaml:"gateway"`
	VmTemplateName string `yaml:"vmTemplateName"`
	UUID           string `yaml:"uuid"`
	VimCluster     string `yaml:"clusterName"`
	NetworksRef    []string
	Mac            []string
	VimName        string

	LogicalSwitch string `yaml:"logicalSwitch"`

	// nsx-t switch
	GenericSwitch struct {
		// switch uuid
		Uuid string
		// dhcp attached to this logical switch
		DhcpUuid string
		// name of dhcp
		DhcpName string
	}

	GenericRouter struct {
		Name string `yaml:"logicalRouter"`
		Uuid string
	}

	Type          NodeType
	VimState      Status // TODO move to a map no need hold separate fields
	DhcpStatus    Status
	NetworkStatus Status
	AnsibleStatus Status
	// Added
	FolderPath string
}

// Sets router name
func (node *NodeTemplate) SetRouterName(name string) {
	node.GenericRouter.Name = name
}

// return router name
func (node *NodeTemplate) GetRouterName() string {
	return node.GenericRouter.Name
}

// sets router uuid
func (node *NodeTemplate) SetRouterUuid(uuid string) {
	node.GenericRouter.Uuid = uuid
}

// return router uuid.  during discovery process jettison will set that
// to actual value
func (node *NodeTemplate) GetRouterUuid() string {
	return node.GenericRouter.Uuid
}

// Sets logical switch name
func (node *NodeTemplate) SetSwitchName(name string) {
	node.LogicalSwitch = name
}

// return logical switch name
func (node *NodeTemplate) GetSwitchName() string {
	return node.LogicalSwitch
}

// sets switch uuid
func (node *NodeTemplate) SetSwitchUuid(uuid string) {
	node.GenericSwitch.Uuid = uuid
}

// return switch uuid.  during discovery process jettison will set that
// to actual value
func (node *NodeTemplate) GetSwitchUuid() string {
	return node.GenericSwitch.Uuid
}

//  setgs dhcp server uuid that must attached to logical switch
func (node *NodeTemplate) SetDhcpId(dhcpId string) {
	node.GenericSwitch.DhcpUuid = dhcpId
}

// returns dhcp server uuid attached to logical switch
func (node *NodeTemplate) GetDhcpId() string {
	return node.GenericSwitch.DhcpUuid
}

//  setgs dhcp server uuid that must attached to logical switch
func (node *NodeTemplate) SetDhcName(name string) {
	node.GenericSwitch.DhcpName = name
}

// returns dhcp server uuid attached to logical switch
func (node *NodeTemplate) GetDhcpName() string {
	return node.GenericSwitch.DhcpName
}

/**

 */
func (node NodeTemplate) Clone() *NodeTemplate {
	newNode := node
	return &newNode
}

/**
Generates a name for a node in format prefix.uuid.sufix and sets the name
*/
func (node *NodeTemplate) GenerateName() {
	folderUuid := uuid.New()
	node.Name = fmt.Sprint(node.Prefix, ".", folderUuid.String(), ".", node.DomainSuffix)
}

func (node *NodeTemplate) SetFolderPath(path string) {
	node.FolderPath = path
}

func (node *NodeTemplate) GetFolderPath() string {
	return node.FolderPath
}

func (node *NodeTemplate) GetVimName() string {
	return node.VimName
}

func (node *NodeTemplate) SetVimName(vimName string) {

	tmpName := vimName
	if strings.Contains(vimName, "VirtualMachine") {
		var s = strings.Split(vimName, ":")
		if len(s) > 0 {
			tmpName = s[1]
		}
	}

	node.VimName = tmpName
}

func (node *NodeTemplate) PrintAsJson() {
	var p []byte
	p, err := json.MarshalIndent(node, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}
