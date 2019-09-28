package nsxtapi

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"

	"github.com/spyroot/jettison/logging"
)

// a interface that require to implement
// if a caller need create a dhcp profile
type DhcpProfile interface {
	// a tenant id or name
	TenantUuid() string
	// a logical switch where dhcp profile will be attach to
	SwitchUuid() string
	// a edge cluster id where dhcp profile will be attach to
	ClusterUuid() string
	// additional semantics attached to dhcp profile as tag
	SegmentName() string
}

const (
	DhcpMacLookup          = "mac"
	DhcpIpLookupHandler    = "ip"
	DhcpIpMacLookupHandler = "ipmac-pair"
)

type DhcpHandler func(manager.DhcpStaticBinding, string) bool

var DhcpLookupHandler = map[string]func(dhcpEntry manager.DhcpStaticBinding, val string) bool{
	DhcpMacLookup: func(dhcpEntry manager.DhcpStaticBinding, macAddr string) bool {
		if dhcpEntry.MacAddress == macAddr {
			return true
		}
		return false
	},
	DhcpIpLookupHandler: func(dhcpEntry manager.DhcpStaticBinding, ipAddr string) bool {
		if dhcpEntry.IpAddress == ipAddr {
			return true
		}
		return false
	},
	DhcpIpMacLookupHandler: func(dhcpEntry manager.DhcpStaticBinding, ipAddr string) bool {
		if dhcpEntry.IpAddress == ipAddr {
			return true
		}
		return false
	},
}

/*
  Packed struct for dhcp server create request.
  It has bunch of attributes so caller need make sure it populate all
  fields. That requirement imposed by NSX-T

  Note - profile must be created before server creation
  Note - logical port must be create before server attached.
*/
type DhcpServerCreateReq struct {
	// name for a server
	ServerName string
	// server ip in cidr format that must be attached and not conflicting with a gateway
	DhcpServerIp string
	// just list
	DnsNameservers []string
	// domain name
	DomainName string
	// gateway must be in same segment where tier 1 attached.
	GatewayIp string
	//
	Options string
	// a logical port uuid attached to lds
	LogicalPortID string
	// profile uuid
	DhcpProfileId string
	// edge custer id
	ClusterId string
	// switch id
	SwitchId string
	// tenant id
	TenantId string
	//
	Segment string
}

func (d *DhcpServerCreateReq) TenantUuid() string {
	if d != nil {
		return d.TenantId
	}
	return ""
}

func (d *DhcpServerCreateReq) SwitchUuid() string {
	if d != nil {
		return d.SwitchId
	}
	return ""
}

func (d *DhcpServerCreateReq) ClusterUuid() string {
	if d != nil {
		return d.ClusterId
	}
	return ""
}

func (d *DhcpServerCreateReq) SegmentName() string {
	if d != nil {
		return d.Segment
	}
	return ""
}

// Get static binding from a DHCP a handle provide a way to lookup
// based on IP or mac or mac and IP
func GetStaticBinding(nsxClient *nsxt.APIClient,
	serverId string, SearchVal string, fn DhcpHandler) (*manager.DhcpStaticBinding, error) {

	var (
		resp        *http.Response
		dhcpSuccess manager.DhcpStaticBindingListResult
	)

	if nsxClient == nil {
		return nil, fmt.Errorf("failed recieve dhcp static binding for server")
	}

	if len(SearchVal) == 0 {
		return nil, &ObjectNotFound{"empty search value"}
	}
	if !IsUuid(serverId) {
		return nil, fmt.Errorf("dhcp server must have valid uuid")
	}

	dhcpSuccess, resp, err := nsxClient.ServicesApi.ListDhcpStaticBindings(nsxClient.Context, serverId, nil)
	if err != nil {
		if resp == nil || (resp.StatusCode == 401 || resp.StatusCode == 403) {
			return nil, fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
		}
	} else {
		for _, val := range dhcpSuccess.Results {
			if fn(val, SearchVal) == true {
				return &val, nil
			}
		}
	}

	return nil, fmt.Errorf("failed recieve dhcp static binding for mac %s", SearchVal)
}

