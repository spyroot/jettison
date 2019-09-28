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

Jettison main config parser populate slices of node template.
A node template a data structure that hold a date related to a template vm
deployed in the VIM.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package jettypes

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
)

/* a node status */
type Status int

const (
	Undefined Status = iota
	Created
	Deleted
	Error
)

type NodeTemplate struct {
	Name           string
	Prefix         string `yaml:"prefix"`
	DomainSuffix   string `yaml:"domainSuffix"`
	DesiredCount   int    `yaml:"desiredCount"`
	DesiredAddress string `yaml:"desiredAddress"`
	IPv4AddrStr    string `yaml:"IPv4address"`
	Gateway        string `yaml:"gateway"`
	VmTemplateName string `yaml:"vmTemplateName"`
	UUID           string `yaml:"uuid"`
	VimCluster     string `yaml:"clusterName"`
	EdgeCluster    string `yaml:"edgeCluster"`
	Static         bool   `yaml:"static"`

	IPv4Addr net.IP
	IPv4Net  *net.IPNet

	NetworksRef []string
	Mac         []string
	VimName     string

	LogicalSwitch string `yaml:"logicalSwitch"`

	// flag indicate whether jettison need use existing network
	existingNetwork bool

	genericSwitch *GenericSwitch
	genericRouter *GenericRouter

	Type          NodeType
	VimState      Status // TODO move to a map no need hold separate fields
	DhcpStatus    Status
	NetworkStatus Status
	AnsibleStatus Status

	// folder where node is deployed
	FolderPath string

	// each worker node get own cidr block and size indicate block size 24, 25 etc
	podCidr           string
	podAllocationSize int

	template bool
}

func (node *NodeTemplate) Template() bool {
	return node.template
}

func (node *NodeTemplate) SetTemplate(template bool) {
	node.template = template
}

func (node *NodeTemplate) IsTemplate() bool {
	if node != nil {
		return node.template
	}
	return false
}

/* */
func (node *NodeTemplate) IsExistingNetwork() bool {
	if node != nil {
		return node.existingNetwork
	}
	return false
}

func (node *NodeTemplate) SetExistingNetwork(flag bool) {
	if node != nil {
		node.existingNetwork = flag
	}
}

/*

 */
func (node *NodeTemplate) DhcpServerUuid() string {
	if node != nil {
		return node.GenericSwitch().DhcpUuid()
	}
	return ""
}

func (node *NodeTemplate) RouterUuid() string {
	if node != nil {
		return node.GenericRouter().Uuid()
	}
	return ""
}

/*
  mac[0] should be a primary mac address
*/
func (node *NodeTemplate) MacAddress() string {
	if node != nil {
		return node.Mac[0]
	}

	return ""
}

/*

 */
func (node *NodeTemplate) IPv4Address() net.IP {
	if node != nil {
		return node.IPv4Addr
	}

	return nil
}

/*

 */
func (node *NodeTemplate) SwitchUuid() string {
	if node != nil {
		return node.GenericSwitch().Uuid()
	}
	return ""
}

/*

 */
func (node *NodeTemplate) GenericRouter() *GenericRouter {
	if node != nil {
		return node.genericRouter
	}
	return &GenericRouter{}
}

/*

 */
func (node *NodeTemplate) SetGenericRouter(genericRouter *GenericRouter) {
	if node != nil {
		node.genericRouter = genericRouter
	}
}

/*

 */
func (node *NodeTemplate) GenericSwitch() *GenericSwitch {
	if node != nil {
		return node.genericSwitch
	}
	return &GenericSwitch{}
}

/*

 */
func (node *NodeTemplate) SetGenericSwitch(genericSwitch *GenericSwitch) {
	if node != nil {
		node.genericSwitch = genericSwitch
	}
}

/*

 */
func (node *NodeTemplate) isWorker() bool {
	if node != nil {
		if node.Type == WorkerType {
			return true
		}
	}
	return false
}

/*

 */
func (node *NodeTemplate) isController() bool {
	if node != nil {
		if node.Type == ControlType {
			return true
		}
	}
	return false
}

/*

 */
func (node *NodeTemplate) isIngress() bool {
	if node != nil {
		if node.Type == IngressType {
			return true
		}
	}
	return false
}

/*

 */
