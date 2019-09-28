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

Basic helpers for unit testing.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package internal

import (
	"context"
	"github.com/google/uuid"
	"github.com/spyroot/jettison/jettypes"
	"log"

	"math/rand"
	"net"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
)

type TestingEnv struct {
	ctx           context.Context
	TestImageName string
	TestImageId   string
	TestVim       *Vim
}

/**
  Unit test helper init environment, open up all required connection
  Create single test image in vcenter for a unit tests.
*/
func VimSetupHelper() *TestingEnv {

	return &TestingEnv{nil, TestImageName, nil, nil}
}

var (
	// path to a file
	TestImage string = "/Users/spyroot/go/tests/ttylinux-disk001.vmdk"
	// cluster name
	TestImageCluster string = "mgmt"
	// data store attached to a cluster
	TestDataStoreName string = "vsanMgmtDatastore"
	// name of vm
	TestImageName = "jettison-test"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func TestTemplate() *jettypes.NodeTemplate {

	var node0 = &jettypes.NodeTemplate{}
	node0.DesiredAddress = "172.16.81.0/24"
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.VimCluster = uuid.New().String()
	node0.VmTemplateName = TestImageName
	node0.VimCluster = TestImageCluster
	node0.Type = jettypes.ControlType

	return node0
}

func TestTemplateBogus01() *jettypes.NodeTemplate {

	var node0 = &jettypes.NodeTemplate{}
	node0.DesiredAddress = "172.16.81.0/24"
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.VimCluster = uuid.New().String()
	node0.VmTemplateName = ""
	node0.VimCluster = TestImageCluster
	node0.Type = jettypes.ControlType

	return node0
}

func CreateSyntheticValidNode(t *testing.T) *jettypes.NodeTemplate {

	var node0 = &jettypes.NodeTemplate{}

	node0.Name = uuid.New().String()
	node0.DesiredAddress = "172.16.81.0/24"
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.VimCluster = uuid.New().String()
	node0.VmTemplateName = uuid.New().String()
	node0.Mac = append(node0.Mac, "abcd:abcd:abc")

	VimName := uuid.New().String()
	node0.SetVimName(VimName)
	assert.Equal(t, VimName, node0.GetVimName(), "switch uuid mismatch")

	node0.SetFolderPath("/Datacenter/vms")
	assert.Equal(t, "/Datacenter/vms", node0.GetFolderPath(), "folder mismatch")

	switchUuid := uuid.New().String()
	node0.GenericSwitch().SetUuid(switchUuid)
	assert.Equal(t, switchUuid, node0.GenericSwitch().Uuid(), "switch uuid mismatch")

	routerUuid := uuid.New().String()
	node0.GenericRouter().SetUuid(routerUuid)
	assert.Equal(t, routerUuid, node0.GenericRouter().Uuid(), "router uuid mismatch")

	dhcpID := uuid.New().String()
	node0.GenericSwitch().SetUuid(dhcpID)
	assert.Equal(t, dhcpID, node0.GenericSwitch().Uuid(), "dhcp id uuid mismatch")

	node0.VimCluster = "mgmt"

	node0.Type = jettypes.ControlType

	return node0
}

// create syntetical node
func CreateSyntheticValidNodes(t *testing.T, numNodes int) *[]*jettypes.NodeTemplate {

	var nodes []*jettypes.NodeTemplate
	for i := 0; i < numNodes; i++ {
		n := CreateSyntheticValidNode(t)
		nodes = append(nodes, n)
	}

	return &nodes
}