/*
  Search DHCP server by server name or id
  TODO refactor
*/
func FindDhcpServer(nsxClient *nsxt.APIClient, searchValue string) (string, error) {

	if nsxClient == nil {
		return "", fmt.Errorf("nsxt client is nil")
	}

	regex := regexp.MustCompile("-")
	matches := regex.FindAllStringIndex(searchValue, -1)

	dhcpServer, httpResponse, err := nsxClient.ServicesApi.ListDhcpServers(nsxClient.Context, nil)
	if err != nil {
		if httpResponse == nil || (httpResponse.StatusCode == 401 || httpResponse.StatusCode == 403) {
			return "", fmt.Errorf("failed recieve dhcp server list: %s", err)
		}
	} else {
		for i := 0; i < len(dhcpServer.Results); i++ {
			if len(matches) != 4 {
				// lookup by name
				if strings.Compare(dhcpServer.Results[i].DisplayName, searchValue) == 0 {
					return dhcpServer.Results[i].Id, nil
				}
			} else {
				// lookup by id
				if strings.Compare(dhcpServer.Results[i].Id, searchValue) == 0 {
					return dhcpServer.Results[i].Id, nil
				}
			}
		}
	}

	return "", fmt.Errorf("dhcp server not found")
}

/*
  Search logical dhcp server based on tag, and return logical dhcp server id
*/
func FindDhcpServerByTag(nsxClient *nsxt.APIClient, tags []common.Tag) (string, error) {

	if nsxClient == nil {
		return "", fmt.Errorf("nsxt client is nil")
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("tag can't empty")
	}

	dhcpServer, httpResponse, err := nsxClient.ServicesApi.ListDhcpServers(nsxClient.Context, nil)
	if err != nil {
		if httpResponse == nil || (httpResponse.StatusCode == 401 || httpResponse.StatusCode == 403) {
			return "", fmt.Errorf("failed recieve dhcp server list: %s", err)
		}
		return "", err
	}

	for i, v := range dhcpServer.Results {
		if reflect.DeepEqual(v.Tags, tags) {
			return dhcpServer.Results[i].Id, nil
		}
	}

	return "", &ObjectNotFound{"dhcp server not found"}
}

// pack everything to a struct so client wont't do mistake with order
type DhcpBindingRequest struct {
	ServerUuid string
	Mac        string
	Ipaddr     string
	Hostname   string
	Gateway    string
	TenantId   string
}

//
func CreateStaticReq(nsxClient *nsxt.APIClient, req *DhcpBindingRequest) (*manager.DhcpStaticBinding, error) {
	return CreateStaticBinding(nsxClient, req.ServerUuid, req.Mac, req.Ipaddr, req.Hostname, req.Gateway, req.TenantId)
}

/*
 Creates a static binding for a DHCP server
  @TODO create method create if need that will check tag and mac address
   if mac address is different than we report about old bindings

*/
func CreateStaticBinding(nsxClient *nsxt.APIClient,
	serverId string,
	macaddr string,
	ipaddr string,
	hostname string,
	gateway string,
	tenantId string) (*manager.DhcpStaticBinding, error) {

	var (
		resp             *http.Response
		dhcpEntrySuccess manager.DhcpStaticBinding
	)

	if len(tenantId) == 0 {
		e := fmt.Errorf("create static binding requst must include tenant id")
		logging.ErrorLogging(e)
		return nil, e
	}

	if !IsUuid(serverId) {
		e := fmt.Errorf("create static binding requst must include dhcp uuid")
		logging.ErrorLogging(e)
		return nil, e
	}
	if net.ParseIP(ipaddr) == nil {
		e := fmt.Errorf("create static binding requst must must include valid ip address")
		logging.ErrorLogging(e)
		return nil, e
	}
	if net.ParseIP(gateway) == nil {
		e := fmt.Errorf("create static binding requst must include valid gateway ip address")
		logging.ErrorLogging(e)
		return nil, e
	}
	_, err := net.ParseMAC(macaddr)
	if err != nil {
		e := fmt.Errorf("create static binding requst must include valid mac address")
		logging.ErrorLogging(e)
		return nil, e
	}

	var newTags = []common.Tag{
		{
			Scope: "jettison-tenant",
			Tag:   tenantId,
		},
	}

	newDhcpBinding := manager.DhcpStaticBinding{}
	newDhcpBinding.MacAddress = macaddr
	newDhcpBinding.IpAddress = ipaddr
	newDhcpBinding.DisplayName = hostname
	newDhcpBinding.HostName = hostname
	newDhcpBinding.GatewayIp = gateway
	newDhcpBinding.Tags = newTags

	dhcpEntrySuccess, resp, err = nsxClient.ServicesApi.CreateDhcpStaticBinding(nsxClient.Context, serverId, newDhcpBinding)
	if err != nil {
		logging.ErrorLogging(err)
		log.Println("error from nsxt-t", err)
		if resp == nil || (resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403) {
			return nil, fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
		}
	}

	return &dhcpEntrySuccess, nil
}

