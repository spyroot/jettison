package main

import (
	"fmt"
	"github.com/spyroot/jettison/vcenter"
	"log"
	"net"

	"github.com/google/uuid"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/netpool"
	"github.com/spyroot/jettison/nsxtapi"
)

//  Dhcp clean up process check each node struct mac address if mac address in the struct
//  it will use that to remove dhcp binding, if not it uses vm id to find object and
//  figure out mac allocated for that vm.
func (p *VmwareVim) DhcpCleanup(projectName string, nodes []*jettypes.NodeTemplate) error {

	for _, node := range nodes {
		if len(node.Mac) == 0 {
			logging.CriticalMessage("node", node.Name, "has no mac address associated")
			continue
		}
		err := nsxtapi.DhcpCleanupEntry(p.GetNsx(), node)
		if err != nil {
			// get the VM from VIM and check if VM has a device with mac
			_, _, vm, err := vcenter.VmFromCluster(p.ctx, p.VimClient(), node.Name, node.VimCluster)
			if err == nil {
				dev, _ := vm.Device(p.ctx)
				if dev.PrimaryMacAddress() == node.Mac[0] {
					logging.CriticalMessage("No dhcp bindings but mac address attached to VM")
					continue
				}
			}
			continue
		}
	}

	return nil
}

//
//  a) Deploys network element a logical switch per segment or single logical switch
//     shared by entire deployment.
//
//   b) A logical route tier 1
func (p *VmwareVim) DeploySegment(projectName string, segmentName string,
	gateway string, prefixLen int) (*jettypes.GenericSwitch, *jettypes.GenericRouter, error) {

	tenantName := projectName
	overlayID := p.nsxtConfig.OverlayTransportUuid()
	clusterID := p.nsxtConfig.EdgeClusterUuid()
	switchName := uuid.New().String()
	routerName := uuid.New().String()

	tierZero, err := p.nsxtConfig.GetActiveTierZero()
	if err != nil {
		return nil, nil, err
	}

	// create logical switch if need
	switchID, switchName, err := nsxtapi.CreateSwitchIfNeed(p.GetNsx(),
		projectName, segmentName, overlayID, switchName)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, nil, err
	}

	// create logical router if need
	routerID, err := nsxtapi.CreateRouterIfNeed(p.GetNsx(), tenantName, segmentName, routerName, clusterID)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, nil, err
	}

	// bind a switch to a router downlink
	logicalPortId, err := nsxtapi.CreateLogicalPortIfNeed(p.GetNsx(), tenantName, switchID, routerID)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, nil, err
	}

	// bind router to a switch that out downlink per segment
	downlinkPort, err := nsxtapi.CreateRoutedPortIfNeed(p.GetNsx(),
		tenantName, routerID, switchID, logicalPortId, gateway, int64(prefixLen))
	if err != nil {
		logging.ErrorLogging(err)
		return nil, nil, err
	}

	_, _, err = nsxtapi.ConnectTier1IfNeed(p.GetNsx(), tenantName, routerID, tierZero)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, nil, err
	}

	_, err = nsxtapi.DefaultRoutingAdvertisement(p.GetNsx(), routerID)
	if err != nil {
		log.Fatal(err)
	}

	gwAddr := net.ParseIP(gateway)
	if gwAddr == nil {
		e := fmt.Errorf("invalid gateway format")
		logging.ErrorLogging(err)
		return nil, nil, e
	}

	dhcpAddr := netpool.NextIP(gwAddr, 1)
	fmtDhcp := fmt.Sprintf("%s/%d", dhcpAddr.String(), prefixLen)

	var dhcpReq = &nsxtapi.DhcpServerCreateReq{
		ServerName:     tenantName + "-" + segmentName,
		DhcpServerIp:   fmtDhcp,
		DnsNameservers: []string{"8.8.8.8"},
		DomainName:     "vmwarelab.edu",
		GatewayIp:      gateway,
		ClusterId:      clusterID,
		SwitchId:       switchID,
		TenantId:       tenantName,
		Segment:        segmentName,
	}

	dhcpServerId, err := nsxtapi.CreateDhcpServiceIfNeed(p.GetNsx(), dhcpReq)
	if err != nil {
		log.Fatal(err)
	}

	logicalSwitch := jettypes.NewGenericSwitch(switchName, switchID, dhcpServerId, routerID)
	logicalSwitch.SetRouterPortUuid(logicalPortId)

	logicalRouter := jettypes.NewGenericRouter(routerName, routerID)
	logicalRouter.SetSwitchPortUuid(downlinkPort)

	return logicalSwitch, logicalRouter, err
}

/**

 */
