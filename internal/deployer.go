package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/spyroot/jettison/ansibleutil"
	"github.com/spyroot/jettison/certsutil"
	"github.com/spyroot/jettison/consts"
	"github.com/spyroot/jettison/dbutil"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/netpool"
	"github.com/spyroot/jettison/osutil"
	"github.com/spyroot/jettison/sshclient"
)

type Deployer struct {
	vim             *Vim
	scenario        *Deployment2
	networkSegments *jettypes.NetworkSegments
	taskStack       []jettypes.DeployerCmd
}

func NewDeployer(scenario *Deployment2, vim *Vim) *Deployer {
	return &Deployer{scenario: scenario, vim: vim}
}

/**
  Build a network segments based on template map
*/
func (d *Deployer) buildSegments() error {

	templates := d.scenario.DeploymentTemplates()
	s, err := jettypes.NewNetworkSegments(templates)
	if err != nil {
		return err
	}

	d.networkSegments = s
	return nil
}

/*
  For each network segment, deploy segment.
  Set of node templates in scenario can share same network or each template
  can describe distinct set of node that need to be deployed in seperate
  network segment.

  A segment defined a single broadcast domain , i.e single logical switch that paired
  with logical router and dhcp server dedicate for that segment.
*/
func (d *Deployer) deployNetworks() error {

	s := d.networkSegments.Segments()

	for _, seg := range s {
		gateway, sharedNet, err := sharedAttributes(seg.Segments())
		if err != nil {
			logging.ErrorLogging(err)
			return err
		}
		prefixLen, _ := sharedNet.Mask.Size()
		if prefixLen < 8 || prefixLen >= 32 {
			e := fmt.Errorf(" subnet mask need to between larger than 7 bit and less than 32 bits")
			logging.ErrorLogging(e)
			return e
		}

		segmentSwitch, segmentRouter, err := d.vim.DeploySegment(d.scenario.DeploymentName, seg.SegmentName(), gateway, prefixLen)
		if err != nil {
			logging.ErrorLogging(err)
			return err
		}
		// update template and set values
		for _, v := range seg.Segments() {
			v.SetGenericSwitch(segmentSwitch)
			v.SetGenericRouter(segmentRouter)
		}
	}

	return nil
}

//  Attach each VM templates to a network, the reason is we don't want
//  issue N REST API call to connect each respected cloned VM after we cloned,
//  so it much better to connect template vm to target switch
//  note if template already contains an nic adapter , method will ask vim
//  disconnect old adapter and will create  new one
func (d *Deployer) attachTemplates() error {

	processed := make(map[string]*jettypes.NodeTemplate)
	templates := d.scenario.DeploymentTemplates()
	for i, t := range templates {
		if val, ok := processed[templates[i].VmTemplateName]; ok {
			templates[i].UUID = val.UUID
			templates[i].VimName = val.VimName
			templates[i].Mac = val.Mac
			templates[i].NetworksRef = val.NetworksRef
			continue
		}

		err := d.vim.DiscoverVmTemplate(t)
		if err != nil {
			logging.CriticalMessage("Failed discover vm template")
			return err
		}

		if t.NetworksRef[0] != t.GenericSwitch().Name() {
			succeed, err := d.vim.DisconnectVm(d.scenario.DeploymentName, t)
			if err != nil {
				return err
			}
			if succeed {
				logging.Notification(" Template disconnected")
				t.NetworksRef = nil
				t.Mac = nil
			}
		}

		if len(t.Mac) == 0 {
			_, err := d.vim.ConnectVm(d.scenario.DeploymentName, t)
			if err != nil {
				logging.CriticalMessage("Failed attach vm template to a logical switch")
				return err
			}
		}
		processed[templates[i].VmTemplateName] = t
	}

	return nil
}

