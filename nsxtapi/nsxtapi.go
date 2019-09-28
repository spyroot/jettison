package nsxtapi

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"

	"github.com/spyroot/jettison/logging"
)

/*
https://vmware.github.io/vsphere-automation-sdk-java/nsx/nsx-policy/constant-values.html#com.vmware.nsx_policy.model.LogicalRouter.HIGH_AVAILABILITY_MODE_STANDBY
*/
const ResourceTypeRoutedPort = "LogicalRouterPort"
const ResourceTypeDownlink = "LogicalRouterDownLinkPort"
const ResourceTypeUplink = "LogicalRouterUpLinkPort"

const HaActiveActive = "ACTIVE_ACTIVE"

const HaActiveStandby = "ACTIVE_STANDBY"

//public static final String	RESOURCE_TYPE_LOGICALROUTERCENTRALIZEDSERVICEPORT	"LogicalRouterCentralizedServicePort"
//public static final String	RESOURCE_TYPE_LOGICALROUTERDOWNLINKPORT	"LogicalRouterDownLinkPort"
//public static final String	RESOURCE_TYPE_LOGICALROUTERIPTUNNELPORT	"LogicalRouterIPTunnelPort"
//public static final String	RESOURCE_TYPE_LOGICALROUTERLINKPORTONTIER0	"LogicalRouterLinkPortOnTIER0"
//public static final String	RESOURCE_TYPE_LOGICALROUTERLINKPORTONTIER1	"LogicalRouterLinkPortOnTIER1"
//public static final String	RESOURCE_TYPE_LOGICALROUTERLOOPBACKPORT	"LogicalRouterLoopbackPort"
//public static final String	RESOURCE_TYPE_LOGICALROUTERUPLINKPORT	"LogicalRouterUpLinkPort"

//public static final String	ADMIN_STATE_DOWN	"DOWN"
//public static final String	ADMIN_STATE_UP	"UP"
//public static final String	REPLICATION_MODE_MTEP	"MTEP"
//public static final String	REPLICATION_MODE_SOURCE	"SOURCE"

/*public static final String	STATE_FAILED	"failed"
public static final String	STATE_IN_PROGRESS	"in_progress"
public static final String	STATE_ORPHANED	"orphaned"
public static final String	STATE_PARTIAL_SUCCESS	"partial_success"
public static final String	STATE_PENDING	"pending"
public static final String	STATE_SUCCESS	"success"
public static final String	STATE_UNKNOWN	"unknown"
*/

//Interface that caller need implement for DHCP lease request,
//currently used in DHCP lease/delete sequence.  Each field is minim
//set that needs to identify a right DHCP binding
type DhcpLeasEntry interface {
	// server uuid
	DhcpServerUuid() string
	// mac address a string.  recommendation serialize to net.MAC and back to string
	MacAddress() string
	// ip address
	IPv4Address() net.IP
	// a switch uuid where that dhcp bounded to
	SwitchUuid() string
}

type AddStaticReq struct {
	// tier 0 or 1 uuid
	RouterUuid string
	// next hop for network
	NextHopAddr net.IP
	// optional port id
	PortUuid string
	// cidr format destination network
	Network string
}

func MakeCustomTags(scope string, val string) []common.Tag {

	var newTags = []common.Tag{
		{
			Scope: scope,
			Tag:   val,
		},
	}
	return newTags
}

func MakeDhcpTags(tenantName string) []common.Tag {

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantName,
		},
	}
	return newTags
}

func MakeSwitchTags(tenantName string, segmentName string) []common.Tag {

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantName,
		},
		{
			Scope: "segment",
			Tag:   segmentName,
		},
	}
	return newTags
}

func MakeDhcpProfileTag(tenantName, switchId string) []common.Tag {
	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantName,
		},
		{
			Scope: "segment", // profile per switch
			Tag:   switchId,
		},
	}

	return newTags
}

