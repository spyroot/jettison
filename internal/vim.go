package internal

import (
	"database/sql"
	"fmt"
	"github.com/spyroot/jettison/dbutil"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/system"
	"log"
	"os"
	"plugin"
	"strconv"
	"sync"
)

type VimChanMessage struct {
	Msg string
	Ok  bool
	Err error
}

/*
   Main data structure that holds all infra related data.
*/
type Vim struct {
	// a database that jettison use to store active deployment
	db *sql.DB

	// Jettison deployment scenario a jet pack
	jetConfig *AppConfig

	vimPlugin *plugin.Plugin

	// a vim plugin that VIM Manager will use
	pluggableVim jettypes.VimPlugin
}

//
// Returns jettison configuration that initialized during
// init process
//
func (p *Vim) GetJetConfig() *AppConfig {
	if p != nil {
		return p.jetConfig
	}
	return nil
}

// Returns vim client
func (p *Vim) Database() *sql.DB {
	if p != nil {
		return p.db
	}
	return nil
}

/*
   Initialize initial configuration, checks dependency that jettison required.
   (Ansible , ssh client etc)

  - It reads configuration yaml file that contain entire configuration semantics
  - During initialization phase NewVIm loads a plugin that used to interact with VIM

*/
func NewVim() (*Vim, error) {

	// check that we have all require dependency
	err := system.CheckDependence()
	if err != nil {
		log.Fatal(err)
	}

	jetConfig, err := ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration %s", err)
	}

	// populate all required dir
	err = system.BuildTenantDirs(jetConfig.GetAnsible().AnsibleTemplates, jetConfig.GetDeploymentName())
	if err != nil {
		return nil, err
	}

	var vim Vim
	vim.jetConfig = &jetConfig

	p, err := plugin.Open("./plugins/vmwarevim.so")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	vim.vimPlugin = p
	symbol, err := p.Lookup("Init")
	if err != nil {
		panic("Failed lookup init")
	}

	vim.db, err = dbutil.CreateDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database")
	}

	pluggableVim, err := symbol.(func() (jettypes.VimPlugin, error))()
	if err != nil {
		fmt.Println("unexpected type from module symbol")
		os.Exit(1)
	}

	vim.pluggableVim = pluggableVim
	err = pluggableVim.InitPlugin(&jetConfig.Infra.Vcenter)
	if err != nil {
		return nil, fmt.Errorf("failed initilize vim")
	}

	return &vim, nil
}

//
//  Discover a VM template that must be already deployed
//  and set a fact  template vm uuid, network attached, adapter etc.
//
func (p *Vim) DiscoverVmTemplate(node *jettypes.NodeTemplate) error {

	if node == nil {
		return fmt.Errorf("node is nil")
	}

	err := p.pluggableVim.DiscoverVmTemplate(node)
	if err != nil {
		logging.CriticalMessage("vim failed discover a vm template", node.VmTemplateName)
		return err
	}

	return nil
}

//
func (p *Vim) CloneVms(projectName string, nodes []*jettypes.NodeTemplate) error {

	if len(nodes) == 0 {
		return fmt.Errorf("node is nil")
	}

	err := p.pluggableVim.CloneVms(projectName, nodes)
	if err != nil {
		logging.CriticalMessage("vim deploy nodes group")
		return err
	}

	err = p.pluggableVim.DiscoverVms(projectName, nodes)
	if err != nil {
		logging.CriticalMessage("failed discover deployed vms")
		return err
	}

	return nil
}

//
func (p *Vim) DeleteVm(projectName string, nodes []*jettypes.NodeTemplate) error {
	return nil
}

//
// Check and connect a VM to switch, attachment based on a nodes generic switch
// struct
//
func (p *Vim) DisconnectVm(projectName string, node *jettypes.NodeTemplate) (bool, error) {

	ok, err := p.pluggableVim.DisconnectVm(node.VmTemplateName, node)
	if err != nil {
		logging.CriticalMessage("vim failed connect vm")
		return false, err
	}

	if ok {
		node.GenericSwitch().SetAttached(false)
	}

	return ok, err
}

//
// Check and connect a VM to switch, attachment based on a node generic switch
//
func (p *Vim) ConnectVm(projectName string, node *jettypes.NodeTemplate) (bool, error) {

	if len(node.GenericSwitch().Uuid()) == 0 || len(node.GenericSwitch().Name()) == 0 {
		return false, fmt.Errorf("node is not atached to any switch")
	}

	// shared or not
	ok, err := p.pluggableVim.ConnectVm(node.VmTemplateName, node)
	if err != nil {
		logging.CriticalMessage("vim failed connect vm")
		return false, err
	}

	if ok {
		node.GenericSwitch().SetAttached(true)
	}

	return ok, err
}

