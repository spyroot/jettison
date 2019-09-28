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
	"github.com/spyroot/jettison/logging"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"regexp"
	"strings"
)

func IsUuid(s string) bool {
	regex := regexp.MustCompile("-")
	matches := regex.FindAllStringIndex(s, -1)
	if len(matches) == 4 {
		return true
	}

	return false
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
		logging.ErrorLogging(err)
		return nsxt.APIClient{}, fmt.Errorf("error creating NSX-T API client: %v", err)
	}

	log.Print("Connected to nsx-t manager and return context")
	return *nsxtClient, nil
}

/**
  Return manager.TransportZone as slice in case someone decided name tz with same name.
*/
func FindTransportZone(nsxClient *nsxt.APIClient, zoneName string) ([]*manager.TransportZone, error) {

	var result []*manager.TransportZone

	if nsxClient == nil {
		return nil, fmt.Errorf("nsxt client is nil")
	}

	tz, _, err := nsxClient.NetworkTransportApi.ListTransportZones(nsxClient.Context, nil)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, fmt.Errorf("failed lookup switch id")
	}

	for i, v := range tz.Results {
		s := strings.TrimSpace(v.DisplayName)
		if s == zoneName {
			result = append(result, &tz.Results[i])
		}
	}

	return result, nil
}