/**
  Function create switch if need, If a switch already present in NSX-T
  it will return existing switch id, otherwise it will create new switch and return new id and set a tag
  to project id.
*/
func CreateSwitchIfNeed(nsxtClient *nsxt.APIClient, tenantId, tenantSegment, overlayTz, switchName string) (string, string, error) {

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "segment",
			Tag:   tenantSegment,
		},
	}

	var actualSwitchName = tenantId + "-" + tenantSegment + "-" + switchName
	logicalSwitch, err := FindLogicalSwitchByTag(nsxtClient, newTags)
	if err != nil {
		if _, ok := err.(*ObjectNotFound); ok {
			logicalSwitch, createErr := CreateLogicalSwitch(nsxtClient, overlayTz, actualSwitchName, newTags)
			if createErr != nil {
				logging.ErrorLogging(createErr)
				return "", "", createErr
			}
			return logicalSwitch.Id, logicalSwitch.DisplayName, nil
		}
		return "", "", err
	}

	// we cross check that it a switch that belong to a tenant
	for _, v := range logicalSwitch.Tags {
		if v.Tag == tenantId {
			log.Println("Found exiting switch")
			return logicalSwitch.Id, logicalSwitch.DisplayName, nil
		}
	}

	return "", "", fmt.Errorf("failed create a switch please solve conflict")
}

/**
  Function creates tier 1 logical router if it not present otherwise just return
  id of exiting one.
*/
func CreateRouterIfNeed(nsxtClient *nsxt.APIClient, tenantId, tenantSegment, routerName, clusterID string) (string, error) {

	if !IsUuid(clusterID) {
		return "", fmt.Errorf("cluster must have valid uuid format %s", clusterID)
	}
	if len(routerName) == 0 || len(tenantId) == 0 || len(tenantSegment) == 0 {
		return "", fmt.Errorf("arguments can't have empty values")
	}

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "segment",
			Tag:   tenantSegment,
		},
	}

	var actualRouterName = tenantId + "-" + tenantSegment + "-" + routerName
	logicalRouter, err := FindLogicalRouterByTag(nsxtClient, newTags)
	if err == nil {
		return logicalRouter.Id, nil
	}

	// check if object not found then we create if we can
	if _, ok := err.(*ObjectNotFound); ok {

		req := RouterCreateReq{}
		req.Name = actualRouterName
		req.RouterType = RouteTypeTier1
		req.ClusterID = clusterID
		req.Tags = newTags

		logicalRouter, createErr := CreateLogicalRouter(nsxtClient, req)
		if createErr != nil {
			logging.ErrorLogging(createErr)
			return "", createErr
		}
		return logicalRouter.Id, nil
	}
	// other error nothing much we can do
	return "", err
}

/*
   Each time router attached we set tag to a port.
   - Check do we have port already or not based on tag if we do use it if don't create new one.
   - Check do w
*/
func CreateLogicalPortIfNeed(nsxtClient *nsxt.APIClient, tenantId, switchId, routerId string) (string, error) {

	var tagScope = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "routing",
			Tag:   routerId,
		},
	}

	// if we found all tenant ports on that switch in result
	logicalPorts, err := FindLogicalPort(nsxtClient, switchId, tagScope)
	if err != nil {
		// case one no port on logical switch for tenant
		if _, ok := err.(*ObjectNotFound); ok {
			logicalPort, createErr := CreateLogicalPort(nsxtClient, "Uplink-"+tenantId, switchId, tagScope)
			if createErr != nil {
				logging.ErrorLogging(createErr)
				return "", createErr
			}
			return logicalPort.Id, nil
		}
		// other error nothing much we can do
		return "", nil
	}

	// check do we have port already
	for _, port := range logicalPorts {
		for _, portTag := range port.Tags {
			// port already in logical switch
			if portTag.Scope == "routing" && portTag.Tag == routerId {
				return port.Id, nil
			}
		}
	}

	//we don't have port yet so we create new one.
	logicalPort, createErr := CreateLogicalPort(nsxtClient, "Uplink-"+tenantId, switchId, tagScope)
	if createErr != nil {
		logging.ErrorLogging(createErr)
		return "", createErr
	}

	return logicalPort.Id, nil
}

