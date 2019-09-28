/*
Copyright (c) 2019 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

NSX-T API integration. Wrapper around NSX-T routing APIs

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package nsxtapi

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"

	"github.com/spyroot/jettison/logging"
)

/*
 a request encapsulate a request to crate router object
 note a tag is mandatory element in jettison since all object
 tracked by internal identifier.
*/
type RouterCreateReq struct {
	Name       string
	RouterType string
	ClusterID  string
	Tags       []common.Tag
}

const (
	RouteTypeTier0 = "TIER0"
	RouteTypeTier1 = "TIER1"
)

/*
 create new resource reference
*/
func newResourceReference(resourceType string, resourceID string) *common.ResourceReference {
	return &common.ResourceReference{
		TargetType: resourceType,
		TargetId:   resourceID,
	}
}

/**
  Function create nsx-t logical router.
  A tier zero must have cluster id.
*/
func CreateLogicalRouter(nsxClient *nsxt.APIClient, request RouterCreateReq) (*manager.LogicalRouter, error) {

	if request.RouterType == RouteTypeTier0 && len(request.ClusterID) == 0 {
		return nil, fmt.Errorf("tier zero router without cluster id")
	}

	if !IsUuid(request.ClusterID) {
		return nil, fmt.Errorf("cluster id must be a valid nsx-t uuid")
	}

	logicalRouter := manager.LogicalRouter{
		Description:   "jettison",
		DisplayName:   request.Name,
		RouterType:    request.RouterType,
		Tags:          request.Tags,
		EdgeClusterId: request.ClusterID,
	}

	router, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouter(nsxClient.Context, logicalRouter)
	if err != nil {
		return nil, fmt.Errorf("failed creater a logical routers: %s err %v", request.Name, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed create routers : %v", err)
	}

	return &router, nil
}

/**
  Finds router by tags attached to router object.  Please note since we using all tags caller must make sure
  that all router contains unique or context specific tags, otherwise method will return only
  first object that conforms a tags.  Jettison uses tenant id as tag to differentiate different project
*/
func FindLogicalRouterByTag(nsxClient *nsxt.APIClient, tags []common.Tag) (*manager.LogicalRouter, error) {

	routers, _, err := nsxClient.LogicalRoutingAndServicesApi.ListLogicalRouters(nsxClient.Context, nil)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, fmt.Errorf("failed obtain routers list : %v", err)
	}
	for i, router := range routers.Results {
		if reflect.DeepEqual(router.Tags, tags) {
			return &routers.Results[i], nil
		}
	}

	return nil, &ObjectNotFound{"logical router not found"}
}

/**
  Find router by field.  Caller provide compare callback and method dispatch
  manager.LogicalRouter to call back with original search value.
*/
func FindLogicalRouter(nsxClient *nsxt.APIClient,
	handler RouterSearchHandler, searchVal string) ([]*manager.LogicalRouter, error) {

	// Find the object by name
	var result []*manager.LogicalRouter
	routers, resp, err := nsxClient.LogicalRoutingAndServicesApi.ListLogicalRouters(nsxClient.Context, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain routers list : err %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to obtain routers list http error : %v", err)
	}

	for i, router := range routers.Results {
		if handler(&router, searchVal) {
			result = append(result, &routers.Results[i])
		}
	}
	if len(result) == 0 {
		return nil, &ObjectNotFound{"logical router not found, search term: " + searchVal}
	}

	return result, nil
}

/*
  Search routed port attached to a given router, tag used to limit a scope and find only
  ports that belong to a tenant.
*/
func FindRoutedPort(nsxClient *nsxt.APIClient, routerID string, tag *common.Tag) ([]*manager.LogicalRouterPort, error) {

	var result []*manager.LogicalRouterPort

	filter := map[string]interface{}{
		"logicalRouterId": routerID,
	}

	ports, resp, err := nsxClient.LogicalRoutingAndServicesApi.ListLogicalRouterPorts(nsxClient.Context, filter)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to obtain router ports list http error : %v", err)
	}

	// we check all ports that belong to a tenant
	for i, port := range ports.Results {
		for _, v := range port.Tags {
			if v.Scope == tag.Scope && v.Tag == tag.Tag {
				result = append(result, &ports.Results[i])
			}
		}
	}
	if len(result) == 0 {
		return nil, &ObjectNotFound{" router logical port not found " + routerID}
	}

	return result, nil
}