//  Return flat slice for all nodes from the map
//
func (d *Deployer) nodeSlice() []*jettypes.NodeTemplate {

	nodes := make([]*jettypes.NodeTemplate, 0)

	for _, n := range d.scenario.nodesGroup {
		for _, v := range n {
			nodes = append(nodes, v)
		}
	}

	return nodes
}

const (
	WriteCheckCmd = "/bin/touch sshclientping"
	ReadCheckCmd  = "/bin/ls sshclientping"
	DeleteCmd     = "rm ~/sshclientping"
	RespondOutput = "sshclientping"
)

/**

 */
func (d *Deployer) deployMgmtChannel(nodes []*jettypes.NodeTemplate) (bool, error) {

	// copy ssh key to each host
	sshDefaults := d.vim.jetConfig.GetSshDefault()
	for _, n := range nodes {
		err := sshclient.SshCopyId(sshDefaults, n.IPv4AddrStr)
		if err != nil {
			return false, fmt.Errorf("failed copy ssh key, error: %v", err)
		}
	}

	// ssh to each host and check write to disk access
	// re-ssh back and validate that file still there.
	for _, h := range nodes {
		logging.Notification("Sending ssh ping to a host \t", h.Name, "\t", h.IPv4AddrStr)

		out, err := sshclient.RunRemoteCommand(sshDefaults, h.IPv4AddrStr, WriteCheckCmd)
		if err != nil {
			return false, fmt.Errorf("failed execute command on remote host %s %v", h.IPv4AddrStr, err)
		}
		out, err = sshclient.RunRemoteCommand(sshDefaults, h.IPv4AddrStr, ReadCheckCmd)
		if err != nil {
			return false, fmt.Errorf("failed execute command on remote host %s %v", h.IPv4AddrStr, err)
		}
		if !strings.Contains(out, RespondOutput) {
			return false, fmt.Errorf("ssh key injection failed for host %s %v", h.Name, h.IPv4AddrStr)
		}
		_, err = sshclient.RunRemoteCommand(sshDefaults, h.IPv4AddrStr, DeleteCmd)
		if err != nil {
			return false, fmt.Errorf("failed execute command on remote host %s %v", h.IPv4AddrStr, err)
		}
		logging.Notification("Got ssh pong back from host \t", h.Name, "\t", h.IPv4AddrStr)
	}

	return true, nil
}

//
//  Method set a discovered fact about the environment
//  to a group of nodes that share same template
func (d *Deployer) setFacts() error {

	for k, v1 := range d.scenario.nodesGroup {
		templateData := d.scenario.nodeTemplates[k]

		switchUuid := templateData.GenericSwitch().Uuid()
		switchName := templateData.GenericSwitch().Name()
		dhcpUuid := templateData.GenericSwitch().DhcpUuid()
		routerUuid := templateData.GenericRouter().Uuid()
		routerName := templateData.GenericRouter().Name()

		if len(switchUuid) == 0 {
			return fmt.Errorf("switch need to be discovered. ")
		}

		if len(switchName) == 0 {
			return fmt.Errorf("switch needs to be discovered. ")
		}

		if len(dhcpUuid) == 0 {
			return fmt.Errorf("dhcp server needs to be discovered. ")
		}

		if len(routerUuid) == 0 {
			return fmt.Errorf("router needs to be discovered. ")
		}

		if len(routerName) == 0 {
			return fmt.Errorf("router  name needs to be discovered. ")
		}

		for _, node := range v1 {

			// switch and dhcp facts
			node.NetworksRef = templateData.NetworksRef

			s := &jettypes.GenericSwitch{}
			node.SetGenericSwitch(s)
			node.GenericSwitch().SetUuid(switchUuid)
			node.GenericSwitch().SetName(switchName)
			node.GenericSwitch().SetDhcpUuid(dhcpUuid)

			// routing facts
			r := &jettypes.GenericRouter{}
			node.SetGenericRouter(r)
			node.GenericRouter().SetUuid(routerUuid)
			node.GenericRouter().SetName(routerName)

		}
	}
	return nil
}