//
//  Clean up compute resources taken by a project.
//  nodes *[]*jettypes.NodeTemplate point to a slice of nodes that will be used
//  to identify stale items.
//
func (p *Vim) ComputeCleanup(projectName string, nodes []*jettypes.NodeTemplate) error {

	logging.Notification("Deployment",
		projectName, "contains", strconv.Itoa(len(nodes)), "nodes")

	err := p.pluggableVim.ComputeCleanup(projectName, nodes)
	if err != nil {
		logging.CriticalMessage("vim failed delete vm")
		return err
	}

	return nil
}

//
func (p *Vim) CreateDhcpBindings(projectName string, nodes []*jettypes.NodeTemplate) error {

	err := p.pluggableVim.CreateDhcpBindings(projectName, nodes)
	if err != nil {
		logging.CriticalMessage("vim failed delete vm")
		return err
	}

	return nil
}

//
//func (p *Vim) GetCurrentDeployment() {
//
//	err := p.pluggableVim.CreateDhcpBindings(projectName, nodes) {
//
//	}
//
//}

//
//  clean up all state dhcp information based on nodes
//  vim can perform lookup based on mac / switch pair and
//  remove each stale records
//
func (p *Vim) DhcpCleanup(projectName string, nodes []*jettypes.NodeTemplate) error {

	err := p.pluggableVim.DhcpCleanup(projectName, nodes)
	if err != nil {
		logging.CriticalMessage("vim failed delete vm")
		return err
	}

	return nil
}

