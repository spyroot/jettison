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

Jettison main config parser.
 Reads configuration and serialize everything to appConfig struct.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package vcenter

import (
	"context"
	"flag"
	"fmt"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/osutil"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
	"log"
	"net/url"
)

const (
	envURL      = "GOVMOMI_URL"
	envUserName = "GOVMOMI_USERNAME"
	envPassword = "GOVMOMI_PASSWORD"
	envInsecure = "GOVMOMI_INSECURE"
)

var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", envURL)
var urlFlag = flag.String("url", osutil.GetEnvString(envURL, ""), urlDescription)

//  Function open up connection to vCenter or ESXi.
//
func Connect(ctx context.Context, hostname string, username string, password string) (*govmomi.Client, error) {

	flag.Parse()

	var err error
	var u *url.URL
	if hostname != "" {
		u, err = soap.ParseURL(hostname)
		if err != nil {
			return nil, err
		}
	} else {
		u, err = soap.ParseURL(*urlFlag)
		if err != nil {
			return nil, err
		}
	}

	if username != "" && password != "" {
		u.User = url.UserPassword(username, password)
	}

	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		logging.CriticalMessage("Failed to connect to vmware vim")
		return nil, err
	}

	log.Print("Connected to vCenter manager and return context")
	return client, nil
}