/*
  Search routed port attached to a given router, tag used to limit a scope and find only
  port that belong to a tenant.
*/
func FindRoutedPortByTag(nsxClient *nsxt.APIClient,
	routerID string, tags []common.Tag) (*manager.LogicalRouterPort, error) {

	filter := map[string]interface{}{
		"logicalRouterId": routerID,
	}

	ports, _, err := nsxClient.LogicalRoutingAndServicesApi.ListLogicalRouterPorts(nsxClient.Context, filter)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, nil
	}

	// we check all ports that belong to a tenant and already attached or not
	for i, port := range ports.Results {
		if reflect.DeepEqual(port.Tags, tags) {
			return &ports.Results[i], nil
		}
	}

	return nil, &ObjectNotFound{"Port not found"}
}

/*
  Search a routed port on tier1.
*/
func FindTierOnePortByTag(nsxClient *nsxt.APIClient,
	routerID string, tags []common.Tag) (*manager.LogicalRouterLinkPortOnTier1, error) {

	// find a logical port
	port, err := FindRoutedPortByTag(nsxClient, routerID, tags)
	if err != nil {
		if err, ok := err.(*ObjectNotFound); !ok {
			// log that
			logging.ErrorLogging(err)
		}
		return nil, err
	}
	// read a tier 1 as LogicalRouterLinkPortOnTier1
	tier1port, resp, err :=
		nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier1(nsxClient.Context, port.Id)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("failed read tier one port details port id %s, %v", port.Id, err)
		logging.ErrorLogging(e)
		return nil, e
	}

	return &tier1port, nil
}

/*
  Search a routed port on exiting tier 0 by a tag
*/
func FindTierZeroPortByTag(nsxClient *nsxt.APIClient,
	routerID string, tags []common.Tag) (*manager.LogicalRouterLinkPortOnTier0, error) {

	// find a logical port
	port, err := FindRoutedPortByTag(nsxClient, routerID, tags)
	if err != nil {
		if err, ok := err.(*ObjectNotFound); !ok {
			logging.ErrorLogging(err)
		}
		return nil, err
	}
	// read a tier 0 as LogicalRouterLinkPortOnTier1
	tier1port, resp, err :=
		nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier0(nsxClient.Context, port.Id)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("failed read tier zero port %s, %v", port.Id, err)
		logging.ErrorLogging(e)
		return nil, e
	}

	return &tier1port, nil
}

/**
  Find router by field.  Caller provide compare callback and method dispatch
  manager.LogicalRouter to call back with original search value.
*/
func FindEdgeCluster(nsxClient *nsxt.APIClient, handler EdgeSearchHandler, searchVal string) ([]*manager.EdgeCluster, error) {

	if len(searchVal) == 0 {
		return nil, fmt.Errorf("search term is empty")
	}

	var result []*manager.EdgeCluster
	edgeClusters, _, err := nsxClient.NetworkTransportApi.ListEdgeClusters(nsxClient.Context, nil)
	if err != nil {
		return nil, fmt.Errorf("failed obtain routers list : %v", err)
	}
	for i, v := range edgeClusters.Results {
		if handler(&v, searchVal) {
			result = append(result, &edgeClusters.Results[i])
		}
	}

	if len(result) == 0 {
		return nil, &ObjectNotFound{}
	}

	return result, nil
}