//
//
func (d *Deployer) deployDhcpBindings(nodes []*jettypes.NodeTemplate) (bool, error) {

	// we need all mac addresses, so we need do it second pass
	for _, v := range d.scenario.nodesGroup {
		err := d.vim.CreateDhcpBindings(d.scenario.DeploymentName, v)
		if err != nil {
			err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
			if err != nil {
				return false, fmt.Errorf("failed cleanup environment error: %v", err)
			}
			return false, fmt.Errorf("failed cleanup environment error: %v", err)
		}
	}
	return true, nil
}

//
func (d *Deployer) ansibleDiscovery(nodes []*jettypes.NodeTemplate) (bool, error) {

	var nodeRoles = []string{"Controllers", "Kubelets", "Ingress"}

	var success = 0
	for _, role := range nodeRoles {
		// each section looks project name-group name
		roleInProject := d.scenario.DeploymentName + role
		jsonRespond, err := ansibleutil.Ping(ansibleutil.AnsibleCommand{
			Path:   "/usr/local/bin/ansible",
			CMD:    []string{roleInProject, "-m", "ping"},
			Config: "",
		})

		ansibleResp := ansibleutil.AnsiblePing{}
		err = json.Unmarshal([]byte(jsonRespond), &ansibleResp)
		if err != nil {
			logging.CriticalMessage("failed parse ansible output. make sure ansible respond as json")
			if strings.Contains(jsonRespond, "SUCCESS") {
				success++
			}
		}
		log.Println(ansibleResp.Stats)
	}

	if success == len(nodes) {
		log.Println("Got respond from all ansible slaves")
	}

	return true, nil
}

//
//
//
func (d *Deployer) ansibleAddWorkers(nodes []*jettypes.NodeTemplate) (bool, error) {

	jetConfig := d.vim.jetConfig

	for _, node := range nodes {
		if node.Type == jettypes.WorkerType {

			// each host has own ansible variables so we store this
			hostsVar := ansibleutil.AnsibleHostVars{
				Podnet:   fmt.Sprintf("%s/%d", node.GetCidr(), jetConfig.GetCluster().AllocateSize),
				Hostname: node.Name,
				HomePath: jetConfig.GetAnsible().AnsibleConfig,
			}

			// write hosts vars to a file
			err := hostsVar.WriteToFile()
			if err != nil {
				logging.ErrorLogging(err)
				return false, fmt.Errorf("failed write ansible hosts file")
			}
		}
	}

	logging.Notification("All workers hosts file generated")

	return true, nil
}

//
// Allocate a network per each pod
//
func (d *Deployer) AllocatePodNetwork(nodes []*jettypes.NodeTemplate) (bool, error) {

	jetConfig := d.vim.jetConfig
	clusterCidr := jetConfig.GetCluster().ClusterCidr
	projectName := d.scenario.DeploymentName

	// create IP pool based on client cluster cidr
	pool, err := netpool.NewSubnetPool(clusterCidr, uint(jetConfig.GetCluster().AllocateSize))
	if err != nil {
		return false, fmt.Errorf("failed create subnet pool manager %v", err)
	}

	for i, node := range nodes {
		if node.Type == jettypes.WorkerType {
			addrBlock, err := pool.AllocateSubnet()
			if err != nil {
				logging.ErrorLogging(err)
				return false, fmt.Errorf("failed allocate ip block for a pod error: %v", err)
			}

			nodes[i].SetPodCidr(addrBlock.String())
			nodes[i].PodAllocationSize(jetConfig.GetCluster().AllocateSize)

			podNetwork := fmt.Sprintf("%s/%d", addrBlock.String(), jetConfig.GetCluster().AllocateSize)
			_, err = dbutil.MakeAllocation(d.vim.db, node, projectName, podNetwork, clusterCidr)
			if err != nil {
				logging.ErrorLogging(err)
				return false, fmt.Errorf("failed allocate cidr block to a pod")
			}

			_, err = d.vim.AddStaticRoute(d.scenario.DeploymentName, node, podNetwork)
			if err != nil {
				logging.CriticalMessage("failed to add static route")
			}
		}
	}

	logging.Notification("All workers hosts file generated")
	return true, nil
}

