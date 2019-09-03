package vcenter

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
)

// Function get VM network and lookup done based UUID
func GetNetworks(ctx context.Context, c *vim25.Client, uuid string) (*mo.Network, error) {

	viewManager := view.NewManager(c)
	v, err := viewManager.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		return nil, err
	}

	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Network.html
	var networks []mo.Network
	err = v.Retrieve(ctx, []string{"Network"}, nil, &networks)
	if err != nil {
		return nil, err
	}

	for _, net := range networks {
		if net.Name == uuid {
			return &net, nil
		}
	}

	return nil, fmt.Errorf("network not found")
}

// Function returns vm lookup done based on uuid
func GetVM(ctx context.Context, c *vim25.Client, uuid string) (*mo.VirtualMachine, error) {

	log.Println("Checking uuid", uuid)

	viewManager := view.NewManager(c)
	v, err := viewManager.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
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
			return &vm, nil
		}
	}

	return nil, fmt.Errorf("vm not found")
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
		}
	}
	return nil, fmt.Errorf("vm not found")
}
