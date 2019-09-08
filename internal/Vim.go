package internal

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/spyroot/jettison/config"
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/spyroot/jettison/vcenter"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"sync"
)

type Vim struct {
	Db        *sql.DB
	Ctx       context.Context
	VimApi    *govmomi.Client
	NsxApi    nsxt.APIClient
	AppConfig *config.AppConfig
	VimFinder *find.Finder
}

// Returns vim client
func (v Vim) GetVimClient() *vim25.Client {
	return v.VimApi.Client
}

/**

 */
func FindVmObject(ctx context.Context, vmName string, vim *Vim) (*object.VirtualMachine, error) {

	vimPath := vim.AppConfig.GetDc().InventoryPath + "/*/" + vmName
	log.Println("Searching vm ", vmName, " in path:", vimPath)
	v, err := vim.VimFinder.VirtualMachine(ctx, vimPath)
	if err != nil {
		return nil, fmt.Errorf("failed retrieve vm %s details from infrastracture", err)
	}

	log.Println(v.InventoryPath)
	return v, nil
}

/**

 */
func GetVmObject(ctx context.Context, vim *Vim, uuid string) (*object.VirtualMachine, error) {

	vm, err := vcenter.GetSummaryVm(ctx, vim.VimApi.Client, uuid)
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

/*
   Function checks that vm attached to a target logical switch in vCenter
*/
func CheckNetwork(ctx context.Context, vimClient *vim25.Client, c *config.AppConfig, vmName string) error {

	if c == nil {
		return fmt.Errorf("invalid confiugration")
	}

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

func DiscoverClusterDhcpServer(vim *Vim, d *Deployment) (bool, error) {

	nodes, err := d.DeployTaskList()
	if err != nil {
		return false, fmt.Errorf("discoverClusterDhcpServer can't build node list %s", err)
	}

	for k, node := range *nodes {
		// find target logical switch for a k8s cluster -- this setting read global DHCP shared
		// by entire cluster
		logicalSwitch, err := nsxtapi.FindLogicalSwitch(&vim.NsxApi, node.GetSwitchName())
		if err != nil {
			return false, fmt.Errorf("can't find logical switch %s error %s. "+
				"please check configuration", node.GetSwitchName(), err)
		}

		// we set switch id after discovery
		(*nodes)[k].SetSwitchUuid(logicalSwitch.Id)
		log.Println("Discovery...", node.GetSwitchName(), " uuid ", logicalSwitch.Id)
		// find a dhcp server for k8s cluster
		logicalDhcpServer, err := nsxtapi.FindDhcpServerProfile(&vim.NsxApi, logicalSwitch.Id)
		if err != nil {
			return false, fmt.Errorf("can't find a dhcp server attached to logical switch error: %s", err)
		}

		// set dhcp server id
		(*nodes)[k].SetDhcpId(logicalDhcpServer.Id)
		(*nodes)[k].SetDhcName(logicalDhcpServer.DisplayName)

		// TODO set tier-1 router/ check default gateway IP/ tier-1 router must be attached to lds.
	}

	return true, nil
}

/*
  Function initialize initial configuration
  - it reads configuration yaml file
  - open up connection to NSX-T Manager and vCenter

*/
func InitEnvironment() (*Vim, error) {

	appConfig, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration %s", err)
	}

	var vim Vim
	vim.AppConfig = &appConfig

	// open connection to vCenter or NSX-T
	nsxtClient, nsxError := nsxtapi.Connect(appConfig.Infra.Nsxt.Hostname,
		appConfig.Infra.Nsxt.Username,
		appConfig.Infra.Nsxt.Password)
	if nsxError != nil {
		return nil, fmt.Errorf("failed to connect to nsx-t manager")
	}
	vim.NsxApi = nsxtClient

	// open connection to vCenter or ESXi
	vim.Ctx = context.Background()
	vimClient, err := vcenter.Connect(vim.Ctx,
		appConfig.Infra.Vcenter.Hostname,
		appConfig.Infra.Vcenter.Username,
		appConfig.Infra.Vcenter.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to vCenter. %s", err)
	}

	vim.VimApi = vimClient

	// find target logical switch for a k8s cluster -- this setting read global DHCP shared
	// by entire cluster
	logicalSwitch, err := nsxtapi.FindLogicalSwitch(&nsxtClient, appConfig.GetLogicalSwitch())
	if err != nil {
		return nil, fmt.Errorf("can't find logical switch %s error %s. please check configuration",
			appConfig.GetLogicalSwitch(), err)
	}

	// find a DHCP server for k8s cluster
	logicalDhcpServer, err := nsxtapi.FindDhcpServerProfile(&nsxtClient, logicalSwitch.Id)
	if err != nil {
		return nil, fmt.Errorf("can't find a dhcp server attached to logical switch error: %s", err)
	}
	appConfig.Infra.Nsxt.DhcpServerID = logicalDhcpServer.Id

	// Find target data center where we will deploy everything
	vim.VimFinder = find.NewFinder(vimClient.Client, true)
	appConfig.Infra.Vcenter.Dc, err = vim.VimFinder.Datacenter(vim.Ctx, appConfig.Infra.Vcenter.Datacenter)
	if err != nil {
		return nil, fmt.Errorf("failed to get data center details check config and vim")
	}

	return &vim, nil
}

/*
   Function searches a VM template and returns *object.VirtualMachine, *[]mo.Network
*/
func FindTemplate(ctx context.Context, vim *Vim, node *config.NodeTemplate) (*object.VirtualMachine, *[]mo.Network, error) {

	vmSummary, err := vcenter.GetVmAttr(ctx, vim.GetVimClient(), vcenter.VmLookupMap["name"], node.VmTemplateName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve vm template. err: %s", err)
	}

	if node.UUID == "" {
		node.UUID = vmSummary.Summary.Config.Uuid
	}

	log.Println("Found vm template", vmSummary.Summary.Config.Name, vmSummary.Summary.Config.Uuid)

	vm, err := FindVmObject(ctx, node.VmTemplateName, vim)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find vm template object. err: %s", err)
	}

	node.SetVimName(vm.Reference().Value)

	devs, err := vm.Device(ctx)
	if err != nil {
		return nil, nil, err
	}

	for _, dev := range devs {
		if nic, ok := dev.(types.BaseVirtualEthernetCard); ok {
			log.Println(nic.GetVirtualEthernetCard().MacAddress)
			log.Println(nic.GetVirtualEthernetCard().ExternalId)
			node.Mac = append(node.Mac, nic.GetVirtualEthernetCard().MacAddress)
		}
	}

	networks, err := vcenter.GetNetworkAttr(ctx, vim.GetVimClient(), vm.Reference().Value)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find vm template networks, err: %s", err)
	}

	for _, net := range *networks {
		node.NetworksRef = append(node.NetworksRef, net.Name)
	}

	return vm, networks, nil
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
  Function get a template vm from vCenter. Note template must have ethernet adapter attached
  to correct logical switch.