// Function deletes dhcp binding from NSX-T.  The called need provide comparator callback
// that will find a target entry.  For example client can pass callback that will search
// entry based on IP in dhcp static entry.
func DeleteStaticBinding(nsxClient *nsxt.APIClient, serverId string, ipAddr string, macAddr string) error {

	var (
		resp *http.Response
	)

	if len(serverId) == 0 || len(ipAddr) == 0 || len(macAddr) == 0 {
		e := fmt.Errorf("Attribute dhcp server uuid, mac and ip address are mandatory. ")
		logging.ErrorLogging(e)
		return e
	}

	if net.ParseIP(ipAddr) == nil {
		e := fmt.Errorf("invalid IP address")
		logging.ErrorLogging(e)
		return e
	}
	_, err := net.ParseMAC(macAddr)
	if err != nil {
		e := fmt.Errorf("invalid mac address")
		logging.ErrorLogging(e)
		return e
	}

	// lookup dhcp entry by IP
	dhcpBinding, err := GetStaticBinding(nsxClient, serverId, ipAddr,
		func(dhcpEntry manager.DhcpStaticBinding, ipAddr string) bool {
			if dhcpEntry.IpAddress == ipAddr {
				return true
			}
			return false
		})
	if err != nil {
		return fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
	}

	// make sure we have entry for desired controller node and not someone other VM
	if dhcpBinding.MacAddress == macAddr {
		resp, err = nsxClient.ServicesApi.DeleteDhcpStaticBinding(nsxClient.Context, serverId, dhcpBinding.Id)
		if err != nil {
			if resp == nil || (resp.StatusCode == 401 || resp.StatusCode == 403) {
				return fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
			}
		}
	}

	return nil
}

//   Finds a dhcp server that attached to logical switch.
//   In case of error return error, if object not found return ObjectNotFound
//
func FindAttachedDhcpServerProfile(nsxClient *nsxt.APIClient, logicalSwitchID string) (*manager.LogicalDhcpServer, error) {

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	if len(logicalSwitchID) == 0 {
		e := fmt.Errorf("logical switch uuid is a mandatory attribute. ")
		logging.ErrorLogging(e)
		return nil, e
	}

	dhcpServer, httpResponse, err := nsxClient.ServicesApi.ListDhcpServers(nsxClient.Context, nil)
	if err != nil {
		if httpResponse == nil || (httpResponse.StatusCode == 401 || httpResponse.StatusCode == 403) {
			return nil, fmt.Errorf("failed recieve dhcp server list from a nsx-t manager: %s", err)
		}
		return nil, fmt.Errorf("failed recieve dhcp server list from a nsx-t manager: %s", err)
	}
	// loop over all DHCP server and check that we have server attached to a target logical switch
	for i := 0; i < len(dhcpServer.Results); i++ {
		logicalPort, resp, err :=
			nsxClient.LogicalSwitchingApi.GetLogicalPort(nsxClient.Context,
				dhcpServer.Results[i].AttachedLogicalPortId)
		if err != nil {
			if resp == nil || (resp.StatusCode == 401 || resp.StatusCode == 403) {
				return nil, fmt.Errorf("failed lookup logical port from nsx-t manager")
			}
			return nil, fmt.Errorf("failed lookup logical port nsx-t manager")
		}
		if logicalPort.LogicalSwitchId == logicalSwitchID {
			return &dhcpServer.Results[i], nil
		}
	}

	return nil, &ObjectNotFound{"dhcp server profile not found"}
}

