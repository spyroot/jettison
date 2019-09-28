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

NSX-T API integration. Wrapper around NSX-T switching API

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package nsxtapi

import (
	"fmt"
	"github.com/spyroot/jettison/logging"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"net/http"
	"reflect"
)

/*
  A generic error use to differentiate error types
*/
type ObjectNotFound struct {
	objectName string
}

func (e *ObjectNotFound) Error() string {
	return fmt.Sprintf("object not found %v:", e.objectName)
}

/**
  Search a logical switches that conforms tags field
*/
func FindLogicalSwitchByTag(nsxClient *nsxt.APIClient, tags []common.Tag) (*manager.LogicalSwitch, error) {

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	if len(tags) == 0 {
		return nil, fmt.Errorf("empty tag")
	}

	// if client passed name we search buy name we find first that match.
	switchList, resp, err := nsxClient.LogicalSwitchingApi.ListLogicalSwitches(nsxClient.Context, nil)
	if err != nil {
		return nil, fmt.Errorf("failed lookup switch id")
	}
	if resp.StatusCode != http.StatusOK {
		log.Println(" nsx-t returned none standard http code")
	}
	for i, logicalSwitch := range switchList.Results {
		if reflect.DeepEqual(logicalSwitch.Tags, tags) {
			return &switchList.Results[i], nil
		}
	}

	return nil, &ObjectNotFound{"logical switch not found"}
}

/*
 Function searches a logical switch and returns manager.LogicalSwitch object
 function handle a case when caller provide a switch name instead UUID
 note in NSX-T two switch might have same name
*/
func FindLogicalSwitch(nsxClient *nsxt.APIClient,
	logicalSwitchID string, tags []*common.Tag) (*manager.LogicalSwitch, error) {

	var (
		switchID = logicalSwitchID
	)

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	// if client passed name we search by name we find first that match
	// and return
	if !IsUuid(logicalSwitchID) {
		switchList, resp, err := nsxClient.LogicalSwitchingApi.ListLogicalSwitches(nsxClient.Context, nil)
		if err != nil {
			return nil, fmt.Errorf("failed lookup switch id")
		}
		if resp.StatusCode != http.StatusOK {
			log.Println(" nsx-t returned none standard http code")
		}
		for i, logicalSwitch := range switchList.Results {
			if logicalSwitch.DisplayName == logicalSwitchID {
				switchID = logicalSwitch.Id
				return &switchList.Results[i], nil
			}
		}
		return nil, &ObjectNotFound{objectName: logicalSwitchID}
	}

	logicalSwitch, resp, err := nsxClient.LogicalSwitchingApi.GetLogicalSwitch(nsxClient.Context, switchID)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return &logicalSwitch, nil
	}

	return nil, &ObjectNotFound{objectName: logicalSwitchID}
}

/*
   Function Create a logical switch in  NSX-T,  a transport zone is mandatory argument
   and required by NSX-T.  It accept transport zone as name or UUID.
   Generally transport zone might have duplicate name and caller
   need resolve name conflict.
*/
func CreateLogicalSwitch(nsxClient *nsxt.APIClient,
	zoneName string, switchName string, tags []common.Tag) (*manager.LogicalSwitch, error) {

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	if len(zoneName) == 0 || len(switchName) == 0 {
		err := fmt.Errorf("transport zone name or switch name is empty")
		logging.ErrorLogging(err)
		return nil, err
	}

	var tzId string
	if !IsUuid(zoneName) {
		zones, err := FindTransportZone(nsxClient, zoneName)
		if err != nil {
			logging.ErrorLogging(err)
		}
		if len(zones) > 1 {
			return nil, fmt.Errorf("two zone with same name. Please remove duplictes transport zones")
		}
		if len(zones) == 0 {
			return nil, fmt.Errorf("transport zone not found")
		}
		tzId = zones[0].Id
	} else {
		tzId = zoneName
	}

	logicalSwitch := manager.LogicalSwitch{
		Description:     "created by jettison tool",
		DisplayName:     switchName,
		Tags:            tags,
		AdminState:      "UP",
		TransportZoneId: tzId,
		ReplicationMode: "MTEP",
	}

	logicalSwitch, resp, err := nsxClient.LogicalSwitchingApi.CreateLogicalSwitch(nsxClient.Context, logicalSwitch)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, fmt.Errorf("error during LogicalSwitch create: %v", err)
	}

	// technically it should be 201 only
	// https://www.vmware.com/support/nsxt/doc/nsxt_20_api.html#Sections.Logical%20Switching.Logical%20Switches
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return &logicalSwitch, nil
	}

	return &logicalSwitch, nil
}

