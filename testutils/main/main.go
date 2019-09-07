package main

import (
	"github.com/spyroot/jettison/internal"
	"github.com/spyroot/jettison/testutils"
	"log"
)

func main() {

	vim, err := internal.InitEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	//
	//ret, err := vcenter.GetSwitchUuid(vim.Ctx, vim.GetVimClient(), "jettison-test")
	//log.Println(ret)
	//
	uuid, err := testutils.CreateVmifneeded(vim.Ctx, vim.GetVimClient())
	if err != nil {
		log.Fatal(err)
	}

	log.Println(uuid)

	//nets, err := find.NewFinder(vim.GetVimClient()).NetworkList(vim.Ctx, "test-segment")
	//for _, n := range nets {
	//	log.Println(n)
	//}
	//
	//net, err := find.NewFinder(vim.GetVimClient()).Network(vim.Ctx, "test-segment")
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//
	//vm, err := find.NewFinder(vim.GetVimClient()).VirtualMachine(vim.Ctx, "jettison-test")
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//devList, _ := vm.Device(vim.Ctx)
	//if err != nil {
	//	log.Println(err)
	//}
	//
	//for _, v := range devList {
	//	if strings.Contains(v.GetVirtualDevice().DeviceInfo.GetDescription().Summary, "nsx.LogicalSwitch") {
	//		log.Println(v.GetVirtualDevice().DeviceInfo)
	//		log.Println(v.GetVirtualDevice().DeviceInfo.GetDescription().GetDescription().Label)
	//		stringSlice := strings.Split(v.GetVirtualDevice().DeviceInfo.GetDescription().Summary, ":")
	//		if len(stringSlice) > 0 {
	//			log.Println("NSX-T switch UUID", stringSlice[1])
	//		}
	//	}
	//}
	//
	//log.Println(vm.Device(vim.Ctx))
	//log.Println(net.Reference())

}
