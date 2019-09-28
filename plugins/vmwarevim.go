package main

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/vim25"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"

	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/spyroot/jettison/vcenter"
)

type TaskMessage struct {
	VmName string
	Status types.TaskInfoState
	Err    error
}

func newTaskMessage(s string, status types.TaskInfoState, err error) TaskMessage {

	return TaskMessage{
		Status: status,
		VmName: s,
		Err:    err,
	}
}

/*
   Main Vmware vCenter VIM implementation
*/
type VmwareVim struct {
	// compute layer
	vimApi *govmomi.Client

	VimFinder *find.Finder

	datacenter *object.Datacenter

	// root context
	ctx context.Context

	// Active nsx connection
	nsxApi nsxt.APIClient

	// nsxt config
	nsxtConfig *NsxtConfig

	dcName string
}

// Returns vim client
func (p *VmwareVim) GetNsx() *nsxt.APIClient {
	if p != nil {
		return &p.nsxApi
	}
	return nil
}

/*
   Discovers baseline network element. For nsx-t it transport zone
   edge cluster.
*/
func (p *VmwareVim) discoverNetwork() error {

	err := p.DiscoverNetworkElements()
	if err != nil {
		e := fmt.Errorf("failed discover nsx-t: error %v", err)
		logging.ErrorLogging(e)
		return e
	}

	// validate that discovery process able find a target transport zone
	if len(p.nsxtConfig.OverlayTransportUuid()) == 0 {
		return fmt.Errorf("failed discover overlay transpot")
	}
	if len(p.nsxtConfig.EdgeClusterUuid()) == 0 {
		return fmt.Errorf("failed discover overlay transpot")
	}

	log.Print("Discovered overlay transport zone\t", p.nsxtConfig.OverlayTransportUuid())
	log.Print("Discovered edge cluster\t\t\t\t", p.nsxtConfig.EdgeClusterUuid())

	return nil
}

// TODO add cluster
//
func (p *VmwareVim) discoverDatacenter() error {

	p.VimFinder = find.NewFinder(p.vimApi.Client, true)
	if p.VimFinder == nil {
		return fmt.Errorf("couldn't acquire finder object")
	}

	if len(p.dcName) != 0 {
		var err error
		p.datacenter, err = p.VimFinder.Datacenter(context.Background(), p.dcName)
		if err != nil {
			return fmt.Errorf("failed to get data center details check config and vim")
		}

		return nil
	}

	// last resort
	var err error
	p.datacenter, err = find.NewFinder(p.vimApi.Client).DefaultDatacenter(p.ctx)
	if err != nil {
		return fmt.Errorf("failed to get data center details check config and vim")
	}

	return nil
}

/*
  Initialize a plugin based on config passed from VIM
*/
func (p *VmwareVim) InitPlugin(vimEndpoint jettypes.VimEndpoint) error {

	if vimEndpoint == nil {
		return fmt.Errorf("can't initilize plugin with nil arguments")
	}

	log.Print("VMware vim loaded ",
		vimEndpoint.Endpoint(), " ", vimEndpoint.VimUsername())

	// open connection to vCenter or ESXi
	p.ctx = context.Background()
	vsphereClient, err := vcenter.Connect(p.ctx,
		vimEndpoint.Endpoint(),
		vimEndpoint.VimUsername(),
		vimEndpoint.VimPassword())
	if err != nil {
		return fmt.Errorf("failed to connect to vCenter. %s", err)
	}
	p.vimApi = vsphereClient

	log.Print("Vmware vim version: ", p.vimApi.Version)

	p.VimFinder = find.NewFinder(p.vimApi.Client, true)
	if p.VimFinder == nil {
		return fmt.Errorf("couldn't acquire finder")
	}

	p.datacenter, err = p.VimFinder.Datacenter(context.Background(), vimEndpoint.VimDatacenter())
	if err != nil {
		return fmt.Errorf("failed to get data center details check config and vim")
	}

	// read plugin configuration
	nsxConfig, err := NewNsxtConfig()
	if err != nil {
		return fmt.Errorf("failed to get data center details check config and vim")
	}
	p.nsxtConfig = nsxConfig

	// open connection to vCenter or NSX-T
	nsxtClient, nsxError := nsxtapi.Connect(p.nsxtConfig.Hostname(),
		p.nsxtConfig.Username(),
		p.nsxtConfig.Password())
	if nsxError != nil {
		return fmt.Errorf("failed to connect to nsx-t manager")
	}
	p.nsxApi = nsxtClient

	err = p.discoverNetwork()
	if err != nil {
		return fmt.Errorf("failed discover nsx-t network elements")
	}

	return nil
}