/**
  Deletes router from NSX-T.  It accepts both uuid or name,
  Note if we have duplicate name it will delete only the first router with given name.
*/
func DeleteLogicalRouter(nsxClient *nsxt.APIClient, routerName string) (bool, error) {

	opt := map[string]interface{}{
		"force": true,
	}

	// Find the object by name
	if IsUuid(routerName) {
		responseCode, err := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouter(
			nsxClient.Context, routerName, opt)
		if err != nil {
			return false, fmt.Errorf("failed delete router %s deletion: %v", routerName, err)
		}
		if responseCode.StatusCode != http.StatusOK {
			return false, fmt.Errorf("failed delete router %s deletion: %v",
				routerName, responseCode.StatusCode)
		}

		return true, nil
	}

	// otherwise it name so we do lookup
	routers, _, err := nsxClient.LogicalRoutingAndServicesApi.ListLogicalRouters(
		nsxClient.Context, nil)
	if err != nil {
		return false, fmt.Errorf("failed obtain routers list : %v", err)
	}
	for _, router := range routers.Results {
		if router.DisplayName != routerName {
			continue
		}
		// delete and return
		responseCode, err := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouter(
			nsxClient.Context, router.Id, opt)
		if err != nil {
			return false, fmt.Errorf("failed delete router %s deletion: %v", routerName, err)
		}
		if responseCode.StatusCode != http.StatusOK {
			return false, fmt.Errorf("failed delete router %s deletion: %v",
				routerName, responseCode.StatusCode)
		}
		return true, nil
	}
	return false, &ObjectNotFound{"router " + routerName + " not found"}
}

/*
  Function create a routed port and attaches to a existing switch.
  Caller must create a logical switch and logical port prior calling
  this function. It doesn't do validation and will return
  error if ports not defined.

  Caller must indicate first none network or broadcast in address in subnet section.
  For example if client wish to attach 172.16.1.0/24 as port
  it must do 172.16.1.1-254

*/
func CreateRoutedPort(nsxClient *nsxt.APIClient,
	routerID string,
	switchPortId string,
	subnet string,
	prefixLen int64,
	tags []common.Tag) (*manager.LogicalRouterDownLinkPort, error) {

	if len(routerID) == 0 || len(switchPortId) == 0 || len(subnet) == 0 || len(tags) == 0 {
		return nil, fmt.Errorf("all function argument is mandantory")
	}

	if prefixLen >= 32 || prefixLen < 8 {
		return nil, fmt.Errorf("invalid prefix len")
	}
	if nsxClient == nil {
		return nil, fmt.Errorf("nsx client context is nil")
	}

	logicalSwitch := manager.LogicalRouterDownLinkPort{
		Description:               "jettison",
		DisplayName:               "downlink-" + switchPortId,
		LogicalRouterId:           routerID,
		Tags:                      tags,
		LinkedLogicalSwitchPortId: newResourceReference("LogicalPort", switchPortId),
		Subnets: []manager.IpSubnet{
			{
				IpAddresses: []string{
					subnet,
				},
				PrefixLength: prefixLen,
			},
		},
	}

	logicalSwitch, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouterDownLinkPort(nsxClient.Context, logicalSwitch)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, fmt.Errorf("error during routed port creation: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		logging.ErrorLogging(err)
		return nil, fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
	}

	return &logicalSwitch, nil
}