//
//
//
func (p *Vim) CleanupRouting(projectName string, nodes []*jettypes.NodeTemplate) (bool, error) {

	for _, v := range nodes {
		_, err := p.DeleteRouter(v)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

//
//
//
func (p *Vim) CleanupSwitching(projectName string, nodes []*jettypes.NodeTemplate) (bool, error) {

	for _, v := range nodes {
		_, err := p.DeleteSwitch(v)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

//
//
//
func (p *Vim) CleanupDhcp(projectName string, nodes []*jettypes.NodeTemplate) (bool, error) {

	for _, v := range nodes {
		_, err := p.DeleteDhcpServer(v)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

//
// Discovery a dhcp server connected to a network segment.
//
func (p *Vim) DiscoverClusterDhcpServer(projectName string, nodes *[]*jettypes.NodeTemplate) (bool, error) {

	ok, err := p.pluggableVim.DiscoverClusterDhcpServer(projectName, nodes)
	if err != nil {
		logging.CriticalMessage("vim failed delete vm")
		return false, err
	}

	return ok, nil
}

/* network interface */
func (p *Vim) DeploySegment(projectName string,
	segmentName string, gateway string, prefixLen int) (*jettypes.GenericSwitch, *jettypes.GenericRouter, error) {

	sw, rt, err := p.pluggableVim.DeploySegment(projectName, segmentName, gateway, prefixLen)
	if err != nil {
		logging.CriticalMessage("failed deploy network segments " + err.Error())
		return nil, nil, err
	}

	return sw, rt, nil
}

//
// Create a deployment snapshot in database that capture post deployment state
// that contains all vm primitives
//
func (p *Vim) CreateDeployment(projectName string, nodes []*jettypes.NodeTemplate) (bool, error) {

	err := dbutil.CreateDeployment(p.db, nodes, projectName)
	if err != nil {
		return false, err
	}

	return true, nil
}

//
// Ask vim to change VM power state On based on received state
// depend on implementation that might block
//
func (p *Vim) PowerOn(projectName string, node *jettypes.NodeTemplate) (bool, error) {

	ok, err := p.pluggableVim.ChangePowerState(node, jettypes.PowerOn)
	if err != nil {
		logging.CriticalMessage("failed deploy network segments " + err.Error())
		return false, err
	}

	return ok, nil
}

//
// Ask vim to change VM power state based on received state
// depend on implementation that might block
//
func (p *Vim) ChangePowerState(node *jettypes.NodeTemplate, state jettypes.PowerState) (bool, error) {

	ok, err := p.pluggableVim.ChangePowerState(node, state)
	if err != nil {
		logging.CriticalMessage("failed deploy network segments " + err.Error())
		return false, err
	}

	return ok, nil
}

//
// Ask vim acquire ip address of vm, depend on implementation that might block
//
func (p *Vim) AcquireIpAddress(node *jettypes.NodeTemplate) (bool, error) {

	ok, ip, err := p.pluggableVim.AcquireIpAddress(node)
	if err != nil {
		logging.CriticalMessage("failed deploy network segments " + err.Error())
		return ok, err
	}

	log.Println("vm", node.Name, " ip address ", ip)
	return ok, nil
}

//
// Chane VM power state and write to channel a status
//
func (p *Vim) powerChange(node *jettypes.NodeTemplate, state jettypes.PowerState,
	sem chan int, wg *sync.WaitGroup, statusChan chan VimChanMessage) {

	defer wg.Done()
	sem <- 1

	ok, err := p.ChangePowerState(node, state)
	if err != nil {
		select {
		case statusChan <- VimChanMessage{
			Msg: node.Name,
			Ok:  ok,
			Err: fmt.Errorf("clone vm task failed")}:
			break
		default:
		}
	} else {
		statusChan <- VimChanMessage{
			Msg: node.Name,
			Ok:  ok,
			Err: fmt.Errorf("clone vm task failed")}
	}

	<-sem

	return
}

//
// Changes VMs power state for list of nodes and it does it concurrently
// Underlying semantics semantic need to provide cancellation behavior
// and powerChange method provides correct reporting back to main thread routine
//
func (p *Vim) PowerChangeAll(nodes []*jettypes.NodeTemplate, state jettypes.PowerState) (bool, error) {

	sem := make(chan int, 3)

	var finalStatus []VimChanMessage
	errChan := make(chan VimChanMessage, len(nodes))

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for _, v := range nodes {
		go p.powerChange(v, state, sem, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	for e := range errChan {
		finalStatus = append(finalStatus, e)
	}

	succeed := 0
	for _, v := range finalStatus {
		if v.Ok {
			succeed++
		}
	}

	if succeed == len(nodes) {
		return true, nil
	}

	return false, nil
}

//
// Chane VM power state and write to channel a status
//
func (p *Vim) acquireIp(node *jettypes.NodeTemplate,
	sem chan int, wg *sync.WaitGroup, statusChan chan VimChanMessage) {

	defer wg.Done()
	sem <- 1

	ok, err := p.AcquireIpAddress(node)
	if err != nil {
		select {
		case statusChan <- VimChanMessage{
			Msg: node.Name,
			Ok:  ok,
			Err: fmt.Errorf("clone vm task failed")}:
			break
		default:
		}
	} else {
		statusChan <- VimChanMessage{
			Msg: node.Name,
			Ok:  ok,
			Err: fmt.Errorf("clone vm task failed")}
	}

	<-sem

	return
}

//
// Changes VMs power state for list of nodes and it does it concurrently
// Underlying semantics semantic need to provide cancellation behavior
// and powerChange method provides correct reporting back to main thread routine
//
func (p *Vim) AcquireIpAddresses(nodes []*jettypes.NodeTemplate) (bool, error) {

	sem := make(chan int, 3)

	var finalStatus []VimChanMessage
	errChan := make(chan VimChanMessage, len(nodes))

	var wg sync.WaitGroup
	wg.Add(len(nodes))

	for _, v := range nodes {
		go p.acquireIp(v, sem, &wg, errChan)
	}

	wg.Wait()
	close(errChan)

	for e := range errChan {
		finalStatus = append(finalStatus, e)
	}

	succeed := 0
	for _, v := range finalStatus {
		if v.Ok {
			succeed++
		}
	}

	if succeed == len(nodes) {
		return true, nil
	}

	return false, nil
}

func (p *Vim) DeleteDhcpServer(node *jettypes.NodeTemplate) (bool, error) {

	if node == nil || node.GenericSwitch() == nil {
		return false, nil
	}
	_, err := p.pluggableVim.DeleteDhcpServer(node)
	if err != nil {
		logging.CriticalMessage("failed delete dhcp server " + node.GenericSwitch().DhcpUuid() + " " + err.Error())
		return false, err
	}

	return true, nil
}

func (p *Vim) DeleteRouter(node *jettypes.NodeTemplate) (bool, error) {

	if node == nil || node.GenericRouter() == nil {
		return false, nil
	}

	_, err := p.pluggableVim.DeleteRouter(node)
	if err != nil {
		logging.CriticalMessage("failed delete router " + node.GenericRouter().Uuid() + " " + err.Error())
		return false, err
	}

	return true, nil
}

// Delete a switch
func (p *Vim) DeleteSwitch(node *jettypes.NodeTemplate) (bool, error) {

	if node == nil || node.GenericSwitch() == nil {
		return false, nil
	}

	_, err := p.pluggableVim.DeleteSwitch(node)
	if err != nil {
		logging.CriticalMessage("failed delete switch " + node.GenericSwitch().Uuid() + " " + err.Error())
		return false, err
	}
	return true, nil
}

//
// Adds static route to a given node for example if a tier1 route needs
// have route to pod or vm
func (p *Vim) AddStaticRoute(projectName string, node *jettypes.NodeTemplate, podNetwork string) (bool, error) {

	_, err := p.pluggableVim.AddStaticRoute(projectName, node, podNetwork)
	if err != nil {
		logging.CriticalMessage("failed deploy network segments " + err.Error())
		return false, err
	}
	return true, nil
}