//
//   Find a dhcp server with given tag
//
func FindDhcpServerProfileByTag(nsxClient *nsxt.APIClient, tags []common.Tag) (*manager.DhcpProfile, error) {

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	profiles, httpResponse, err := nsxClient.ServicesApi.ListDhcpProfiles(nsxClient.Context, nil)
	if err != nil {
		if httpResponse == nil || (httpResponse.StatusCode == 401 || httpResponse.StatusCode == 403) {
			return nil, fmt.Errorf("failed recieve dhcp server list from a nsx-t manager: %s", err)
		}
		return nil, fmt.Errorf("failed recieve dhcp server list from a nsx-t manager: %s", err)
	}

	for i := 0; i < len(profiles.Results); i++ {
		for k, v := range profiles.Results {
			if reflect.DeepEqual(v.Tags, tags) {
				return &profiles.Results[k], nil
			}
		}
	}

	return nil, &ObjectNotFound{"dhcp server profile not found"}
}

//  Create a new dhcp profile,  profile it create must contain a tag
//  that tag later can be used
func CreateDhcpProfile(nsxClient *nsxt.APIClient,
	clusterId string, profileName string, tags []common.Tag) (string, error) {

	profile := manager.DhcpProfile{
		Description:   "jettison",
		DisplayName:   profileName,
		EdgeClusterId: clusterId,
		Tags:          tags,
	}

	profile, resp, err := nsxClient.ServicesApi.CreateDhcpProfile(nsxClient.Context, profile)
	if err != nil {
		logging.ErrorLogging(err)
		return "", err
	}
	if resp.StatusCode != http.StatusCreated {
		e := fmt.Errorf("nsx-t return unexpected status code for create dhcp profile")
		logging.ErrorLogging(e)
		return "", e
	}

	return profile.Id, nil
}

//  Creates a new dhcp server and attaches to a target edge cluster.
//  Attachment based on dhcp profile.
//
func CreateDhcpServer(nsxClient *nsxt.APIClient,
	req *DhcpServerCreateReq, tags []common.Tag) (*manager.LogicalDhcpServer, error) {

	server := manager.IPv4DhcpServer{
		DhcpServerIp:   req.DhcpServerIp,
		DnsNameservers: req.DnsNameservers,
		DomainName:     req.DomainName,
		GatewayIp:      req.GatewayIp,
	}

	logging.Notification("Attaching dhcp to ", req.LogicalPortID)
	ldsdDhcpServer := manager.LogicalDhcpServer{
		Description:           "jettison",
		DisplayName:           req.ServerName,
		AttachedLogicalPortId: req.LogicalPortID,
		DhcpProfileId:         req.DhcpProfileId,
		Ipv4DhcpServer:        &server,
		Tags:                  tags,
	}

	ldsdDhcpServer, resp, err := nsxClient.ServicesApi.CreateDhcpServer(nsxClient.Context, ldsdDhcpServer)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		e := fmt.Errorf("nsx-t return unexpected status code for create dhcp profile")
		logging.ErrorLogging(e)
		return nil, e
	}

	return &ldsdDhcpServer, nil
}

