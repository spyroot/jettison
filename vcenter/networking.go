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

networking related staff

Author spyroot
mbaraymov@vmware.com
*/

package vcenter

import (
	"context"
	"fmt"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net"
	"strings"
)

type NetworkAdapter struct {
	connected  bool
	switchUuid string
	switchName string
	deviceKey  string
}

// Returns all network VM attach to
//
func GetNetworkAttr(ctx context.Context, c *vim25.Client, vmVimName string) (*[]mo.Network, error) {

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

	for i, _ := range networks {
		for _, vm := range networks[i].Vm {
			if vm.Value == vmVimName {
				listNework = append(listNework, networks[i])
			}
		}
	}

	return &listNework, nil
}

//
//  Function lookup network based on VM name, please not it not safe operation,
//  if you expecting vm name duplicate names.
//
func GetNetworkRef(ctx context.Context, c *vim25.Client, vmName string) ([]string, error) {

	var networksRefs []string

	got, err := GetVmAttr(ctx, c, VmSearchHandler["name"], vmName)
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

//
//  Function adds network adapter to existing VM and return
//  a new mac address allocated for VM and map that contain all
//  VirtualEthernetCard
//
func AddNetworkAdapter(ctx context.Context, c *vim25.Client,
	networkName string, vmName string) (string, map[string]*types.VirtualEthernetCard, error) {

	var (
		n   object.NetworkReference
		err error
	)

	// we either search by name or uuid of opaque switch
	if !nsxtapi.IsUuid(networkName) {
		n, err = find.NewFinder(c).Network(ctx, networkName)
		if err != nil {
			return "", nil, err
		}
	} else {
		opaque, err := GetOpaqueByUuid(ctx, c, networkName)
		if err != nil {
			return "", nil, fmt.Errorf("opaque network %s no found", networkName)
		}
		r, err := find.NewFinder(c).NetworkList(ctx, "*")
		for _, v := range r {
			if v.Reference().Value == opaque.Reference().Value {
				n = v
				break
			}
		}
	}

	vm, err := find.NewFinder(c).VirtualMachine(ctx, vmName)
	if err != nil {
		return "", nil, err
	}

	backing, err := n.EthernetCardBackingInfo(ctx)
	if err != nil {
		return "", nil, err
	}

	devices, err := vm.Device(ctx)
	if err != nil {
		return "", nil, err
	}

	// take old list of all adapters
	var nicList = make(map[string]*types.VirtualEthernetCard)
	for i := 0; i < len(devices); i++ {
		if devices.Type(devices[i]) == "ethernet" {
			card := devices[i].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
			m, err := net.ParseMAC(card.MacAddress)
			if err != nil {
				return "", nil, fmt.Errorf("failed parse mac address")
			}
			// note duplicate will be added once
			nicList[m.String()] = card
		}
	}

	// add adapter
	device, err := object.EthernetCardTypes().CreateEthernetCard("vmxnet3", backing)
	if err != nil {
		return "", nil, err
	}

	err = vm.AddDevice(ctx, device)
	if err != nil {
		log.Println("failed add adapter")
	}

	// https://github.com/vmware/govmomi/issues/1624
	//newCard, _ := device.(types.BaseVirtualEthernetCard)
	//	_, _ = NetworkAdapterList(ctx, c, vmName)
	//	log.Print(newCard.GetVirtualEthernetCard().MacAddress)

	devices, err = vm.Device(ctx)
	if err != nil {
		return "", nil, err
	}

	newMac := ""
	for i := 0; i < len(devices); i++ {
		if devices.Type(devices[i]) == "ethernet" {
			card := devices[i].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()

			m, err := net.ParseMAC(card.MacAddress)
			if err != nil {
				return "", nil, fmt.Errorf("failed parse mac address")
			}

			if _, ok := nicList[m.String()]; !ok {
				nicList[m.String()] = card
				newMac = m.String()
			}
		}
	}

	return newMac, nicList, nil
}

/**
  Function remove network adapter from a VM, caller needs
  to provide network adapter mac address.  VM must be in shutdown state.
*/
func DeleteNetworkAdapter(ctx context.Context, c *vim25.Client, vmName string, macAddr string) error {

	m, err := net.ParseMAC(macAddr)
	if err != nil {
		return fmt.Errorf("incorect mac address format")
	}

	vm, err := find.NewFinder(c).VirtualMachine(ctx, vmName)
	if err != nil {
		return err
	}

	devices, err := vm.Device(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < len(devices); i++ {
		if devices.Type(devices[i]) == "ethernet" {
			VmMacAddr := devices[i].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard().MacAddress
			vmMac, err := net.ParseMAC(VmMacAddr)
			if err != nil {
				return fmt.Errorf("failed parse mac address")
			}
			if m.String() == vmMac.String() {
				err := vm.RemoveDevice(ctx, true, devices[i])
				if err != nil {
					log.Println("failed remove adapter")
				}
			}
		}
	}

	return nil
}

/**
  Function return all virtual adapters for a given vmName
*/
func NetworkAdapterList(ctx context.Context, c *vim25.Client, vmName string) ([]*types.VirtualEthernetCard, error) {

	vm, err := find.NewFinder(c).VirtualMachine(ctx, vmName)
	if err != nil {
		return nil, err
	}

	devices, err := vm.Device(ctx)
	if err != nil {
		return nil, err
	}

	var nicList []*types.VirtualEthernetCard

	for i := 0; i < len(devices); i++ {
		if devices.Type(devices[i]) == "ethernet" {
			card := devices[i].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
			nicList = append(nicList, card)
		}
	}

	return nicList, nil
}

//
// Return vSphere network reference based on external opaque uuid
//
func GetOpaqueByUuid(ctx context.Context, c *vim25.Client, switchUuid string) (*mo.OpaqueNetwork, error) {

	if c == nil {
		return nil, fmt.Errorf("vsphere client is nil")
	}

	kind := []string{"OpaqueNetwork"}
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, kind, true)
	if err != nil {
		return nil, fmt.Errorf("failed create a view")
	}

	var networks []mo.OpaqueNetwork
	err = v.Retrieve(ctx, kind, []string{"summary"}, &networks)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, fmt.Errorf("failed retriev network details")
	}

	for i, net := range networks {
		summary := net.Summary.(*types.OpaqueNetworkSummary)
		if summary.OpaqueNetworkId == switchUuid {
			return &networks[i], nil
		}
	}

	return nil, fmt.Errorf("network not found")
}

/**
  Function check if switch is Opaque switch or not
*/
func isOpaqueNetwork(ctx context.Context, c *vim25.Client, switchName string) (bool, error) {

	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		logging.ErrorLogging(err)
		return false, err
	}

	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		logging.ErrorLogging(err)
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

/*
  Function returns all Opaque Networks as map where key uuid and value a name
*/
func GetOpaqueNetworks(ctx context.Context, c *vim25.Client) (map[string]string, error) {

	if c == nil {
		return nil, fmt.Errorf("vsphere client is nil")
	}

	var resultMap = map[string]string{}

	kind := []string{"OpaqueNetwork"}
	m := view.NewManager(c)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, kind, true)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}
	defer v.Destroy(ctx)

	var networks []mo.OpaqueNetwork
	err = v.Retrieve(ctx, kind, []string{"summary"}, &networks)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	for _, net := range networks {
		summary := net.Summary.(*types.OpaqueNetworkSummary)
		_, ok := resultMap[summary.OpaqueNetworkId]
		if !ok {
			resultMap[summary.OpaqueNetworkId] = net.Summary.GetNetworkSummary().Name
		}
	}

	return resultMap, nil
}
