package internal

import (
	"context"
	"fmt"
	"github.com/spyroot/jettison/src/config"
	"github.com/spyroot/jettison/src/vcenter"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"log"
)

type Vim struct {
	VimFinder *find.Finder
	VimApi    *govmomi.Client
	NsxApi    nsxt.APIClient
	AppConfig *config.AppConfig
}

// Returns vim client
func (v Vim) GetVimClient() *vim25.Client {
	return v.VimApi.Client
}

func GetVm(ctx context.Context, uuid string, vim *Vim) (*object.VirtualMachine, error) {

	vm, err := vcenter.GetVM(ctx, vim.VimApi.Client, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve vm %s details from infrastracture", uuid)
	}

	vimPath := vim.AppConfig.GetDc().InventoryPath + "/*/" + vm.Summary.Config.Name
	log.Println("Searching vm ", vm.Summary.Config.Name, " in path:", vimPath)
	v, err := vim.VimFinder.VirtualMachine(ctx, vimPath)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve vm %s details from infrastracture", err)
	}

	return v, nil
}

// Checks that vm attached to a target logical switch.
func CheckNetwork(ctx context.Context, vimClient *vim25.Client, c *config.AppConfig, vmName string) error {
	net, err := vcenter.GetNetworks(ctx, vimClient, c.Infra.Nsxt.LogicalSwitch)
	if err != nil {
		return fmt.Errorf("invalid logical switch: %s", c.Infra.Nsxt.LogicalSwitch)
	}

	for _, h := range net.Vm {
		if vmName == h.Value {
			return nil
		}
	}

	return fmt.Errorf("VM reference not found: %s", vmName)
}
