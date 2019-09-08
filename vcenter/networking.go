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

networking related staff

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package vcenter

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"strings"
)

type NetworkAdapter struct {
	connected  bool
	switchUuid string
	switchName string
	deviceKey  string
}

func GetNetworkAttr(ctx context.Context, c *vim25.Client, vmVimName string) (*[]mo.Network, error) {

	log.Println("Looking network for vm", vmVimName)
	listNework := []mo.Network{}

	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		return nil, fmt.Errorf("failed create network view: %s", err)
	}

	defer v.Destroy(ctx)

	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve network list: %s", err)
	}

	for _, net := range networks {
		for _, vm := range net.Vm {
			if vm.Value == vmVimName {
				listNework = append(listNework, net)
			}
		}
	}

	return &listNework, nil
}

/**
  Function lookup network based on reference
*/
func GetNetworkRef(ctx context.Context, c *vim25.Client, uuid string) ([]string, error) {

	var networksRefs []string

	got, err := GetVmAttr(ctx, c, VmLookupMap["name"], uuid)
	if err != nil {
		return networksRefs, fmt.Errorf("object name failed %s", err)
	}

	if got != nil {
		vm := *got
		for _, v := range vm.Network {
			networksRefs = append(networksRefs, v.Reference().String())
		}
	}

	return networksRefs, err
}

/**
  Function adds network adapter to existing vm
*/
func AddNetworkAdapter(ctx context.Context, c *vim25.Client, networkName string, vmName string) error {

	net, err := find.NewFinder(c).Network(ctx, networkName)
	if err != nil {
		return err
	}

	vm, err := find.NewFinder(c).VirtualMachine(ctx, "jettison-test")
	if err != nil {
		return err
	}

	backing, err := net.EthernetCardBackingInfo(ctx)
	if err != nil {
		return err
	}

	device, err := object.EthernetCardTypes().CreateEthernetCard("", backing)
	if err != nil {
		return err
	}

	return vm.AddDevice(ctx, device)
}

/**
  Function check if switch is Opaque switch
*/
func isOpaqueNetwork(ctx context.Context, c *vim25.Client, switchName string) (bool, error) {

	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		return false, err
	}

	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		return false, fmt.Errorf("failed retrieve network list: %s", err)
	}

	for _, net := range networks {
		// type for NSX_T OpaqueNetwork
		if net.Name == switchName {
			if net.Summary.GetNetworkSummary().Network.Type == "OpaqueNetwork" {
				return true, nil
			}
		}
	}

	return false, nil
}

/**

 */
func GetSwitchUuid(ctx context.Context, c *vim25.Client, vmName string) ([]NetworkAdapter, error) {

	var adapters []NetworkAdapter

	vm, err := find.NewFinder(c).VirtualMachine(ctx, vmName)
	if err != nil {
		return adapters, err
	}

	devList, err := vm.Device(ctx)
	if err != nil {
		return adapters, err
	}

	for _, v := range devList {
		vdev := v.GetVirtualDevice()
		if vdev != nil && vdev.DeviceInfo.GetDescription() != nil {
			vdevSummary := v.GetVirtualDevice().DeviceInfo.GetDescription().Summary
			if strings.Contains(vdevSummary, "nsx.LogicalSwitch") {
				opaqueData := strings.Split(vdevSummary, ":")
				if len(opaqueData) > 0 {
					adapters = append(adapters, NetworkAdapter{
						connected:  vdev.Connectable.Connected,
						switchUuid: opaqueData[1],
						switchName: "",
						deviceKey:  vdev.DeviceInfo.GetDescription().Label,
					})
				}
			}
		}
	}

	return adapters, nil
}