//
//  Main entry point for plugin
//
func Init() (jettypes.VimPlugin, error) {

	vmwareVim := &VmwareVim{}

	log.Print("Loaded")

	return vmwareVim, nil
}

// Returns vim client
func (p *VmwareVim) VimClient() *vim25.Client {
	if p != nil {
		if p.vimApi == nil {
			log.Print("Wrong state")
		}
		return p.vimApi.Client
	}
	return nil
}

/*
   Function check that switch indicate by switchName present in vCenter.
   TODO Add check based on UUID.  Not clear from Opaqu how to extract that.
*/
func checkNetworkAttachment(networks *[]mo.Network, switchName string) bool {

	for _, network := range *networks {
		if network.Name == switchName {
			return true
		}
	}
	return false
}

/**
  vCenter object finder in path
*/
func (p *VmwareVim) findVmObject(vmName string) (*object.VirtualMachine, error) {

	vimPath := p.datacenter.InventoryPath + "/*/" + vmName
	vm, err := p.VimFinder.VirtualMachine(p.ctx, vimPath)
	if err != nil {
		newErr := fmt.Errorf("failed retrieve vm %s details from infrastracture", err)
		logging.ErrorLogging(newErr)
		return nil, newErr
	}

	return vm, nil
}

/*
   Function searches a VM template and returns respected
   *object.VirtualMachine, *[]mo.Network,

   It sets a template node uuid to vim uuid value
*/
func (p *VmwareVim) discoverVmTemplate(node *jettypes.NodeTemplate) (*object.VirtualMachine, *[]mo.Network, error) {

	if node == nil {
		return nil, nil, fmt.Errorf("node is nil")
	}

	vmSummary, err := vcenter.GetVmAttr(p.ctx, p.VimClient(), vcenter.VmSearchHandler["name"], node.VmTemplateName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve vm template. err: %s", err)
	}

	if node.UUID == "" {
		node.UUID = vmSummary.Summary.Config.Uuid
	}

	vm, err := p.findVmObject(node.VmTemplateName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find vm template object. err: %s", err)
	}

	node.SetVimName(vm.Reference().Value)

	devs, err := vm.Device(p.ctx)
	if err != nil {
		return nil, nil, err
	}

	for _, dev := range devs {
		if nic, ok := dev.(types.BaseVirtualEthernetCard); ok {
			node.Mac = append(node.Mac, nic.GetVirtualEthernetCard().MacAddress)
		}
	}

	networks, err := vcenter.GetNetworkAttr(p.ctx, p.VimClient(), vm.Reference().Value)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find vm template networks, err: %s", err)
	}

	for _, net := range *networks {
		node.NetworksRef = append(node.NetworksRef, net.Name)
	}

	return vm, networks, nil
}

/**
  Function fetch a template VM from vCenter
  Note template must have ethernet adapter attached to correct logical switch.
*/
func (p *VmwareVim) DiscoverVmTemplate(node *jettypes.NodeTemplate) error {

	_, _, err := p.discoverVmTemplate(node)
	if err != nil {
		return err
	}

	return nil
}

/**
  Function fetch a template VM from vCenter
  Note template must have ethernet adapter attached to correct logical switch.
*/
func (p *VmwareVim) DiscoverVmTemplates(node *jettypes.NodeTemplate) (*object.VirtualMachine, *[]mo.Network, error) {

	if node == nil {
		return nil, nil, fmt.Errorf("node is nil")
	}

	vm, networks, err := p.discoverVmTemplate(node)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retieve template %s", err)
	}

	if len(*networks) == 0 {
		//		p.ConnectVm()
		return nil, nil, fmt.Errorf("vm has no network attached %s. "+
			"Please check that template has network adapter and attached to correct switch",
			node.GenericSwitch().Name())
	}

	// TODO do re-binding here
	// check template attached to correct logical switch
	if checkNetworkAttachment(networks, node.GenericSwitch().Name()) == false {
		return nil, nil, fmt.Errorf("vm has no network adapter attached to %s", node.GenericSwitch().Name())
	}

	return vm, networks, nil
}