/*
 Deletes local port from a router.
*/
func DeleteRoutedPort(nsxClient *nsxt.APIClient, portId string) error {

	if !IsUuid(portId) {
		return fmt.Errorf("routed port must be valid uuid")
	}

	resp, err :=
		nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouterPort(nsxClient.Context, portId, nil)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("error during routed port creation: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		logging.ErrorLogging(err)
		return fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

/**
  Finds all disconnected ports and return a slice that hold all ids.
*/
func FindDisconnectedPorts(nsxClient *nsxt.APIClient, routerID string) ([]string, error) {

	var disconnectedPorts []string

	if len(routerID) == 0 {
		err := fmt.Errorf("empty router id")
		logging.ErrorLogging(err)
		return disconnectedPorts, err
	}

	filter := map[string]interface{}{
		"logicalRouterId": routerID,
	}
	ports, resp, err := nsxClient.LogicalRoutingAndServicesApi.ListLogicalRouterPorts(nsxClient.Context, filter)
	if err != nil {
		logging.ErrorLogging(err)
		return disconnectedPorts, fmt.Errorf("error during router port listing: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		logging.ErrorLogging(err)
		return disconnectedPorts, fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
	}

	router, resp, err :=
		nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, routerID)
	if err != nil {
		logging.ErrorLogging(err)
		return disconnectedPorts, fmt.Errorf("failed read router object %s", router)
	}
	if resp.StatusCode != http.StatusOK {
		logging.ErrorLogging(err)
		return disconnectedPorts, fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
	}

	// tier 0 case
	if router.RouterType == "TIER0" {
		for _, v := range ports.Results {
			port, _, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier0(nsxClient.Context, v.Id)
			if err != nil {
				logging.ErrorLogging(err)
				return disconnectedPorts, fmt.Errorf("error during routed port creation: %v", err)
			}
			// port not linked to anything
			if len(port.LinkedLogicalRouterPortId) == 0 {
				disconnectedPorts = append(disconnectedPorts, port.Id)
				log.Println("Port", v.Id, " disconnected")
			}
		}
		return disconnectedPorts, nil
	}

	// tier 1 case
	for _, v := range ports.Results {
		port, _, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier1(nsxClient.Context, v.Id)
		if err != nil {
			logging.ErrorLogging(err)
			return disconnectedPorts, fmt.Errorf("error during routed port creation: %v", err)
		}
		// TODO  create a test for this this case
		if port.LinkedLogicalRouterPortId != nil {
			if len(port.LinkedLogicalRouterPortId.TargetId) == 0 {
				disconnectedPorts = append(disconnectedPorts, port.Id)
				log.Println("Port", v.Id, " disconnected")
			}
		}
	}

	return disconnectedPorts, nil
}

/**
	Because NSX-T API use explicitly tier 1 or tier zero we need duplicate methods
    otherwise overcompensate with reflection

*/
func CreateTierOneUplinkIfNeed(nsxClient *nsxt.APIClient,
	routerId string, tags []common.Tag) (*manager.LogicalRouterLinkPortOnTier1, error) {

	// search for existing port on tier 1
	port, err := FindTierOnePortByTag(nsxClient, routerId, tags)
	if err == nil {
		return port, err
	}
	// port not found create a new port
	if _, ok := err.(*ObjectNotFound); ok {
		newPort := manager.LogicalRouterLinkPortOnTier1{
			Description:     "jettison",
			DisplayName:     "uplink-tier0",
			ResourceType:    "",
			Tags:            tags,
			LogicalRouterId: routerId,
		}
		newPort, resp, err :=
			nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouterLinkPortOnTier1(nsxClient.Context, newPort)
		if err != nil {
			logging.ErrorLogging(err)
			return nil, fmt.Errorf("error during routed port creation for tier one %s: %v", routerId, err)
		}
		if resp.StatusCode != http.StatusCreated {
			logging.ErrorLogging(err)
			return nil, fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
		}
		return &newPort, nil
	}

	return nil, err
}

/**
  dst port is port where tier 0 will be connected.
*/
func CreateTierZeroUplinkIfNeed(nsxClient *nsxt.APIClient,
	routerId string, dstPort string, tags []common.Tag) (*manager.LogicalRouterLinkPortOnTier0, error) {

	port, err := FindTierZeroPortByTag(nsxClient, routerId, tags)
	if err == nil {
		return port, nil
	}

	if _, ok := err.(*ObjectNotFound); ok {
		newDstPort := manager.LogicalRouterLinkPortOnTier0{
			Description:               "jettison",
			DisplayName:               "downlink-tier1",
			ResourceType:              "",
			Tags:                      tags,
			LogicalRouterId:           routerId,
			LinkedLogicalRouterPortId: dstPort,
		}
		newPort, resp, err :=
			nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouterLinkPortOnTier0(nsxClient.Context, newDstPort)
		if err != nil {
			logging.ErrorLogging(err)
			return nil, fmt.Errorf("error during routed port creation on tier zero: %s %v", routerId, err)
		}
		if resp.StatusCode != http.StatusCreated {
			logging.ErrorLogging(err)
			return nil, fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
		}
		return &newPort, nil
	}

	return nil, err
}

/*
  Function create two routed port, one on tier one and another port on tier zero
  and connects both routers,
  At first it check if tier1 port already exists if it does it will re-use same port,
  the same for tier zero.

  TODO Graph maybe
*/
func CreateUplinkPort(nsxClient *nsxt.APIClient,
	tierOneUuid, tierZeroUuid string, srcTag []common.Tag, dstTag []common.Tag) (string, string, error) {

	if !IsUuid(tierOneUuid) {
		return "", "", fmt.Errorf("routed id must be valid uuid format")
	}
	if !IsUuid(tierZeroUuid) {
		return "", "", fmt.Errorf("routed id must be valid uuid format")
	}

	srcPort, err := CreateTierOneUplinkIfNeed(nsxClient, tierOneUuid, srcTag)
	if err != nil {
		logging.ErrorLogging(err)
		return "", "", err
	}
	dstPort, err := CreateTierZeroUplinkIfNeed(nsxClient, tierZeroUuid, srcPort.Id, dstTag)
	if err != nil {
		logging.ErrorLogging(err)
		return "", "", err
	}

	// create binding between tier 1 and tier 0
	srcPort.LinkedLogicalRouterPortId = newResourceReference("LogicalPort", dstPort.Id)
	// update tier one and point to tier zero
	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateLogicalRouterLinkPortOnTier1(nsxClient.Context, srcPort.Id, *srcPort)
	if err != nil {
		logging.ErrorLogging(err)
		return "", "", fmt.Errorf("error during routed port creation: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		logging.ErrorLogging(err)
		return "", "", fmt.Errorf("nsx-t return unexpected status code: %v", resp.StatusCode)
	}

	return srcPort.Id, dstPort.Id, nil
}

/**
  Update default routing advertisement policy
*/
func DefaultRoutingAdvertisement(nsxClient *nsxt.APIClient, routerId string) (bool, error) {

	if !IsUuid(routerId) {
		return false, fmt.Errorf("routed id must be valid uuid format")
	}

	policy, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadAdvertisementConfig(nsxClient.Context, routerId)
	if err != nil {
		logging.ErrorLogging(err)
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		logging.ErrorLogging(err)
	}

	policy.Enabled = true
	policy.AdvertiseNsxConnectedRoutes = true
	policy.AdvertiseStaticRoutes = true
	policy.AdvertiseLbSnatIp = true
	policy.AdvertiseLbVip = true

	policy, resp, err = nsxClient.LogicalRoutingAndServicesApi.UpdateAdvertisementConfig(nsxClient.Context, routerId, policy)
	if err != nil {
		logging.ErrorLogging(err)
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code for update rt advertisement config")
		logging.ErrorLogging(e)
		return false, e
	}

	return true, nil
}

//
//  Adds static route
//
func AddStaticRoute(nsxClient *nsxt.APIClient, req AddStaticReq) (bool, error) {

	if !IsUuid(req.RouterUuid) {
		return false, fmt.Errorf("routed id must be valid uuid format")
	}

	nextHop := manager.StaticRouteNextHop{}
	nextHop.IpAddress = req.NextHopAddr.String()
	nextHop.AdministrativeDistance = 1
	nextHop.BfdEnabled = false

	staticRoute := manager.StaticRoute{}
	staticRoute.DisplayName = ""
	staticRoute.LogicalRouterId = req.RouterUuid
	staticRoute.Network = req.Network
	staticRoute.NextHops = append(staticRoute.NextHops, nextHop)

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.AddStaticRoute(nsxClient.Context,
		req.RouterUuid, staticRoute)
	if err != nil {
		logging.ErrorLogging(err)
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code for req add static route")
		logging.ErrorLogging(e)
		return false, e
	}

	return true, nil
}