//
//  Searches a port by given switchID and logical port name
//  Because we are searching by name we might have more than one port
//  name with same name.
//
//  In Jettison name can't be the same since each name identified by uuid
//
func FindLogicalPort(nsxClient *nsxt.APIClient, switchId string, tags []common.Tag) ([]*manager.LogicalPort, error) {

	var result []*manager.LogicalPort
	filter := map[string]interface{}{
		"logicalSwitchId": switchId,
	}
	ports, _, err := nsxClient.LogicalSwitchingApi.ListLogicalPorts(nsxClient.Context, filter)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	// we check all ports that belong to a tenant
	for i, port := range ports.Results {
		if port.LogicalSwitchId == switchId {
			if reflect.DeepEqual(port.Tags, tags) {
				result = append(result, &ports.Results[i])
			}
		}
	}

	if len(result) == 0 {
		return nil, &ObjectNotFound{" logical port not found " + switchId}
	}

	return result, nil
}

/*
  Creates a new logical port
*/
func CreateLogicalPort(nsxClient *nsxt.APIClient, portName string, switchId string, tags []common.Tag) (*manager.LogicalPort, error) {

	if len(portName) == 0 || len(switchId) == 0 {
		return nil, fmt.Errorf("empty arguments")
	}

	logicalSwitch := manager.LogicalPort{
		Description:     "jettison",
		DisplayName:     portName,
		AdminState:      "UP",
		Tags:            tags,
		Attachment:      nil,
		LogicalSwitchId: switchId,
	}

	lds, _, err := nsxClient.LogicalSwitchingApi.GetLogicalSwitch(nsxClient.Context, switchId)
	if err != nil {
		e := fmt.Errorf("failed get switch details %s %v", switchId, err)
		logging.ErrorLogging(e)
		return nil, e
	}
	// check switch state
	if lds.AdminState == "UP" {
		port, resp, err := nsxClient.LogicalSwitchingApi.CreateLogicalPort(nsxClient.Context, logicalSwitch)
		if err != nil {
			logging.ErrorLogging(err)
			return nil, fmt.Errorf("error during LogicalSwitch create: %v", err)
		}
		// technically it should be 200 only
		// https://www.vmware.com/support/nsxt/doc/nsxt_20_api.html#Sections.Logical%20Switching.Logical%20Switches
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
			return &port, nil
		}
		err = fmt.Errorf("nsx return not expected status code %v", resp.StatusCode)
		logging.ErrorLogging(err)
		return nil, err
	}

	return nil, fmt.Errorf("cant create switch port switch not ready")
}

/*
  Detach and delete a logical port
*/
func DeleteLogicalPort(nsxClient *nsxt.APIClient, portId string) (bool, error) {

	opt := map[string]interface{}{
		"detach": true,
	}
	resp, err := nsxClient.LogicalSwitchingApi.DeleteLogicalPort(nsxClient.Context, portId, opt)
	if err != nil {
		return false, fmt.Errorf("failed delete a logical port : %v", err)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return true, nil
	}

	return false, nil
}

/**

 */
func DeleteLogicalSwitch(nsxClient *nsxt.APIClient, switchUuid string) (bool, error) {

	if !IsUuid(switchUuid) {
		return false, fmt.Errorf("switchId must be a valid uuid")
	}

	//Cascade

	opt := map[string]interface{}{
		"cascade": true,
		"force":   true,
	}

	resp, err := nsxClient.LogicalSwitchingApi.DeleteLogicalSwitch(nsxClient.Context, switchUuid, opt)
	if err != nil {
		return false, fmt.Errorf("failed delete a switch : %v", err)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return true, nil
	}

	return false, nil
}