//
func (p *VmwareVim) disconnectVm(vmName string, node *jettypes.NodeTemplate) (bool, error) {

	vm, err := find.NewFinder(p.VimClient()).VirtualMachine(p.ctx, vmName)
	if err != nil {
		return false, err
	}

	_, _ = p.ChangePowerState(node, jettypes.PowerOff)
	//	vm.PowerOff()

	devices, err := vm.Device(p.ctx)
	if err != nil {
		return false, err
	}

	for i := 0; i < len(devices); i++ {
		if devices.Type(devices[i]) == "ethernet" {
			card := devices[i].(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
			err = devices.Disconnect(card)
			if err != nil {
				logging.CriticalMessage("failed disconnect adapter")
			}
			logging.Notification("Disconnected and removing a device")
			err = vm.RemoveDevice(p.ctx, false, card)
			if err != nil {
				logging.CriticalMessage("failed remove adapter")
			}
		}
	}

	return true, nil
}

//
//
//
func (p *VmwareVim) DisconnectVm(projectName string, node *jettypes.NodeTemplate) (bool, error) {

	if node.IsTemplate() {
		return p.disconnectVm(node.VmTemplateName, node)
	}
	return p.disconnectVm(node.Name, node)
}

//
//  Function checks that vm attached to a target logical switch
//  if vm not found it returns false in case or problem with request error
//  otherwise true
func (p *VmwareVim) isAttached(vmName string, switchName string) (bool, error) {

	net, err := vcenter.GetNetworks(p.ctx, p.VimClient(), switchName)
	if err != nil {
		if _, ok := err.(*vcenter.VmNotFound); ok {
			return false, nil
		}
		return false, fmt.Errorf("invalid logical switch: %s", switchName)
	}

	for _, v := range net.Vm {
		if vmName == v.Value {
			return true, nil
		}
	}

	return false, nil
}

// clean up routine
func (p *VmwareVim) cleanupNode(projectName string,
	node *jettypes.NodeTemplate, folders map[string]bool) error {

	if len(node.GetFolderPath()) > 0 {
		folders[node.GetFolderPath()] = true
	}

	logging.Notification("Trying deleting vm", node.Name, " folder ", node.GetFolderPath())
	_, _, vm, err := vcenter.VmFromCluster(p.ctx, p.VimClient(), node.Name, node.VimCluster)
	if err != nil {
		return err
	}

	//acquire  create a task and block
	pState, err := vm.PowerState(context.Background())
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}
	// check power state
	if pState == types.VirtualMachinePowerStatePoweredOn {
		powerOfTask, err := vm.PowerOff(context.Background())
		if err != nil {
			logging.ErrorLogging(err)
			return nil
		}
		// wait for result and block
		_, err = powerOfTask.WaitForResult(context.Background(), nil)
		if err != nil {
			logging.ErrorLogging(err)
			return err
		}
	}
	// destroy vm and wait for result
	task, err := vm.Destroy(context.Background())
	if err != nil {
		logging.ErrorLogging(err)
		return nil
	}
	_, err = task.WaitForResult(context.Background(), nil)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}

	logging.Notification("VM successfully deleted ")

	return nil
}

// Clean up routine called from go routine
//
func (p *VmwareVim) cleanupNodeTask(projectName string,
	node *jettypes.NodeTemplate, folders map[string]bool,
	sem chan int, wg *sync.WaitGroup, statusChan chan TaskMessage) {

	defer wg.Done()
	sem <- 1

	err := p.cleanupNode(projectName, node, folders)
	if err != nil {
		select {
		case statusChan <- TaskMessage{
			VmName: node.Name,
			Status: types.TaskInfoStateError,
			Err:    fmt.Errorf("clone vm task failed")}:
			break
		default:
		}
	} else {
		statusChan <- TaskMessage{
			VmName: node.Name,
			Status: types.TaskInfoStateSuccess,
			Err:    fmt.Errorf("clone vm task failed")}
	}

	<-sem
}