*/
func getTemplateData(ctx context.Context, vim *Vim, node *config.NodeTemplate) (*object.VirtualMachine, *[]mo.Network, error) {

	if node == nil {
		return nil, nil, fmt.Errorf("node is nil")
	}

	vm, networks, err := FindTemplate(ctx, vim, node)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retieve template %s", err)
	}

	if len(*networks) == 0 {
		return nil, nil, fmt.Errorf("vm has no network attached %s. "+
			"Please check that template has network adapter and attached to correct switch", node.GetSwitchName())
	}

	// check template attached to correct logical switch
	if checkNetworkAttachment(networks, node.GetSwitchName()) == false {
		return nil, nil, fmt.Errorf("vm has no network adapter attached to %s", node.GetSwitchName())
	}

	return vm, networks, nil
}

/**

 */
func runInstantiateTask(ctx context.Context, vim *Vim,
	f *object.Folder, name string, template *object.VirtualMachine,
	sem chan int, wg *sync.WaitGroup, statuChan chan TaskMessage) {

	defer wg.Done()
	sem <- 1

	vmConfigSpec := types.VirtualMachineCloneSpec{}
	t, err := template.Clone(vim.Ctx, f, name, vmConfigSpec)
	if err != nil {
		statuChan <- TaskMessage{
			VmName: name,
			Status: types.TaskInfoStateError,
			Err:    err}
	} else {
		taskInfo, err := t.WaitForResult(ctx, nil)
		if err != nil {
			select {
			case statuChan <- TaskMessage{
				VmName: name,
				Status: types.TaskInfoStateError,
				Err:    fmt.Errorf("clone vm task failed")}:
				log.Println(taskInfo.Error)

				break
			default:
			}
		} else {
			statuChan <- TaskMessage{
				VmName: name,
				Status: types.TaskInfoStateSuccess,
				Err:    fmt.Errorf("clone vm task failed")}
		}
	}

	<-sem
}

/**

 */
func CreateVms(ctx context.Context, vim *Vim, dep *Deployment) error {

	if vim == nil {
		return fmt.Errorf("vim is nil")
	}

	if dep == nil {
		return fmt.Errorf("deployment is nil")
	}

	nodes, err := dep.DeployTaskList()
	if err != nil {
		return err
	}

	// get data center root folder
	dataCenterFolder, err := vim.AppConfig.GetDc().Folders(ctx)
	if err != nil {
		return fmt.Errorf("failed retriev folder, err: %s", err)
	}

	// create folder where first half a deployment name and uuid
	folderUuid := uuid.New()
	var newFolderName = dep.DeploymentName + "-" + folderUuid.String()
	newFolder, err := dataCenterFolder.VmFolder.CreateFolder(ctx, newFolderName)
	if err != nil {
		return err
	}

	sem := make(chan int, 3)

	var finalStatus []TaskMessage
	errChan := make(chan TaskMessage, len(*nodes))

	var wg sync.WaitGroup
	wg.Add(len(*nodes))

	for _, node := range *nodes {
		// get the template for a node
		vmTemplate, _, err := getTemplateData(ctx, vim, node)
		node.SetFolderPath(newFolderName)
		if err != nil {
			return err
		}
		go runInstantiateTask(ctx, vim, newFolder, node.Name, vmTemplate, sem, &wg, errChan)
	}

	wg.Wait()
	close(errChan)
	for e := range errChan {
		//log.Println("Read from thread ", e)
		finalStatus = append(finalStatus, e)
	}

	for _, v := range finalStatus {
		log.Println("job", v.VmName, "\tstatus", v.Status)
	}

	return nil
}

func DestroyVm() {
}

/**
  Function is debug routine
*/
func DebugDeployment(dep *Deployment) {

	allNodes, err := dep.DeployTaskList()
	if err == nil {
		for _, v := range *allNodes {
			v.PrintAsJson()
		}
	}
}

func DebugNodes(allNodes *[]*config.NodeTemplate) {
	for _, v := range *allNodes {
		v.PrintAsJson()
	}
}