//
//   Generate ansible inventory
//
func (d *Deployer) createAnsibleInventory(nodes []*jettypes.NodeTemplate) (bool, error) {

	ansibleConfig := d.vim.jetConfig.GetAnsible()
	sshConfig := d.vim.jetConfig.GetSshDefault()

	ansibleInventory, err := ansibleutil.CreateFromInventory(ansibleConfig.AnsibleInventory)
	if err != nil {
		// clean up need happen here.
		return false, fmt.Errorf("failed create new ansible scenario check access to %s",
			ansibleConfig.AnsibleInventory)
	}

	// add all node to ansible inventory
	for _, n := range nodes {
		// get ansible group name for a given project
		var groupName = ansibleutil.GetAnsibleGroupName(n.Type, d.scenario.DeploymentName)
		if osutil.CheckIfExist(ansibleConfig.AnsibleInventory) == false {
			log.Println("Generating a new ansible inventory")
		}
		// add ansible host to ansible inventory
		ansibleHost := ansibleutil.AnsibleHosts{
			Name:     n.Name,
			Hostname: n.IPv4AddrStr,
			Port:     sshConfig.SshPort,
			User:     sshConfig.SshUsername,
			Group:    groupName,
		}
		err := ansibleInventory.AddSlaveHost(&ansibleHost, d.scenario.DeploymentName)
		if err != nil {
			return false, fmt.Errorf("failed create ansible inventory %v", err)
		}
	}

	// write to a file
	err = ansibleInventory.WriteToFile()
	if err != nil {
		return false, fmt.Errorf("failed create ansible inventory file %v", err)
	}

	logging.Notification("Ansible successfully generated")

	return true, nil
}

//
//
//
func (d *Deployer) createAnsibleGlobals(nodes []*jettypes.NodeTemplate) (bool, error) {

	ansibleEnv := d.vim.jetConfig.GetAnsible()
	jetConfig := d.vim.jetConfig

	projectName := d.scenario.DeploymentName
	serviceCidr := jetConfig.GetCluster().ServiceCidr

	// generate ansible global variables for a project
	ansibleGlobal := ansibleutil.NewAnsibleGlobalVars()
	ansibleGlobal.HomePath = ansibleEnv.AnsibleConfig
	ansibleGlobal.Project = projectName
	ansibleGlobal.TenantHome = path.Join(ansibleEnv.AnsibleConfig,
		consts.DefaultAnsibleTenantDir, ansibleGlobal.Project)

	// default owner is same username as ssh.
	ansibleGlobal.Owner = jetConfig.Infra.SshDefaults.Username()
	ansibleGlobal.ClusterDns = jetConfig.GetCluster().ClusterDns

	// TODO write test for pool so it doesn't do anything funny. I didn't do edge cases check
	// TODO add check for parse so it take only value between 0 to 32
	ansibleGlobal.ClusterCidr = jetConfig.GetCluster().ClusterCidr
	ansibleGlobal.ServiceCidr = jetConfig.GetCluster().ServiceCidr

	ansibleGlobal.EncyrptionKey = "w7zi7kwrXgD0XfHs3VRyOoaTvUlzC7VoGCW/vU1ULKk="

	certClients := make([]certsutil.CertClient, 0)
	for _, n := range nodes {
		certClients = append(certClients, n)
	}

	// generate certs
	keys, err := certsutil.GenerateTenantCerts(certClients,
		ansibleEnv.AnsibleTemplates, projectName, serviceCidr)
	if err != nil {
		logging.ErrorLogging(err)
		return false, err
	}

	// global variables shared by entire project
	for _, node := range nodes {
		if node.Type == jettypes.ControlType {
			// add all controller
			ansibleGlobal.MasterNode = append(ansibleGlobal.MasterNode, node.IPv4AddrStr)
		}
		if node.Type == jettypes.IngressType {
			ansibleGlobal.IngressIP = node.IPv4AddrStr
			ansibleGlobal.IngressHostname = node.Name
		}
	}

	// write all global ansible
	err = ansibleGlobal.WriteToFile()
	if err != nil {
		return false, fmt.Errorf("failed to write ansible config files")
	}

	err = ansibleGlobal.WriteVars(keys)
	if err != nil {
		return false, fmt.Errorf("failed to write ansible vars to a file")
	}

	return true, nil
}