/*
   Each time router attached we set tag to a port.
   - Check do we have port already or not based on tag if we do use it if don't create new one.
*/
func CreateRoutedPortIfNeed(nsxtClient *nsxt.APIClient,
	tenantId, routerId, switchId, logicalPortId, subnet string, prefix int64) (string, error) {

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "switching", // we maintain own state
			Tag:   switchId,
		},
	}

	// if we found all tenant ports on that switch in result
	logicalPorts, err := FindRoutedPortByTag(nsxtClient, routerId, newTags)
	if err == nil {
		return logicalPorts.Id, nil
	}
	if _, ok := err.(*ObjectNotFound); ok {
		logicalPort, createErr := CreateRoutedPort(nsxtClient, routerId, logicalPortId, subnet, prefix, newTags)
		if createErr != nil {
			logging.ErrorLogging(createErr)
			return "", createErr
		}
		return logicalPort.Id, nil
	}

	// other error nothing much we can do
	return "", err
}

/*
   Each time router attached we set tag to a port.
   - Check do we have port already or not based on tag if we do use it if don't create new one.
   - Check do w
*/
func ConnectTier1IfNeed(nsxtClient *nsxt.APIClient, tenantId, tierOneId, tierZeroId string) (string, string, error) {

	if !IsUuid(tierOneId) {
		return "", "", fmt.Errorf("tier one id, needs to be uuid")
	}

	if !IsUuid(tierZeroId) {
		return "", "", fmt.Errorf("tier one id, needs to be uuid")
	}

	var tier1Tag = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "tierzero", // we maintain own state tier one point to tier zero
			Tag:   tierZeroId,
		},
	}

	var tierZeroTag = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "tierone", // we maintain own state tier zero points to tier one
			Tag:   tierOneId,
		},
	}

	return CreateUplinkPort(nsxtClient, tierOneId, tierZeroId, tier1Tag, tierZeroTag)
}

/*
   Each time router attached we set tag to a port.
   - Check do we have port already or not.
   - Lookup based on tag if we do use exiting port, if don't create new one that
     avoid re-crete port each time for a same network construct
*/
func CreateDhcpPortIfNeed(nsxtClient *nsxt.APIClient, tenantId, switchId string) (string, error) {

	var tagScope = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
		{
			Scope: "dhcp",
			Tag:   tenantId,
		},
	}
	// if we found tenants dhcp ports on that switch
	logicalPorts, err := FindLogicalPort(nsxtClient, switchId, tagScope)
	if err == nil {
		log.Print("Found existing port: ", logicalPorts[0].Id)
		return logicalPorts[0].Id, nil
	}
	// if port not found create a new one
	if _, ok := err.(*ObjectNotFound); ok {
		log.Print("Creating new port ")
		logicalPort, createErr := CreateLogicalPort(nsxtClient, "Dhcp-"+tenantId, switchId, tagScope)
		if createErr != nil {
			logging.ErrorLogging(createErr)
			return "", createErr
		}
		return logicalPort.Id, nil
	}

	log.Print("Creating new port ")

	return "", err
}

/*
   Create dhcp profile if it not already exists for a tenant and return profile id.
*/
func CreateDhcpProfileIfNeed(nsxtClient *nsxt.APIClient, req DhcpProfile) (string, error) {

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   req.TenantUuid(),
		},
		{
			Scope: "segment", // profile per switch
			Tag:   req.SwitchUuid(),
		},
	}

	profile, err := FindDhcpServerProfileByTag(nsxtClient, newTags)
	if err == nil {
		log.Print("Found existing profile: ", profile.Id)
		return profile.Id, nil
	}
	if _, ok := err.(*ObjectNotFound); ok {
		log.Print("Creating new dhcp profile")
		profileName := req.TenantUuid() + "-" + req.SegmentName() + "-" + req.SwitchUuid()
		newProfileId, createErr := CreateDhcpProfile(nsxtClient, req.ClusterUuid(), profileName, newTags)
		if createErr != nil {
			logging.ErrorLogging(createErr)
			return "", createErr
		}
		return newProfileId, nil
	}

	return "", err
}

