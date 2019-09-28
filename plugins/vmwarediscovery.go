package main

import (
	"fmt"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/nsxtapi"
	"log"
)

/**

 */
func (p *VmwareVim) DiscoverNetworkElements() error {

	err := p.discoveryCluster()
	if err != nil {
		return err
	}

	err = p.discoveryTransportZone()
	if err != nil {
		return err
	}

	return nil
}

/*
   Discovers transport zone
*/
func (p *VmwareVim) discoveryTransportZone() error {

	if len(p.nsxtConfig.OverlayTransportName()) == 0 {
		err := fmt.Errorf("can't discover without tranport name, please check configuration")
		logging.ErrorLogging(err)
		return err
	}

	transport, err := nsxtapi.FindTransportZone(p.GetNsx(), p.nsxtConfig.OverlayTransportName())
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}

	if len(transport) == 0 {
		err := fmt.Errorf("couldn't find transport zone %s", p.nsxtConfig.OverlayTransportName())
		logging.ErrorLogging(err)
		return err
	}

	if len(transport) > 1 {
		err := fmt.Errorf("duplicate transport zone names. please rename transport zone")
		logging.ErrorLogging(err)
		return err
	}

	p.nsxtConfig.SetOverlayTzId(transport[0].Id)
	return nil
}

/**

 */
func (p *VmwareVim) discoverySwitching() error {

	if len(p.nsxtConfig.OverlayTransportName()) == 0 {
		err := fmt.Errorf("can't discover without edge cluster name or uuid")
		logging.ErrorLogging(err)
		return err
	}

	return nil
}

/**

 */
func (p *VmwareVim) discoveryDhcp() error {

	segment := p.nsxtConfig.LogicalSwitch()

	// Find target logical switch for a k8s cluster -- this setting read global DHCP shared
	// by entire cluster
	logicalSwitch, err := nsxtapi.FindLogicalSwitch(p.GetNsx(), segment, nil)
	if err != nil {
		return fmt.Errorf("can't find logical switch %s error %s. please check configuration",
			segment, err)
	}

	// Bind a DHCP server for k8s cluster
	logicalDhcpServer, err := nsxtapi.FindAttachedDhcpServerProfile(p.GetNsx(), logicalSwitch.Id)
	if err != nil {
		return fmt.Errorf("can't find a dhcp server attached to logical switch error: %s", err)
	}

	p.nsxtConfig.SetOverlayTzId(logicalSwitch.TransportZoneId)
	p.nsxtConfig.SetDhcpUuid(logicalDhcpServer.Id)

	return nil
}

/**

 */
func (p *VmwareVim) discoveryCluster() error {

	config := p.nsxtConfig

	if len(config.EdgeCluster()) == 0 {
		err := fmt.Errorf("can't discover without edge cluster name or uuid")
		logging.ErrorLogging(err)
		return err
	}

	edgecmp := nsxtapi.EdgeClusterCallback[nsxtapi.SearchByName]
	edgeCluster, err := nsxtapi.FindEdgeCluster(p.GetNsx(), edgecmp, config.EdgeCluster())
	if err != nil {
		e := fmt.Errorf("couldn't find edge cluster %s error: %v", config.EdgeCluster(), err)
		logging.ErrorLogging(e)
		return err
	}
	if len(edgeCluster) == 0 {
		return fmt.Errorf("cluster not found")
	}
	if len(edgeCluster) > 1 {
		log.Fatal("NSX-T have duplicate names for an edge cluster, please indicate cluster uuid instead of a name")
	}

	// set cluster id and discover all edge and routers
	config.SetEdgeClusterUuid(edgeCluster[0].Id)
	routers, err := nsxtapi.FindLogicalRouter(p.GetNsx(), nsxtapi.RouterCallback["edgeid"], edgeCluster[0].Id)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}
	for _, v := range routers {
		if v.RouterType == "TIER0" {
			config.AddTierZero(v)
		} else {
			config.AddTierOne(v)
		}
	}
	return nil
}

func cleanPorts() {
	// for each router in the cluster check ports
	//for _, v := range routers {
	//	disconnectedPorts, err := nsxtapi.FindDisconnectedPorts(vim.GetNsx(), v.Id)
	//	if err != nil {
	//		logging.ErrorLogging(err)
	//		return err
	//	}
	//	log.Println("Port", disconnectedPorts, " disconnected, device", v.DisplayName, "id", v.Id)

	//for _, v := range disconnectedPorts {
	//	err := nsxtapi.DeleteLogicalPort(&nsxtClient, v)
	//	if err != nil {
	//		log.Fatal("Failed delete", err)
	//	}
	//	log.Println("Delete port", v)
	//}
	//	}
}