// Get all DHCP biding based server id and tag attached
// to a binding entry, tagging each binding used to identify
// a a project or tenant, additionally sub-tag can be used to
// identify a segment.
//
func GetStaticBindingByTag(nsxClient *nsxt.APIClient, serverId string,
	tags []common.Tag) ([]*manager.DhcpStaticBinding, error) {

	var (
		resp        *http.Response
		result      []*manager.DhcpStaticBinding
		dhcpSuccess manager.DhcpStaticBindingListResult
	)

	if nsxClient == nil {
		return nil, fmt.Errorf("failed recieve dhcp static binding for server")
	}

	if len(tags) == 0 {
		return nil, &ObjectNotFound{"empty search value"}
	}

	if len(serverId) == 0 {
		return nil, fmt.Errorf("dhcp server must have valid uuid")
	}

	if !IsUuid(serverId) {
		return nil, fmt.Errorf("dhcp server must have valid uuid")
	}

	dhcpSuccess, resp, err := nsxClient.ServicesApi.ListDhcpStaticBindings(nsxClient.Context, serverId, nil)
	if err != nil {
		if resp == nil || (resp.StatusCode == 401 || resp.StatusCode == 403) {
			return nil, fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
		}
		return nil, err
	}

	for i, val := range dhcpSuccess.Results {
		if reflect.DeepEqual(val.Tags, tags) {
			result = append(result, &dhcpSuccess.Results[i])
		}
	}

	if len(result) == 0 {
		return nil, &ObjectNotFound{}
	}

	return result, nil
}

//
// Delete dhcp server
//
func DeleteDhcpServer(nsxClient *nsxt.APIClient, serverUuid string) (string, string, error) {

	// read server
	dhcpServer, resp, err := nsxClient.ServicesApi.ReadDhcpServer(nsxClient.Context, serverUuid)
	if err != nil {
		e := fmt.Errorf("nsx-t failed read dhcp server %s", serverUuid)
		logging.ErrorLogging(e)
		return "", "", e
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code for read dhcp server %s", serverUuid)
		logging.ErrorLogging(e)
		return "", "", e
	}

	profileId := dhcpServer.DhcpProfileId
	PortId := dhcpServer.AttachedLogicalPortId

	// detatch port
	opt := map[string]interface{}{
		"detach": true,
	}
	resp, err = nsxClient.LogicalSwitchingApi.DeleteLogicalPort(nsxClient.Context, PortId, opt)
	if err != nil {
		e := fmt.Errorf("nsx-t delete dhcp port %s", serverUuid)
		logging.ErrorLogging(err)
		return "", "", e
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code during delete operation %s", serverUuid)
		logging.ErrorLogging(e)
		return "", "", e
	}

	// delete dhcp server
	resp, err = nsxClient.ServicesApi.DeleteDhcpServer(nsxClient.Context, serverUuid)
	if err != nil {
		e := fmt.Errorf("nsx-t delete dhcp profile %s", serverUuid)
		logging.ErrorLogging(err)
		return "", "", e
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code during delete operation %s", serverUuid)
		logging.ErrorLogging(e)
		return "", "", e
	}

	return profileId, PortId, nil

}

func DeleteDhcpProfile(nsxClient *nsxt.APIClient, profileUuid string) (bool, string, error) {

	dhcpProfile, resp, err := nsxClient.ServicesApi.ReadDhcpProfile(nsxClient.Context, profileUuid)
	if err != nil {
		e := fmt.Errorf("nsx-t failed read dhcp profile %s", profileUuid)
		logging.ErrorLogging(e)
		return true, "", e
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code for read dhcp profile %s", profileUuid)
		logging.ErrorLogging(e)
		return true, "", e
	}

	resp, err = nsxClient.ServicesApi.DeleteDhcpProfile(nsxClient.Context, profileUuid)
	if err != nil {
		e := fmt.Errorf("nsx-t delete dhcp profile %s", profileUuid)
		logging.ErrorLogging(err)
		return true, "", e
	}
	if resp.StatusCode != http.StatusOK {
		e := fmt.Errorf("nsx-t return unexpected status code during delete operation %s", profileUuid)
		logging.ErrorLogging(e)
		return true, "", e
	}

	return true, dhcpProfile.EdgeClusterId, nil
}