/**
  vCenter Cleanup routine that tries to delete all object from old deployment.
  TODO move that to vim clean up routine and de-couple vCenter logic from deployer.
*/
func (p *VmwareVim) ComputeCleanup(projectName string, nodes []*jettypes.NodeTemplate) error {

	var (
		folders map[string]bool
	)

	folders = make(map[string]bool)
	sem := make(chan int, 3)

	var finalStatus []TaskMessage
	errChan := make(chan TaskMessage, len(nodes))

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for i, _ := range nodes {
		go p.cleanupNodeTask(projectName, nodes[i], folders, sem, &wg, errChan)
	}

	wg.Wait()
	close(errChan)
	for e := range errChan {
		finalStatus = append(finalStatus, e)
	}

	/** TODO refactor this check each status print green in console */
	success := 0
	for _, v := range finalStatus {
		if v.Status == types.TaskInfoStateSuccess {
			success++
		}
		log.Println("job", v.VmName, "\tstatus", v.Status)
	}

	if success == len(nodes) {
		logging.Notification("All vm destroyed")
	}

	// All VM deleted, cleanup all folders now.
	for f, _ := range folders {
		err := vcenter.DeleteFolder(p.ctx, p.VimClient(), f)
		if err != nil {
			logging.CriticalMessage("Deployment", projectName, " failed delete folder", f)
		}
	}

	return nil
}

/**
  TODO add timeout for a thread in context
*/
func (p *VmwareVim) runInstantiateTask(ctx context.Context, f *object.Folder,
	name string, template *object.VirtualMachine,
	sem chan int, wg *sync.WaitGroup, statusChan chan TaskMessage) {

	defer wg.Done()
	sem <- 1

	vmConfigSpec := types.VirtualMachineCloneSpec{}
	t, err := template.Clone(p.ctx, f, name, vmConfigSpec)
	if err != nil {
		statusChan <- TaskMessage{
			VmName: name,
			Status: types.TaskInfoStateError,
			Err:    err}
	} else {
		taskInfo, err := t.WaitForResult(ctx, nil)
		if err != nil {
			select {
			case statusChan <- TaskMessage{
				VmName: name,
				Status: types.TaskInfoStateError,
				Err:    fmt.Errorf("clone vm task failed")}:
				log.Println(taskInfo.Error)
				break
			default:
			}
		} else {
			statusChan <- TaskMessage{
				VmName: name,
				Status: types.TaskInfoStateSuccess,
				Err:    fmt.Errorf("clone vm task failed")}
		}
	}

	<-sem
}

//
// Clone set of VM from a node templates
//
func (p *VmwareVim) CloneVms(projectName string, nodes []*jettypes.NodeTemplate) error {

	if p == nil {
		return fmt.Errorf("vim is nil")
	}

	if len(nodes) == 0 {
		return nil
	}

	if p.datacenter == nil {
		err := p.discoverDatacenter()
		if err != nil {
			return err
		}
	}

	// get root data center root folder
	dataCenterFolder, err := p.datacenter.Folders(p.ctx)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed retriev folder list, err: %s", err)
	}

	// create folder where first half a deployment name and uuid
	folderUuid := uuid.New()
	var newFolderName = projectName + "-" + folderUuid.String()
	newFolder, err := dataCenterFolder.VmFolder.CreateFolder(p.ctx, newFolderName)
	if err != nil {
		return fmt.Errorf("failed create a folder for deployement %v", err)
	}

	sem := make(chan int, 3)

	var finalStatus []TaskMessage
	errChan := make(chan TaskMessage, len(nodes))

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for _, node := range nodes {
		// get the template for a node
		vmTemplate, _, err := p.DiscoverVmTemplates(node)
		node.SetFolderPath(newFolderName)
		if err != nil {
			return err
		}
		go p.runInstantiateTask(p.ctx, newFolder, node.Name, vmTemplate, sem, &wg, errChan)
	}

	wg.Wait()
	close(errChan)
	for e := range errChan {
		//log.Println("Read from thread ", e)
		finalStatus = append(finalStatus, e)
	}

	/** TODO refactor this check each status print green in console */
	for _, v := range finalStatus {
		log.Println("job", v.VmName, "\tstatus", v.Status)
	}

	return nil
}

//
//  Function connects a vm to a target network switch, in case VM has no
//  adapter it will add network adapter.
//
func (p *VmwareVim) ConnectVm(projectName string, node *jettypes.NodeTemplate) (bool, error) {

	ok, err := p.isAttached(node.GetVimName(), node.GenericSwitch().Name())
	if err != nil {
		logging.CriticalMessage("failed to check vm attachment error: " + err.Error())
		return false, err
	}

	if !ok {
		deadline, cancel := context.WithDeadline(p.ctx, time.Now().Add(10*time.Second))
		_, _, err := vcenter.AddNetworkAdapter(deadline,
			p.VimClient(),
			node.GenericSwitch().Uuid(),
			node.VmTemplateName)
		if err != nil {
			if deadline.Err() != context.DeadlineExceeded {
				cancel()
				return false, fmt.Errorf("failed to connect vm, request timeout %s", err)
			}
		}
	} else {
		log.Print("VM already attached to network segment ",
			node.GenericSwitch().Name(), " uuid: ", node.GenericSwitch().Uuid())
	}

	return true, nil
}

