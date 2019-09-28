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

vSphere routines low level primitives

Author spyroot
mbaraymov@vmware.com
*/
package vcenter

import (
	"context"
	"fmt"
	"github.com/spyroot/jettison/logging"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
)

/*
  Functions return this error in case VM not found
*/
type VmNotFound struct {
	msg string
}

func (e *VmNotFound) Error() string {
	return e.msg
}

func NewVmNotFound() error {
	return &VmNotFound{"vm not found"}
}

/*
  Functions return this error if network not found
*/
type NetworkNotFound struct {
	msg string
}

func (error *NetworkNotFound) Error() string {
	return error.msg
}
func NewNetworkNotFound() error {
	return &NetworkNotFound{"network not found"}
}

/*
  Functions return this error if cluster not found
*/
type ClusterNotFound struct {
	msg string
}

func (error *ClusterNotFound) Error() string {
	return error.msg
}
func NewClusterNotFound() error {
	return &ClusterNotFound{"cluster not found"}
}

// hold a comparator map.  Caller can pass own function or use existing
// VmComparator["name"]
type VmComparator func(*types.VirtualMachineSummary, string) bool

var VmSearchHandler = map[string]func(vmSummary *types.VirtualMachineSummary, val string) bool{
	"name": func(vmSummary *types.VirtualMachineSummary, val string) bool {
		if vmSummary.Config.Name == val {
			return true
		}
		return false
	},
	"uuid": func(vmSummary *types.VirtualMachineSummary, val string) bool {
		if vmSummary.Config.Uuid == val {
			return true
		}
		return false
	},
	"vimname": func(vmSummary *types.VirtualMachineSummary, val string) bool {
		if vmSummary.Vm.Value == val {
			return true
		}
		return false
	},
}

/**

 */
func GetNetworks(ctx context.Context, c *vim25.Client, name string) (*mo.Network, error) {

	if c == nil {
		return nil, fmt.Errorf("vim client is nil")
	}

	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	viewManager := view.NewManager(c)
	v, err := viewManager.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		return nil, err
	}

	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		return nil, err
	}

	for _, net := range networks {
		if net.Name == name {
			return &net, nil
		}
	}

	return nil, NewVmNotFound()
}

/**

 */
func deleteFolder(ctx context.Context, c *vim25.Client, name string) (*mo.Network, error) {

	if c == nil {
		return nil, fmt.Errorf("vim client is nil")
	}

	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	viewManager := view.NewManager(c)
	v, err := viewManager.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		return nil, err
	}

	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		return nil, err
	}

	for _, net := range networks {
		if net.Name == name {
			return &net, nil
		}
	}

	return nil, fmt.Errorf("network not found")
}

/**
  Function return summary information.
  Client can access value summary.config
*/
func GetSummaryVm(ctx context.Context, c *vim25.Client, uuid string) (*mo.VirtualMachine, error) {

	if c == nil {
		return nil, fmt.Errorf("vim client is nil")
	}

	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	viewManager := view.NewManager(c)
	v, err := viewManager.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, fmt.Errorf("failed create a view: %s", err)
	}

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "config"}, &vms)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve vm list: %s", err)
	}

	for _, vm := range vms {
		log.Println(vm.Summary.Config.Name, vm.Summary.Config.Uuid, uuid)
		if vm.Summary.Config.Uuid == uuid {
			return &vm, nil
		}
	}

	return nil, &VmNotFound{"vm not found"}
}

// Return virtual
func ChangePowerState(ctx context.Context, c *vim25.Client, uuid string) (*mo.VirtualMachine, error) {

	m := view.NewManager(c)

	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, fmt.Errorf("failed create a view: %s", err)
	}

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve vm list: %s", err)
	}

	for _, vm := range vms {
		if vm.Summary.Config.InstanceUuid == uuid {
			//			TODO
		}
	}
	return nil, fmt.Errorf("vm not found")
}

//
// Function take callback and pass to it vm.Summary data.
// it return VM data,  to get vim name we can use for example
// vm.Summary.Vm.Value
//
func GetVmAttr(ctx context.Context, c *vim25.Client,
	vmComparator VmComparator, attrValue string) (*mo.VirtualMachine, error) {

	if c == nil {
		return nil, fmt.Errorf("vim client is nil")
	}

	if vmComparator == nil {
		return nil, fmt.Errorf("callback is nil")
	}

	viewManager := view.NewManager(c)
	v, err := viewManager.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, fmt.Errorf("failed create a view: %s", err)
	}

	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "config", "network"}, &vms)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve vm list: %s", err)
	}

	var m mo.VirtualMachine
	for i, vm := range vms {
		if vmComparator(&vm.Summary, attrValue) {
			m = vms[i]
			break
			//			return &vm[i], nil
		}
	}

	return &m, nil
}

//
//   Function returns a VirtualMachine, Host and Cluster where it deployed.
//
func VmFromCluster(ctx context.Context, c *vim25.Client, vmName string, clusterName string) (
	*object.HostSystem, *object.ComputeResource, *object.VirtualMachine, error) {

	if c == nil {
		return nil, nil, nil, fmt.Errorf("vim client is nil")
	}

	vm, err := find.NewFinder(c).VirtualMachine(ctx, vmName)
	if err != nil {
		return nil, nil, nil, &VmNotFound{"vm not found"}
	}

	crs, err := find.NewFinder(c).ComputeResource(ctx, clusterName)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, fmt.Errorf("compute resource not found %s", err)
	}

	clusterHost, err := crs.Hosts(ctx)
	if err != nil {
		log.Println(err)
		return nil, nil, nil, fmt.Errorf("host resource not found %s", err)
	}

	host, err := vm.HostSystem(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("host resource not found %s", err)
	}

	esxiName, err := host.ObjectName(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("object name failed %s", err)
	}

	for _, cHost := range clusterHost {
		if cHost.Name() == esxiName {
			return cHost, crs, vm, nil
		}
	}

	return nil, nil, nil, &VmNotFound{"vm not found"}
}

//  Function return data store for a given cluster and data center.
//
func FindDatastore(ctx context.Context, c *vim25.Client, clusterName string, dsName string) (*object.Datastore, error) {

	if c == nil {
		return nil, fmt.Errorf("vim client is nil")
	}

	crs, err := find.NewFinder(c).ComputeResource(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("compute resource not found %s", err)
	}

	ds, err := crs.Datastores(ctx)
	if err != nil {
		return nil, fmt.Errorf("no data store found under cluster%s", err)
	}

	for _, v := range ds {
		n, err := v.ObjectName(ctx)
		if err != nil {
			return nil, err
		}
		if n == dsName {
			return v, nil
		}
	}

	return nil, fmt.Errorf("data store not found")
}

/**
  Function deletes folder from Data Center.  Note that folder can't have running VM
  Caller need stop all VM inside a folder.
*/
func DeleteFolder(ctx context.Context, c *vim25.Client, folderName string) error {

	if c == nil {
		return fmt.Errorf("vim client is nil")
	}

	if len(folderName) == 0 {
		return nil
	}

	folder, err := find.NewFinder(c).Folder(ctx, folderName)
	if err != nil {
		newErr := fmt.Errorf("compute resource not found %s", err)
		logging.ErrorLogging(newErr)
		return newErr
	}

	if folder.Reference().Type == "Folder" {
		task, err := folder.Destroy(ctx)
		if err != nil {
			return nil
		}

		_, err = task.WaitForResult(context.Background(), nil)
		if err != nil {
			newErr := fmt.Errorf("WaitForResult for folder delete failed %s", err)
			return newErr
		}
	}

	return nil
}
