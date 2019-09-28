package jettypes

import (
	"strings"
)

type VimEndpoint interface {
	Endpoint() string
	VimUsername() string
	VimPassword() string
	VimDatacenter() string
}

type PowerState int

const (
	PowerOn PowerState = iota
	PowerOff
	Reboot
	Reset
)

type VimPlugin interface {
	// entry point fo plugin used by vim to load a plugin
	// TODO remove argument plugin must do own configuration mgmt same as per nsx
	InitPlugin(VimEndpoint) error

	// connect vm to a switch, it should create adapter if no adapter present
	// vim use a switch that auto discovered
	ConnectVm(projectName string, node *NodeTemplate) (bool, error)

	DisconnectVm(projectName string, node *NodeTemplate) (bool, error)

	// plug implementation need provide semantics to cleanup all stale object
	ComputeCleanup(projectName string, nodes []*NodeTemplate) error

	//
	DhcpCleanup(projectName string, nodes []*NodeTemplate) error

	// network interface
	DeploySegment(projectName string, segmentName string, gateway string, prefixLen int) (*GenericSwitch, *GenericRouter, error)

	//
	CreateDhcpBindings(projectName string, nodes []*NodeTemplate) error

	// discovery cluster dhcp
	DiscoverClusterDhcpServer(projectName string, nodes *[]*NodeTemplate) (bool, error)

	// discovery vm template
	DiscoverVmTemplate(node *NodeTemplate) error

	// discovery vm
	DiscoverVms(projectName string, nodes []*NodeTemplate) error

	// clone group of vm from nodes
	CloneVms(projectName string, nodes []*NodeTemplate) error

	// change vm power state
	ChangePowerState(node *NodeTemplate, state PowerState) (bool, error)

	// acquire vm ip address
	AcquireIpAddress(node *NodeTemplate) (bool, string, error)

	DeleteDhcpServer(node *NodeTemplate) (bool, error)

	DeleteRouter(node *NodeTemplate) (bool, error)

	DeleteSwitch(node *NodeTemplate) (bool, error)

	AddStaticRoute(projectName string, node *NodeTemplate, podNetwork string) (bool, error)
}

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

	s := strings.ToLower(nodeType)
	switch s {
	case "controller":
		return ControlType
	case "worker":
		return WorkerType
	case "ingress":
		return IngressType
	case "template":
		return TemplateType
	default:
	}

	return Unknown
}

func (n NodeType) String() string {

	names := [...]string{
		"Controller",
		"Worker",
		"Ingress",
		"Template",
		"Unknown",
	}
	return names[n]
}

/* node type */
type DeployerCmd int

const (
	// action to re-generate ansible playbook
	AnsiblePlaybook DeployerCmd = iota

	// re-generate  ansible files that include roles, templates
	AnsibleFiles

	// re-generate ansible inventory file
	AnsibleInventory

	// run ansible playbook
	AnsibleDeploy

	// re-run ansible playbook
	AnsibleRun

	// re-deploy networking
	VimNetworking

	// re-deploy
	VimCompute
)

func (n DeployerCmd) String() string {

	names := [...]string{
		"AnsiblePlaybook",
		"AnsibleFiles",
		"AnsibleInventory",
		"VimNetworking",
		"VimCompute",
		"Unknown",
	}

	return names[n]
}