//
//   Function synchronize a node  after all VM deployed
//
func (p *VmwareVim) DiscoverVms(projectName string, nodes []*jettypes.NodeTemplate) error {

	for i, node := range nodes {

		// since we cloned a VM old mac belong to a template
		nodes[i].Mac = append(nodes[i].Mac[:0], nodes[i].Mac[1:]...)
		_, _, m, err := vcenter.VmFromCluster(p.ctx, p.VimClient(), node.Name, node.VimCluster)
		if err != nil {
			return fmt.Errorf("vm not found")
		}

		// set new uuid and vm name
		nodes[i].UUID = m.UUID(p.ctx)
		nodes[i].SetVimName(m.Reference().Value)

		devs, err := m.Device(p.ctx)
		if err != nil {
			return fmt.Errorf("failed get device list")
		}

		// append actual mac addresses to a node struct
		for _, dev := range devs {
			if nic, ok := dev.(types.BaseVirtualEthernetCard); ok {
				nodes[i].Mac = append(node.Mac, nic.GetVirtualEthernetCard().MacAddress)
			}
		}

		networks, err := vcenter.GetNetworkAttr(p.ctx, p.VimClient(), m.Reference().Value)
		if err != nil {
			return fmt.Errorf("failed to find deployed VM networks, err: %s", err)
		}

		for _, net := range *networks {
			nodes[i].NetworksRef = append(node.NetworksRef, net.Name)
		}
	}

	return nil
}

//
//
//
func (p *VmwareVim) ChangePowerState(node *jettypes.NodeTemplate, state jettypes.PowerState) (bool, error) {

	logging.Notification("Powering on vm", node.Name)

	_, _, vm, err := vcenter.VmFromCluster(p.ctx, p.VimClient(), node.Name, node.VimCluster)
	if err != nil {
		return false, err
	}

	var task *object.Task
	switch state {
	case jettypes.PowerOn:
		task, err = vm.PowerOn(context.Background())
		if err != nil {
			return false, err
		}
	case jettypes.PowerOff:
		task, err = vm.PowerOn(context.Background())
		if err != nil {
			return false, err
		}
	case jettypes.Reboot:
		err = vm.RebootGuest(context.Background())
		if err != nil {
			return false, err
		}
		return true, nil
	case jettypes.Reset:
		task, err = vm.Reset(context.Background())
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("unkown command")
	}

	deadline, cancel := context.WithDeadline(p.ctx, time.Now().Add(60*time.Second))
	_, err = task.WaitForResult(deadline, nil)
	if err != nil {
		if deadline.Err() != context.DeadlineExceeded {
			cancel()
			return false, fmt.Errorf("failed to acquire ip address, request timeout")
		}
		return false, err
	}

	return true, nil
}

// AcquireIpAddress of VM
func (p *VmwareVim) AcquireIpAddress(node *jettypes.NodeTemplate) (bool, string, error) {

	if len(node.Name) == 0 || len(node.VimCluster) == 0 {
		return false, "", nil
	}
	_, _, vm, err := vcenter.VmFromCluster(p.ctx, p.VimClient(), node.Name, node.VimCluster)
	if err != nil {
		return false, "", fmt.Errorf("failed find a vm %s %v", node.Name, err)
	}

	deadline, cancel := context.WithDeadline(p.ctx, time.Now().Add(60*time.Second))
	ip, err := vm.WaitForIP(deadline)
	if err != nil {
		if deadline.Err() != context.DeadlineExceeded {
			cancel()
			return false, "", fmt.Errorf("failed to acquire ip address, request timeout")
		}
	}

	if ip != "" {
		if node.IPv4AddrStr != ip {
			logging.CriticalMessage(
				" vm booted with different address, expected ", node.IPv4AddrStr, " actual ", ip)
		}
		return true, ip, nil
	}

	return false, "", nil
}
