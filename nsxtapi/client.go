/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

NSX-T API Intergration

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package nsxtapi

import (
	"errors"
	"fmt"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Comparator func(manager.DhcpStaticBinding, string) bool

var DhcpLookupMap = map[string]func(dhcpEntry manager.DhcpStaticBinding, val string) bool{
	"mac": func(dhcpEntry manager.DhcpStaticBinding, macAddr string) bool {
		if dhcpEntry.MacAddress == macAddr {
			return true
		}
		return false
	},
	"ip": func(dhcpEntry manager.DhcpStaticBinding, ipAddr string) bool {
		if dhcpEntry.IpAddress == ipAddr {
			return true
		}
		return false
	},
	"ipmac-pair": func(dhcpEntry manager.DhcpStaticBinding, ipAddr string) bool {
		if dhcpEntry.IpAddress == ipAddr {
			return true
		}
		return false
	},
}

// Open NSX connection and return nsxtapi.APIClient context.
func Connect(managerHost string, user string, password string) (nsxt.APIClient, error) {

	if managerHost == "" {
		return nsxt.APIClient{}, errors.New("missing NSX-T manager host")
	}

	if user == "" {
		return nsxt.APIClient{}, errors.New("missing NSX-T username")
	}

	if password == "" {
		return nsxt.APIClient{}, errors.New("missing NSX-T password")
	}

	nsxtClient, err := nsxt.NewAPIClient(&nsxt.Configuration{
		BasePath: fmt.Sprintf("https://%s/api/v1", managerHost),
		UserName: user,
		Password: password,
		Host:     managerHost,
		Insecure: true,
		RetriesConfiguration: nsxt.ClientRetriesConfiguration{
			MaxRetries:    1,
			RetryMinDelay: 100,
			RetryMaxDelay: 500,
		},
	})

	if err != nil {
		return *nsxtClient, fmt.Errorf("error creating NSX-T API client: %s", err)
	}

	log.Print("Connected to nsx-t manager and return context")
	return *nsxtClient, nil
}

// Get static binding from a DHCP
func GetStaticBinding(nsxClient *nsxt.APIClient,
	serverId string, SearchVal string, fn Comparator) (manager.DhcpStaticBinding, error) {

	var (
		resp             *http.Response
		dhcpSuccess      manager.DhcpStaticBindingListResult
		dhcpEntrySuccess manager.DhcpStaticBinding
	)

	if nsxClient == nil {
		return dhcpEntrySuccess, fmt.Errorf("failed recieve dhcp static binding for server")
	}

	dhcpSuccess, resp, err := nsxClient.ServicesApi.ListDhcpStaticBindings(nsxClient.Context, serverId, nil)
	if err != nil {
		if resp == nil || (resp.StatusCode == 401 || resp.StatusCode == 403) {
			return dhcpEntrySuccess, fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
		}
	} else {
		for _, val := range dhcpSuccess.Results {
			if fn(val, SearchVal) == true {
				return val, nil
			}
		}
	}

	return dhcpEntrySuccess, fmt.Errorf("failed recieve dhcp static binding for mac %s", SearchVal)
}

// Search a DHCP server and return a DHCP server ID
func FindDhcpServer(nsxClient *nsxt.APIClient, serverName string) (string, error) {

	if nsxClient == nil {
		return "", fmt.Errorf("nsxt client is nil")
	}

	regex := regexp.MustCompile("-")
	matches := regex.FindAllStringIndex(serverName, -1)

	dhcpServer, httpResponse, err := nsxClient.ServicesApi.ListDhcpServers(nsxClient.Context, nil)
	if err != nil {
		if httpResponse == nil || (httpResponse.StatusCode == 401 || httpResponse.StatusCode == 403) {
			return "", fmt.Errorf("failed recieve dhcp server list: %s", err)
		}
	} else {
		for i := 0; i < len(dhcpServer.Results); i++ {
			if len(matches) != 4 {
				// lookup by name
				if strings.Compare(dhcpServer.Results[i].DisplayName, serverName) == 0 {
					return dhcpServer.Results[i].Id, nil
				}
			} else {
				// lookup by id
				if strings.Compare(dhcpServer.Results[i].Id, serverName) == 0 {
					return dhcpServer.Results[i].Id, nil
				}
			}
		}
	}

	return "", fmt.Errorf("dhcp server not found")
}

/*
CreateStaticBinding Create a static binding for a DHCP server
*@param ctx context.Context Authentication Context
@param serverId
@param dhcpStaticBinding
@return manager.DhcpStaticBinding */
func CreateStaticBinding(nsxClient *nsxt.APIClient,
	serverId string,
	macaddr string,
	ipaddr string,
	hostname string,
	gateway string) (manager.DhcpStaticBinding, error) {

	var (
		resp             *http.Response
		dhcpEntrySuccess manager.DhcpStaticBinding
	)

	newDhcpBinding := manager.DhcpStaticBinding{}
	newDhcpBinding.MacAddress = macaddr
	newDhcpBinding.IpAddress = ipaddr
	newDhcpBinding.DisplayName = hostname
	newDhcpBinding.HostName = hostname
	newDhcpBinding.GatewayIp = gateway

	dhcpEntrySuccess, resp, err := nsxClient.ServicesApi.CreateDhcpStaticBinding(nsxClient.Context, serverId, newDhcpBinding)
	if err != nil {
		log.Println("error from nsxt-t", err)
		if resp == nil || (resp.StatusCode == 400 || resp.StatusCode == 401 || resp.StatusCode == 403) {
			return dhcpEntrySuccess, fmt.Errorf("failed recieve dhcp static binding for server %s: %s", serverId, err)
		}
	} else {
		log.Println("got succsess for", newDhcpBinding.MacAddress, newDhcpBinding.IpAddress)

	}

	return dhcpEntrySuccess, nil
}

// Function deletes dhcp binding from NSX-T.  The called need provide comparator callback
// that will find a target entry.  For example client can pass callback that will search
// entry based on IP in dhcp static entry.
func DeleteStaticBinding(nsxClient *nsxt.APIClient, serverId string, ipAddr string, macAddr string) error {

	var (
		resp *http.Response
	)

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

// Function finds a DHCP server that attached to logical switch
func FindDhcpServerProfile(nsxClient *nsxt.APIClient, logicalSwitchID string) (*manager.LogicalDhcpServer, error) {

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	dhcpServer, httpResponse, err := nsxClient.ServicesApi.ListDhcpServers(nsxClient.Context, nil)
	if err != nil {
		if httpResponse == nil || (httpResponse.StatusCode == 401 || httpResponse.StatusCode == 403) {
			return nil, fmt.Errorf("failed recieve dhcp server list from nsx-t manager: %s", err)
		}
		return nil, fmt.Errorf("failed recieve dhcp server list from nsx-t manager: %s", err)
	}
	// loop over all DHCP server and check that we have server attached to a target logical switch
	for i := 0; i < len(dhcpServer.Results); i++ {
		logicalPort, resp, err :=
			nsxClient.LogicalSwitchingApi.GetLogicalPort(
				nsxClient.Context,
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

	return nil, fmt.Errorf("dhcp server not found in %s", logicalSwitchID)
}

// Function finds a logical switch and returns manager.LogicalSwitch object
// function handle a case when caller provide a switch name instead UUID
// note in NSX-T two switch might have same name
func FindLogicalSwitch(nsxClient *nsxt.APIClient, logicalSwitchID string) (*manager.LogicalSwitch, error) {

	var (
		switchID string = logicalSwitchID
	)

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	regex := regexp.MustCompile("-")
	matches := regex.FindAllStringIndex(switchID, -1)

	// if client passed name instead switch ID
	if len(matches) != 4 {
		switchList, _, err := nsxClient.LogicalSwitchingApi.ListLogicalSwitches(nsxClient.Context, nil)
		if err != nil {
			return nil, fmt.Errorf("failed lookup switch id")
		}

		for _, logicalSwitch := range switchList.Results {
			if logicalSwitch.DisplayName == logicalSwitchID {
				switchID = logicalSwitch.Id
				break
			}
		}
	}

	ret, _, err := nsxClient.LogicalSwitchingApi.GetLogicalSwitch(nsxClient.Context, switchID)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