func (p *VmwareVim) DiscoverClusterDhcpServer(projectName string, nodes *[]*jettypes.NodeTemplate) (bool, error) {

	// find target logical switch for a k8s cluster -- this setting read global DHCP shared
	// by entire cluster
	for k, node := range *nodes {
		logicalSwitch, err := nsxtapi.FindLogicalSwitch(p.GetNsx(), node.GenericSwitch().Name(), nil)
		if err != nil {
			return false, fmt.Errorf("can't find logical switch %s error %s. "+
				"please check configuration", node.GenericSwitch().Name(), err)
		}

		// we set switch id after discovery
		(*nodes)[k].GenericSwitch().SetUuid(logicalSwitch.Id)
		log.Println("Discovery...", node.GenericSwitch().Name(), " uuid ", logicalSwitch.Id)
		// find a dhcp server for k8s cluster
		logicalDhcpServer, err := nsxtapi.FindAttachedDhcpServerProfile(p.GetNsx(), logicalSwitch.Id)
		if err != nil {
			return false, fmt.Errorf("can't find a dhcp server attached to logical switch error: %s", err)
		}

		// set dhcp server id
		(*nodes)[k].GenericSwitch().SetDhcpUuid(logicalDhcpServer.Id)
		// TODO set tier-1 router/ check default gateway IP/ tier-1 router must be attached to lds.
	}

	return true, nil
}

/**
  Create dhcp binding for all nodes
*/
func (p *VmwareVim) CreateDhcpBindings(projectName string, nodes []*jettypes.NodeTemplate) error {

	for _, node := range nodes {
		err := p.SelectDhcpBinding(projectName, node)
		if err != nil {
			log.Println("failed create dhcp bind for", node.Name, err)
			return err
		}
	}

	return nil
}

//
// Return existing DHCP binding
//
func (p *VmwareVim) SelectDhcpBinding(projectName string, node *jettypes.NodeTemplate) error {

	if len(node.Mac[0]) == 0 {
		return fmt.Errorf("node has no mac address")
	}

	dhcpId := node.GenericSwitch().DhcpUuid()
	log.Println("Creating binding for node ", node.Name, node.Mac, node.IPv4Addr.String())

	// lookup dhcp binding
	dhcpBinding, err := nsxtapi.GetStaticBinding(p.GetNsx(), dhcpId, node.Mac[0], nsxtapi.DhcpLookupHandler["mac"])
	// binding already in system.
	if err == nil {
		node.DhcpStatus = jettypes.Created
		if dhcpBinding.IpAddress == node.IPv4Addr.String() {
			logging.Notification("Found existing binding for controller node: " + node.Name)
			node.DhcpStatus = jettypes.Created
			return nil
		}
	}

	// check that we don't have IP attached to anything
	val, err := nsxtapi.GetStaticBinding(p.GetNsx(), dhcpId, node.IPv4Addr.String(), nsxtapi.DhcpLookupHandler["ip"])
	// TODO refactor to object not found
	if err != nil {
		// IP attached not in use, we can create static binding
		_, err := nsxtapi.CreateStaticBinding(p.GetNsx(),
			dhcpId,
			node.Mac[0],
			node.IPv4Addr.String(),
			node.Name,
			node.Gateway,
			projectName)
		log.Println("Created binding for", node.Name)
		if err != nil {
			return fmt.Errorf("failed create static dhcp binding")
		}
	} else {
		// we can two case a) left over for a same node b) someone already allocate same IP to a node
		if val.HostName != node.Name {
			logging.CriticalMessage("Failed create binding. Another host",
				val.HostName, "mac address", val.MacAddress, "already has a binding", node.DesiredAddress)
			return fmt.Errorf("failed %s create binding another host dhcp conflict", node.DesiredAddress)
		} else {
			log.Println("Host", val.HostName, "already has dhcp static binding", node.DesiredAddress)
		}
	}

	return nil
}

func (p *VmwareVim) DeleteDhcpServer(node *jettypes.NodeTemplate) (bool, error) {
	profileId, _, err := nsxtapi.DeleteDhcpServer(p.GetNsx(), node.DhcpServerUuid())
	if err != nil {
		return false, err
	}

	ok, _, err := nsxtapi.DeleteDhcpProfile(p.GetNsx(), profileId)
	if err != nil {
		return false, err
	}

	return ok, err
}

// Implementation of plugin that use nsx-t api interface in router
// delete semantics. It deletes a logical router with force flag
// that will remove all attached ports
func (p *VmwareVim) DeleteRouter(node *jettypes.NodeTemplate) (bool, error) {
	return nsxtapi.DeleteLogicalRouter(p.GetNsx(), node.DhcpServerUuid())
}

// Implementation that use nsx-t to delete a logical switch with force flag
// that will remove all attached ports
// TODO split logic between nsx or dvs
func (p *VmwareVim) DeleteSwitch(node *jettypes.NodeTemplate) (bool, error) {
	return nsxtapi.DeleteLogicalSwitch(p.GetNsx(), node.DhcpServerUuid())
}

// Implementation that use nsx-t to add a static
// route to a given tier 1 router
func (p *VmwareVim) AddStaticRoute(projectName string,
	node *jettypes.NodeTemplate, podNetwork string) (bool, error) {

	req := nsxtapi.AddStaticReq{}
	req.RouterUuid = node.GenericRouter().Uuid()
	req.Network = podNetwork
	req.NextHopAddr = node.IPv4Addr

	return nsxtapi.AddStaticRoute(p.GetNsx(), req)
}