//
//
//
func (d *Deployer) createAnsiblePlaybook() (bool, error) {

	baseDir := d.vim.jetConfig.GetAnsible().AnsibleConfig
	filePath := path.Join(baseDir, consts.PlaybookTemplate)

	playbook, err := ansibleutil.MakeNewFromFile(filePath)
	if err != nil {
		return false, err
	}

	for key, _ := range d.scenario.nodesGroup {
		groupName := d.scenario.DeploymentName + key
		playbook.Transform(key, groupName)
	}

	fiName := d.scenario.DeploymentName + ".yml"
	fiPath := path.Join(baseDir, fiName)
	logging.Notification("Generating playbook ", filePath)

	fi, err := os.OpenFile(fiPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false, err
	}
	defer fi.Close()

	err = playbook.Write(fi)
	if err != nil {
		return false, err
	}

	return true, nil
}

//
// Main routine called by jettison to deploy a given scenario.
//
func (d *Deployer) Deploy() error {

	err := d.buildSegments()
	if err != nil {
		//		d.taskStack = append(d.taskStack, "networksegments")
		logging.ErrorLogging(err)
		return fmt.Errorf("failed build network segment list")
	}

	ok, err := d.createAnsiblePlaybook()
	if ok == false {
		return nil
	}

	err = d.deployNetworks()
	if err != nil {
		return nil
	}

	var isNew = false
	_, _, ok, err = dbutil.GetDeployment(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		log.Fatal("error ", err)
	}

	if ok {
		isNew, err = d.promptForRedeploy()
		if err != nil {
			return fmt.Errorf("error %v", err)
		}
		if isNew == false {
			return nil
		}

		ok, err := d.promptDeploy()
		if err != nil {
			return fmt.Errorf("error %v", err)
		}
		if !ok {
			return nil
		}
	}

	// attach templates
	err = d.attachTemplates()
	if err != nil {
		return err
	}

	// set and check facts
	err = d.setFacts()
	if err != nil {
		return err
	}

	// for each group of node deploy
	//	d.taskStack = append(d.taskStack, "clonevm")
	for _, v := range d.scenario.nodesGroup {
		err = d.vim.CloneVms(d.scenario.DeploymentName, v)
		if err != nil {
			return fmt.Errorf("failed clone a vms err: %v", err)
		}
	}

	nodes := d.nodeSlice()

	for _, g := range d.scenario.nodesGroup {
		for _, v := range g {
			v.PrintAsJson()
		}
	}

	if ok, err = d.deployDhcpBindings(nodes); !ok {
		//		d.taskStack = append(d.taskStack, "dhcp")
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	if ok, err = d.vim.CreateDeployment(d.scenario.DeploymentName, nodes); !ok {
		//		d.taskStack = append(d.taskStack, "database")
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	if ok, err = d.vim.PowerChangeAll(nodes, jettypes.PowerOn); !ok {
		//		d.taskStack = append(d.taskStack, "poweredon")
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	log.Println("Acquiring ip addresses please wait...")
	if ok, err = d.vim.AcquireIpAddresses(nodes); !ok {
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	log.Println("All addresses acquired")

	//	d.taskStack = append(d.taskStack, "sshkeys")
	if ok, err = d.deployMgmtChannel(nodes); !ok {
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	//	d.taskStack = append(d.taskStack, "ansibleinventory")
	if ok, err = d.createAnsibleInventory(nodes); !ok {
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	if ok, err = d.ansibleDiscovery(nodes); !ok {
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	if ok, err = d.createAnsibleGlobals(nodes); !ok {
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	if ok, err = d.ansibleAddWorkers(nodes); !ok {
		err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			return err
		}
	}

	if ok, err := d.Execute(jettypes.AnsibleDeploy); !ok {
		if err != nil {
			return err
		}
		logging.Notification("looks like we have failed task.")
	}

	return nil
}

/*
   Check that all template in slice use same IPv4 gateway, in case of
   mismatch return error otherwise a gateway that all template
   need use.
*/
func sharedAttributes(templates []*jettypes.NodeTemplate) (string, *net.IPNet, error) {

	if len(templates) == 0 {
		return "", nil, fmt.Errorf("empty tempalte list")
	}

	gateway := templates[0].Gateway
	network := templates[0].IPv4Net

	for _, v := range templates {
		if v.Gateway != gateway {
			return "", nil, fmt.Errorf("all template in group must use same gateway for a same segment")
		}
		if v.IPv4Net.String() != network.String() {
			return "", nil, fmt.Errorf("all template in group must use same network for a network segment")
		}
	}
	return gateway, network, nil
}

/**

 */
func (d *Deployer) promptDeploy() (bool, error) {

	log.Println("Deployment already in the system and num nodes deployed")
	fmt.Print("Do you want deploy new project ?: (yes/no/show)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// if exist delete and re-create
		if strings.Contains(scanner.Text(), "yes") || scanner.Text()[0] == 'y' {
			return true, nil
		} else if strings.Contains(scanner.Text(), "show") || scanner.Text()[0] == 's' {
			nodes, ok, err := dbutil.GetDeploymentNodes(d.vim.Database(), d.scenario.DeploymentName)
			if err != nil {
				log.Fatal(err)
			}
			if ok {
				DebugNodes(&nodes)
			}
			return false, nil
		} else {
			return false, nil
		}
	}

	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}

	return false, nil
}

/**

 */
func (d *Deployer) promptForRedeploy() (bool, error) {

	log.Println("Deployment already in the system and num nodes deployed")
	fmt.Print("Do you want kill existing deployment ?: (yes/no/show)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// if exist delete and re-create
		if strings.Contains(scanner.Text(), "yes") || scanner.Text()[0] == 'y' {
			ok, err := d.cleanupFromDb()
			if err != nil {
				return false, err
			}
			return ok, nil
		} else if strings.Contains(scanner.Text(), "show") || scanner.Text()[0] == 's' {
			nodes, ok, err := dbutil.GetDeploymentNodes(d.vim.Database(), d.scenario.DeploymentName)
			if err != nil {
				log.Fatal(err)
			}
			if ok {
				DebugNodes(&nodes)
			}
			return false, nil
		} else {
			return false, nil
		}
	}

	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}

	return false, nil
}

/**

 */
func (d *Deployer) cleanupFromDb() (bool, error) {

	// remove dhcp binding
	err := d.networkCleanupFromDb()
	if err != nil {
		logging.CriticalMessage("Failed delete dhcp binding")
		return false, err
	}

	// remove all vm
	err = d.computeCleanupFromDb()
	if err != nil {
		logging.CriticalMessage("Failed delete vms")
		return false, err
	}

	// remove from ansible all old inventory
	err = d.ansibleCleanupFromDb()
	if err != nil {
		logging.CriticalMessage("Failed delete host from ansible inventory")
		return false, err
	}

	// remove from database old deployment.
	err = dbutil.DeleteDeployment(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		return false, err
	}

	return true, nil
}

/*

 */
func (d *Deployer) Cleanup(nodes []*jettypes.NodeTemplate) (bool, error) {

	err := d.vim.DhcpCleanup(d.scenario.DeploymentName, nodes)
	if err != nil {
		logging.CriticalMessage("Failed delete dhcp binding from dhcp server")
		return false, err
	}
	err = d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
	if err != nil {
		logging.CriticalMessage("Failed delete vms from vim")
		return false, err
	}
	err = d.ansibleCleanup(nodes)
	if err != nil {
		logging.CriticalMessage("Failed delete hosts from ansible inventory")
		return false, err
	}
	err = dbutil.DeleteDeployment(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		return false, err
	}

	return true, nil
}

/**

 */
func (d *Deployer) promptDeleteNetworking() bool {

	fmt.Print("Do you want delete network object  ?: (yes/no/show)")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "yes") || scanner.Text()[0] == 'y' {
			return true
		} else if strings.Contains(scanner.Text(), "no") || scanner.Text()[0] == 'n' {
			return false
		} else {
			return false
		}
	}

	if scanner.Err() != nil {
		return false
	}

	return false
}

//
//  Clean up all object from NSX based on snapshot take and stored in database.
//  before deleting nsx-t object it will prompt for confirmation
//  TODO that can be moved to yaml itself. or flag via cmd
func (d *Deployer) networkCleanupFromDb() error {

	nodes, ok, err := dbutil.GetDeploymentNodes(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}

	if ok {
		err := d.vim.DhcpCleanup(d.scenario.DeploymentName, nodes)
		if err != nil {
			logging.ErrorLogging(err)
			return err
		}

		// ask if no don't delete anything and we can re-use later
		if !d.promptDeleteNetworking() {
			return nil
		}

		_, err = d.vim.CleanupDhcp(d.scenario.DeploymentName, nodes)
		if err != nil {
			logging.ErrorLogging(err)
		}

		_, err = d.vim.CleanupRouting(d.scenario.DeploymentName, nodes)
		if err != nil {
			logging.ErrorLogging(err)
		}

		_, err = d.vim.CleanupSwitching(d.scenario.DeploymentName, nodes)
		if err != nil {
			logging.ErrorLogging(err)
		}

	}

	return nil
}

/*
   Clean up deployment from a snapshot taken post deployment that must be stored
   from database.
*/
func (d *Deployer) computeCleanupFromDb() error {

	nodes, ok, err := dbutil.GetDeploymentNodes(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}
	if ok {
		return d.vim.ComputeCleanup(d.scenario.DeploymentName, nodes)
	}
	return nil
}

/*
  Load existing deployment from database and clean up ansible inventory.
  based on snapshot taken post deployment.
  It used mainly when we need re-deploy entire cluster for a same project
  and we need clean up ansible inventory for stale entire.
*/
func (d *Deployer) ansibleCleanupFromDb() error {

	nodes, ok, err := dbutil.GetDeploymentNodes(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}
	if ok {
		return d.ansibleCleanup(nodes)
	}
	return nil
}

/*
   Function clean up ansible inventory for a given project.
*/
func (d *Deployer) ansibleCleanup(nodes []*jettypes.NodeTemplate) error {

	jetConfig := d.vim.GetJetConfig()
	ansibleConfig := jetConfig.GetAnsible()
	inventory, err := ansibleutil.CreateFromInventory(ansibleConfig.AnsibleInventory)
	if err != nil {
		logging.CriticalMessage("looks like inventory file already empty err: ", err.Error())
		return nil
	}

	for _, node := range nodes {
		logging.Notification("Deleting from ansible inventory a vm: ", node.Name)
		ok := inventory.DeleteSaveHost(node.Name)
		if !ok {
			logging.CriticalMessage("failed delete ansible host", node.Name, "from the inventory.")
		}

		file := path.Join(jetConfig.GetAnsible().AnsibleConfig, "/", consts.DefaultHostsVarsPath, "/", node.Name)
		if osutil.CheckIfExist(file) {
			err := os.Rename(file, file+".old")
			if err != nil {
				logging.CriticalMessage("failed rename old file for host ", node.Name)
			}
		}
	}

	err = inventory.WriteToFile()
	if err != nil {
		log.Println("looks like we can't to write to ansible file: ", err)
		return nil
	}

	return nil
}

func (d *Deployer) LoadScenario() ([]*jettypes.NodeTemplate, error) {

	_, _, ok, err := dbutil.GetDeployment(d.vim.Database(), d.scenario.DeploymentName)
	if err != nil {
		log.Fatal("error ", err)
	}

	if !ok {
		err := fmt.Errorf("failed fetch deployment from database")
		logging.ErrorLogging(err)
		return nil, err
	}

	nodes, ok, err := dbutil.GetDeploymentNodes(d.vim.Database(), d.vim.jetConfig.GetDeploymentName())
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	if !ok {
		err := fmt.Errorf("failed fetch deployment from database")
		logging.ErrorLogging(err)
		return nil, err
	}

	return nodes, err
}

func (d *Deployer) LoadCmd() ([]*jettypes.NodeTemplate, error) {

	nodes, err := d.LoadScenario()
	if err != nil {
		return nil, err
	}

	for i, v := range nodes {
		ok, template := d.scenario.Template(v.GetNodeType())
		if ok {
			nodes[i].IPv4Net = template.IPv4Net
			nodes[i].Static = template.Static
		}
	}

	return nodes, nil
}

func (d *Deployer) AnsiblePlaybookCmd() error {

	d.taskStack = append(d.taskStack, jettypes.AnsiblePlaybook)
	_, err := d.createAnsiblePlaybook()
	if err != nil {
		return nil
	}

	nodes, err := d.LoadCmd()

	for _, v := range nodes {
		d.scenario.Template(v.GetNodeType())
		v.PrintAsJson()
	}

	d.taskStack = d.taskStack[:len(d.taskStack)-1]

	return nil
}

func (d *Deployer) AnsiblePrintCmd() (bool, error) {

	d.taskStack = append(d.taskStack, jettypes.AnsiblePlaybook)
	_, err := d.createAnsiblePlaybook()
	if err != nil {
		return false, nil
	}

	nodes, err := d.LoadCmd()

	for _, v := range nodes {
		d.scenario.Template(v.GetNodeType())
		v.PrintAsJson()
	}

	d.taskStack = d.taskStack[:len(d.taskStack)-1]

	return true, nil
}

func (d *Deployer) AnsibleInventoryCmd() (bool, error) {

	d.taskStack = append(d.taskStack, jettypes.AnsibleInventory)

	nodes, err := d.LoadCmd()
	if err != nil {
		return false, err
	}

	ok, err := d.createAnsibleInventory(nodes)
	if err != nil {
		return false, err
	}

	d.taskStack = d.taskStack[:len(d.taskStack)-1]

	return ok, nil
}

// run ansible playbook.
func (d *Deployer) AnsibleDeployCmd() (bool, error) {

	cmd := ansibleutil.AnsibleCommand{Path: "/usr/local/bin/ansible-playbook",
		CMD:    []string{"-b", d.scenario.DeploymentName + ".yml"},
		Config: "",
	}

	_, err := ansibleutil.RunAnsible(cmd)
	if err != nil {
		return false, err
	}

	return true, err
}

//
func (d *Deployer) Execute(cmd jettypes.DeployerCmd) (bool, error) {

	switch cmd {
	case jettypes.AnsiblePlaybook:
		return true, d.AnsiblePlaybookCmd()
	case jettypes.AnsibleFiles:
	case jettypes.AnsibleInventory:
		return d.AnsibleInventoryCmd()
	case jettypes.AnsibleDeploy:
		return d.AnsibleDeployCmd()
	case jettypes.VimNetworking:
	case jettypes.VimCompute:
	default:
		return false, fmt.Errorf("unknonw command")
	}

	return false, nil
}

// stop deployment
func (d *Deployer) Stop() {
	d.vim.Database().Close()
}