func (node *NodeTemplate) GetNodeType() NodeType {
	if node != nil {
		return node.Type
	}
	return Unknown
}

/*

 */
func (node *NodeTemplate) GetNodeTypeAsString() string {
	if node != nil {
		return node.Type.String()
	}
	return Unknown.String()
}

/*

 */
func (node *NodeTemplate) GetUuidName() string {
	if node != nil {
		return node.Name
	}
	return ""
}

// implements interface need generate certificate
func (node *NodeTemplate) GetIpAddress() string {
	if node != nil {
		return node.IPv4AddrStr
	}
	return ""
}

// implements interface need generate certificate
func (node *NodeTemplate) GetHostname() string {
	if node != nil {
		return node.Name
	}
	return ""
}

/**
  Clone existing template
*/
func (node NodeTemplate) Clone() *NodeTemplate {
	newNode := node
	return &newNode
}

/**
  Generates a name for a node in format prefix.uuid.suffix and sets the name
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

/* Set a name how it stored in the vim for example in vCenter it vm-xxx */
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

	p, err = json.MarshalIndent(node.GenericSwitch(), "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}

/* Sets pod cidr */
func (node *NodeTemplate) SetPodCidr(cidr string) {
	if node != nil {
		node.podCidr = cidr
	}
}

func (node *NodeTemplate) PodAllocationSize(size int) {
	if node != nil {
		node.podAllocationSize = size
	}
}

func (node *NodeTemplate) GetCidr() string {
	if node != nil {
		return node.podCidr
	}
	return ""
}

/*
   A size of network segment for example express in number one bits
   i.e /24 indicate 24 bits is network and remaining bits (zero)
   for host ports.
*/
func (node *NodeTemplate) GetAllocation() int {
	if node != nil {
		return node.podAllocationSize
	}
	return 0
}

/*
  Each set of template can be in same or different network segment
*/
type NetworkSegment struct {
	segmentName string
	segments    []*NodeTemplate
}

/**

 */
func (n *NetworkSegment) Segments() []*NodeTemplate {
	if n != nil {
		return n.segments
	}
	return []*NodeTemplate{}
}

/**

 */
func (n *NetworkSegment) SetSegments(segments []*NodeTemplate) {
	if n != nil {
		n.segments = segments
	}
}

/**

 */
func (n *NetworkSegment) SegmentName() string {
	if n != nil {
		return n.segmentName
	}
	return ""
}

func (n *NetworkSegment) SetSegmentName(segmentName string) {
	if n != nil {
		n.segmentName = segmentName
	}
}

/*
  Contains a mapping of network to set of segment that share that network.
  For example
							0 - Segment 02
     172.16.0.1 --> Shared --
							1 - Segment-01
*/
type NetworkSegments struct {
	vertex map[string]*NetworkSegment
}

func NewNetworkSegments(templates map[string]*NodeTemplate) (*NetworkSegments, error) {

	if len(templates) == 0 {
		return nil, fmt.Errorf("node templates is empty")
	}

	s := &NetworkSegments{}
	s.build(templates)

	return s, nil
}

func (d *NetworkSegments) Segments() map[string]*NetworkSegment {
	if d != nil {
		return d.vertex
	}
	return map[string]*NetworkSegment{}
}

/*
   Build a segment map where each key is network address
   and value that will hold a map NetworkSegment.

   If set of tempalates share network than singe NetworkSegment
   hold a slice of all template that link to that network.
*/
func (d *NetworkSegments) build(templates map[string]*NodeTemplate) {

	if d.vertex == nil {
		d.vertex = make(map[string]*NetworkSegment)
	}

	var shared = 1
	for _, v := range templates {
		if v.IPv4Net == nil {
			continue
		}
		seg, ok := d.vertex[v.IPv4Net.String()]
		if !ok {
			// each node time has own network segment if and only if network they are using
			// distinct otherwise different type share network segment
			s := &NetworkSegment{}
			s.segmentName = v.Type.String()
			s.segments = append(s.segments, v)
			d.vertex[v.IPv4Net.String()] = s
		} else {
			// all shared segment identified as shared
			seg.segmentName = fmt.Sprintf("shared0%d", shared)
			seg.segments = append(seg.segments, v)
			shared++
		}
	}
}
