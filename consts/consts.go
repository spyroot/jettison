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

All Jettison consts should be here

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package consts

// map store all exec that need to be in the system
// before we run a program we need check that each exist and if not
// report to a user
var ExecDependency = map[string]bool{
	// shell
	"/bin/bash": false,
	"/bin/sh":   false,

	// cfssl util
	"/usr/local/bin/cfssl":     false,
	"/usr/local/bin/cfssljson": false,
	"cfssl":                    false,
	"cfssljson":                false,

	// default for ssh pass and ssh copy id
	"/usr/local/bin/sshpass": false,
	"/usr/bin/ssh-copy-id":   false,
}

var TenantDirs = []string{
	"ca",
	"api",
	"admin",
	"proxy",
	"workers",
}

const (
	DefaultExpire string = "8760h"

	InitCertCa = "cfssl gencert -initca %s | cfssljson -bare ca"

	AdminCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
						-config=%s -profile=kubernetes %s | cfssljson -bare admin`

	WorkerCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
			        	-config=%s -hostname=%s,%s  -profile=kubernetes %s | cfssljson -bare %s`

	ProxyCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
						-config=%s -profile=kubernetes %s | cfssljson -bare kube-proxy`

	ApiCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
			        	-config=%s -hostname=%s -profile=kubernetes %s | cfssljson -bare kubernetes`

	// ansible path to initial ca request json file
	AnsibleKeyCaJson = "cacertRequest"
	//
	AnsibleJsonApiReq = "apiRequest"
	// ansible ca config variable
	AnsibleCaConfigJson = "caConfig"
	//
	AnsibleJsonRequest = "adminRequest"
	//
	AnsibleJsonProxyReq = "proxyRequest"
)
