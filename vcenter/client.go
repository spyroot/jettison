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

//var insecureDescription = fmt.Sprintf("Don't verify the server's certificate chain [%s]", envInsecure)
//var insecureFlag = flag.Bool("insecure", osutil.GetEnvBool(envInsecure, false), insecureDescription)

/*

 */
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

	// Connect and log in to ESX or vCenter
	log.Print("Connected to vCenter manager and return context")
	return govmomi.NewClient(ctx, u, true)
}