func CreateDhcpServerIfNeed(nsxtClient *nsxt.APIClient, req *DhcpServerCreateReq) (string, error) {

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   req.TenantId,
		},
		{
			Scope: "segment", // we maintain own state
			Tag:   req.SwitchId,
		},
	}

	serverId, err := FindDhcpServerByTag(nsxtClient, newTags)
	if err == nil {
		log.Print("Found existing dhcp server:", serverId)
		return serverId, nil
	}
	if _, ok := err.(*ObjectNotFound); ok {
		log.Print("Creating new dhcp server for segment:", req.SwitchId)
		dhcpServer, createErr := CreateDhcpServer(nsxtClient, req, newTags)
		if createErr != nil {
			logging.ErrorLogging(createErr)
			return "", createErr
		}
		return dhcpServer.Id, nil
	}

	return "", err
}

/*
   - Each time router attached we set tag to a port.
   - Check do we have port already or not based on tag if we do use it if don't create new one.
*/
func CreateDhcpServiceIfNeed(nsxtClient *nsxt.APIClient, req *DhcpServerCreateReq) (string, error) {

	if !IsUuid(req.ClusterId) {
		return "", fmt.Errorf("invalid edge cluster id")
	}
	if !IsUuid(req.SwitchId) {
		return "", fmt.Errorf("invalid logical switch id")
	}

	if len(req.TenantId) == 0 {
		return "", fmt.Errorf("invalid tenant id")
	}

	if len(req.Segment) == 0 {
		return "", fmt.Errorf("invalid segment name")
	}

	if len(req.DhcpServerIp) == 0 {
		return "", fmt.Errorf("invalid dhcp ip")
	}

	_, _, err := net.ParseCIDR(req.DhcpServerIp)
	if err != nil {
		return "", fmt.Errorf("invalid dhcp server format")
	}

	log.Print("Creating new dhcp profile")
	profileId, err := CreateDhcpProfileIfNeed(nsxtClient, req)
	if err != nil {
		return "", err
	}

	log.Print("Creating logical port for dhcp")
	logicalPortID, err := CreateDhcpPortIfNeed(nsxtClient, req.TenantId, req.SwitchId)
	if err != nil {
		logging.ErrorLogging(err)
		return "", err
	}

	// dhcp server needs a profile and logical port where to attach
	req.DhcpProfileId = profileId
	req.LogicalPortID = logicalPortID
	log.Print("Creating dhcp server")
	dhcpServerId, err := CreateDhcpServerIfNeed(nsxtClient, req)
	if err != nil {
		logging.ErrorLogging(err)
		return "", err
	}

	//  logical port must have right attachments
	lp, resp, err := nsxtClient.LogicalSwitchingApi.GetLogicalPort(nsxtClient.Context, logicalPortID)
	if err != nil {
		logging.ErrorLogging(err)
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t returned wrong http code for get port: uuid %s", logicalPortID)
		logging.ErrorLogging(e)
		return "", e
	}

	if lp.Attachment == nil {
		logging.Notification("Attach logical port to dhcp server ", dhcpServerId)
		lp.Attachment = &manager.LogicalPortAttachment{
			AttachmentType: "DHCP_SERVICE",
			Id:             dhcpServerId,
		}
		_, _, err = nsxtClient.LogicalSwitchingApi.UpdateLogicalPort(nsxtClient.Context, logicalPortID, lp)
		if err != nil {
			logging.ErrorLogging(err)
			return "", err
		}
	} else {
		if lp.Attachment.Id == dhcpServerId {
			logging.Notification("Logical port already attached to dhcp server ", dhcpServerId)
		}
	}

	return dhcpServerId, nil
}

// clean up routine
func DhcpCleanupEntry(nsxtClient *nsxt.APIClient, entry DhcpLeasEntry) error {
	return DeleteStaticBinding(nsxtClient, entry.DhcpServerUuid(), entry.IPv4Address().String(), entry.MacAddress())
}
