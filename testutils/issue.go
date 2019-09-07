package testutils

import (
	"context"
	"fmt"
	"github.com/spyroot/jettison/internal"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"time"
)

func TestIssue242() {

	vim, err := internal.InitEnvironment()
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithCancel(context.Background())
	//defer cancel()

	h := NewHelper(vim.GetVimClient())
	df, err := h.Datacenter().Folders(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cr := h.ComputeResource("mgmt")

	// Get local datastores for compute resource
	dss, err := h.LocalDatastores(ctx, cr)
	if err != nil {
		log.Fatal(err)
	}
	if len(dss) == 0 {
		log.Fatalf("No local datastores")
	}

	// Get root resource pool for compute resource
	rp, err := cr.ResourcePool(ctx)
	if err != nil {
		log.Fatal(err)
	}

	spec := types.VirtualMachineConfigSpec{
		Name:     fmt.Sprintf("govmomi-test-%s", time.Now().Format(time.RFC3339)),
		Files:    &types.VirtualMachineFileInfo{VmPathName: fmt.Sprintf("[%s]", dss[0].Name())},
		NumCPUs:  1,
		MemoryMB: 32,
	}

	// Create new VM
	task, err := df.VmFolder.CreateVM(context.Background(), spec, rp, nil)
	if err != nil {
		log.Fatal(err)
	}

	info, err := task.WaitForResult(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	vm := object.NewVirtualMachine(h.c, info.Result.(types.ManagedObjectReference))
	defer func() {
		task, err := vm.Destroy(context.Background())
		if err != nil {
			panic(err)
		}
		err = task.Wait(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	// Mark VM as template
	err = vm.MarkAsTemplate(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Get "environmentBrowser" property for VM template
	var mvm mo.VirtualMachine
	err = property.DefaultCollector(h.c).RetrieveOne(ctx, vm.Reference(), []string{"environmentBrowser"}, &mvm)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}
