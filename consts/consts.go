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

All Jettison shared constants by entire project

Author spyroot
mbaraymov@vmware.com
*/

package consts

type ExecDependency struct {
	Path    string
	Present bool
}

// map store all exec that need to be in the system
// before we run a program we need check that each exist and if not
// report to a user
var ExecDependencyMap = map[string][]ExecDependency{
	// shell
	"sentinel": {
		{"sentinel", false},
	},
	"bash": {
		{"/bin/bash", false},
		{"/usr/bin/bash", false},
		{"/usr/local/bin/bash", false},
	},
	"sh": {
		{"/bin/sh", false},
	},
	"cfssl": {
		{"/usr/local/bin/cfssl", false},
	},
	"cfssljson": {
		{"/usr/local/bin/cfssljson", false},
	},
	"sshpass": {
		{"/usr/local/bin/sshpass", false},
		{"/usr/bin/sshpass", false},
	},
	"ssh-copy-id": {
		{"/usr/bin/ssh-copy-id", false},
		{"/usr/local/bin/ssh-copy-id", false},
	},
	"openssl": {
		{"/usr/local/bin/openssl", false},
		{"/usr/bin/openssl", false},
	},
	"ansible-playbook": {
		{"/usr/local/bin/ansible-playbook", false},
	},
}

var TenantDirs = []string{
	"ca",
	"api",
	"admin",
	"workers",
	"kube-proxy",
	"kubernetes",
}

const (
	DefaultExpire string = "8760h"

	InitCertCa = "cfssl gencert -initca %s | cfssljson -bare ca"

	AdminCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
						-config=%s -profile=kubernetes %s | cfssljson -bare admin`

	WorkerCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
			        	-config=%s -hostname=%s,%s -profile=kubernetes %s | cfssljson -bare %s`

	ProxyCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
						-config=%s -profile=kubernetes %s | cfssljson -bare kube-proxy`

	ApiCertGen = `cfssl gencert -ca=../ca/ca.pem -ca-key=../ca/ca-key.pem \
			        	-config=%s -hostname=%s -profile=kubernetes %s | cfssljson -bare kubernetes`

	// ansible path to initial ca request json file
	AnsibleKeyCaJson = "cacertRequest"
	// path to api request json file
	AnsibleJsonApiReq = "apiRequest"
	// ansible ca config variable
	AnsibleCaConfigJson = "caConfig"
	//
	AnsibleJsonRequest = "adminRequest"
	//
	AnsibleJsonProxyReq = "proxyRequest"
	//
	AnsibleTenantHome = "tenanthome"

	K8SAdminPrivateKey = "K8sAdminPem"
	K8SAdminCert       = "k8sAdminCsr"
	K8SAdminPem        = "k8sAdminPem"

	// default key names
	DefaultK8SAdminCsr        = "admin.csr"
	DefaultK8SAdminPem        = "admin.pem"
	DefaultK8SAdminPrivateKey = "admin-key.pem"

	DefaultCaCertificate = "ca.csr"
	DefaultCaKey         = "ca-key.pem"

	DefaultAnsibleLocation = "/.ansible/"

	DefaultAnsibleHostFile = "hosts"
	DefaultAnsibleTempFile = "hosts.tmp"

	DefaultPublicKey = ".ssh/id_rsa"

	DefaultAnsibleUsername = "vmware"
	DefaultAnsiblePort     = 22
	DefaultAnsibleProtocol = "ssh"

	DefaultAnsibleTenantDir = "tenant"

	DefaultHostsVarsPath = "host_vars"
	DefaultGroupVarsPath = "group_vars"
	DefaultGroupVarFile  = "all"

	PlaybookTemplate = "playbook-template.yml"
)

// return a map of all assets where key is relative path
func AllAssets() map[string]string {
	return assets
}

// returns asset
func Asset(name string) string {
	return assets[name]
}

// returns tree serialize as array of dirs
// that jettison need deploy in
// ansible project home dir
func AssetsDirs() []string {
	return dirs
}
