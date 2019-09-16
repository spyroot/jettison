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

Supplemental tool that main purpose generate certs by leveraging external
tool or ca.

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package certsutil

import (
	"fmt"
	"github.com/spyroot/jettison/consts"
	"github.com/spyroot/jettison/internal"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/osutil"
	"log"
	"net"
	"path"
	"strings"
)

/** K8S default hosts in cert allow path */
var additionalHost = []string{
	"127.0.0.1",
	"localhost",
	"kubernetes",
	"kubernetes.default",
	"kubernetes.default.svc",
	"kubernetes.default.svc.cluster",
	"kubernetes.default.svc.cluster.local",
}

/*
RBAC Node Permissions
In 1.6, the system:node cluster role was automatically bound to
the system:nodes group when using the RBAC Authorization mode.

In 1.7, the automatic binding of the system:nodes group to the
system:node role is deprecated because the node authorizer accomplishes the same purpose
with the benefit of additional restrictions on secret and configmap access.
If the Node and RBAC authorization modes are both enabled,
In 1.8, the binding will not be created at all.
When using RBAC, the system:node cluster role will continue to be created,
for compatibility with deployment methods that bind other users or groups to that role
*/

// source
//https://kubernetes.io/docs/reference/access-authn-authz/rbac/

/* Default ClusterRole */

const DefaultNodeRol = "system:node"
const DefaultkKbeScheRole = "system:kube-scheduler"
const DefaultkKbeVolRole = "system:volume-scheduler"
const DefaultKubeCtrlRol = "system:kube-controller-manager"
const DefaultNodeProxylRol = "system:node-proxier"
const DefaultClusterAdminRole = "cluster-admin"
const DefaultAdminRole = "admin"

/* bindings */
const DefaultClusterBind = "system:masters"

const DefaultNodeBind = "system:nodes"

const DefaultNodeProxylBind = "system:kube-proxy"

//Default ClusterRoleBinding
//
//system:masters group
//system:kube-scheduler user
//system:kube-scheduler user
//system:kube-controller-manager user

/**
  Function create a list of host as comma separated list
*/
func hostAsList(certClient []CertClient) string {

	var list []string

	// add all hostname to a list
	for _, v := range certClient {
		list = append(list, v.GetHostname())
		list = append(list, v.GetIpAddress())
	}

	// add all defaults to a list
	for _, v := range additionalHost {
		list = append(list, v)
	}
	// return comma as  comma list
	return strings.Join(list[:], ",")
}

/**
  Create generate ca config file and initialize CA certificate
  File will be stored in specified in request
*/
func MakeCaConfig(certClient *CertRequest) (string, error) {

	jsonConfig := JsonCaConfig{}
	jsonConfig.Signing.Default.Expiry = consts.DefaultExpire
	// that default */
	jsonConfig.Signing.Profiles.Kubernetes.Usages = []string{
		"signing",
		"key encipherment",
		"server auth",
		"client auth",
	}
	jsonConfig.Signing.Profiles.Kubernetes.Expiry = consts.DefaultExpire
	jsonReq := certClient.tenant + ".ca-config.json"

	dir := path.Join(certClient.path, "/", certClient.tenant, "/ca/")
	jsonFile := path.Join(dir, jsonReq)
	err := jsonConfig.WriteToFile(jsonFile)
	if err != nil {
		logging.ErrorLogging(err)
		return "", err
	}

	return jsonFile, nil
}

/**

 */
func MakeCaCertificateReq(certClient *CertRequest) (CertificateFileRespond, error) {

	certRequest := NewCertRequest()
	certRequest.CN = "Kubernetes"
	certRequest.Names[0].O = "Kubernetes"

	jsonFileName := certClient.tenant + "." + "ca-csr.json"
	dir := path.Join(certClient.path, "/", certClient.tenant, "/ca/")
	jsonFile := path.Join(dir, jsonFileName)

	err := certRequest.WriteToFile(jsonFile)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	log.Println("Generating ca certificate for tenant", certRequest.tenant)

	respond, err := InitCa(certClient, dir)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	return respond, nil
}

func lookupCNBinding(role string) string {

	if role == "proxy" {
		return "system:kube-proxy"
	} else if role == "admin" {
		return "admin"
	} else if role == "api" {
		return "kubernetes"
	} else if role == "controller" {
		return "system:kube-controller-manager"
	} else if role == "scheduler" {
		return "system:kube-scheduler"
	} else if role == "service-accounts" {
		return "service-accounts"
	}
	return ""
}

func lookupRole(role string) string {

	if role == "admin" {
		return "system:masters"
	} else if role == "worker" {
		return "system:nodes"
	} else if role == "proxy" {
		return "system:node-proxier"
	} else if role == "api" {
		return "Kubernetes"
	} else if role == "controler" {
		return "system:kube-controller-manager"
	} else if role == "scheduler" {
		return "system:kube-scheduler"
	} else if role == "service-accounts" {
		return "Kubernetes"
	}
	return ""
}

/**
  Generate kubernetes admin certs
*/
func MakeKubeCertificateReq(certClient *CertRequest,
	caCert, caKey, caConfig, kubeRole, hostnames string) (CertificateFileRespond, error) {

	certRequest := NewCertRequest()
	certRequest.CN = lookupCNBinding(kubeRole)
	certRequest.Names[0].O = lookupRole(kubeRole)
	certRequest.Names[0].OU = "Kubernetes"

	jsonFileName := certClient.tenant + "." + kubeRole + "-csr.json"
	dir := path.Join(certClient.path, "/", certClient.tenant, "/"+kubeRole+"/")
	jsonFile := path.Join(dir, jsonFileName)

	err := certRequest.WriteToFile(jsonFile)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	req := map[string]string{
		"cfssl":     certClient.cfssl,
		"cfssljson": certClient.cfssljson,
		"cacert":    caCert,
		"cakey":     caKey,
		"config":    caConfig,
		"profile":   "kubernetes",
		"jsonfile":  jsonFile,
		"bare":      kubeRole,
	}

	if len(hostnames) > 0 {
		req["hostname"] = hostnames
	}

	certResp, err := MakeCert(certClient, req, dir)
	if err != nil {
		return nil, err
	}

	return certResp, nil
}

/**
  Generate kubernetes admin certs
*/
func MakeNodeCertificateReq(certClient *CertRequest,
	caCert, caKey, caConfig, hostname, hostnames string) (CertificateFileRespond, error) {

	certRequest := NewCertRequest()
	certRequest.CN = DefaultNodeRol + hostname
	certRequest.Names[0].O = DefaultClusterBind
	certRequest.Names[0].OU = "Kubernetes"

	jsonFileName := certClient.tenant + "." + hostname + "-csr.json"
	dir := path.Join(certClient.path, "/", certClient.tenant, "/workers/")
	jsonFile := path.Join(dir, jsonFileName)
	err := certRequest.WriteToFile(jsonFile)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	req := map[string]string{
		"cfssl":     certClient.cfssl,
		"cfssljson": certClient.cfssljson,
		"cacert":    caCert,
		"cakey":     caKey,
		"config":    caConfig,
		"profile":   "kubernetes",
		"jsonfile":  jsonFile,
		"bare":      hostname,
		"hostname":  hostnames,
	}

	certResp, err := MakeCert(certClient, req, dir)
	if err != nil {
		return nil, err
	}

	return certResp, nil
}

func verify(openssl, cacert, clientcert string) (bool, error) {

	cmd := fmt.Sprintf("%s verify -verbose -CAfile %s %s", openssl, cacert, clientcert)
	out, err := osutil.ChangeAndExecWithSdout("/", cmd)
	if err != nil {
		logging.ErrorLogging(err)
		return false, err
	}
	return strings.Contains(out, "OK"), nil
}

/*
 *  Function generate certs
 */
func GenerateTenantCerts(certClient []CertClient, path, tenant, serviceCidr string) (map[string]string, error) {

	var (
		cfsslLoc     string
		cfssljsonLoc string
		openssl      string
		cacert       string
		cakey        string
	)

	if len(path) == 0 || len(tenant) == 0 {
		return nil, fmt.Errorf("invalid path or tenant name")
	}

	cfsslLoc, err := internal.GetExecLocation("cfssl")
	if err != nil {
		return nil, err
	}
	cfssljsonLoc, err = internal.GetExecLocation("cfssljson")
	if err != nil {
		return nil, err
	}
	openssl, err = internal.GetExecLocation("openssl")
	if err != nil {
		return nil, err
	}

	var certFiles = make(map[string]string, 0)
	certRequest := &CertRequest{}
	certRequest.path = path
	certRequest.tenant = tenant
	certRequest.cfssl = cfsslLoc
	certRequest.cfssljson = cfssljsonLoc
	// make ca config and certs
	config, err := MakeCaConfig(certRequest)
	if err != nil {
		return nil, err
	}
	caResp, err := MakeCaCertificateReq(certRequest)
	if err != nil {
		return nil, err
	}
	certFiles["cacert"] = caResp.getCertificate()
	cacert = caResp.getCertificate()
	certFiles["cakey"] = caResp.getKey()
	cakey = caResp.getKey()
	certFiles["config"] = config

	caResp, err = MakeKubeCertificateReq(certRequest,
		cacert, cakey, config, "admin", "")
	if err != nil {
		return nil, err
	}
	if ok, err := verify(openssl, cacert, caResp.getCertificate()); ok && err == nil {
		log.Println("Generated certificates for admin : verified")
	}

	certFiles["admin.cert"] = caResp.getCertificate()
	certFiles["admin.key"] = caResp.getKey()

	// workers nodes
	for _, v := range certClient {
		if ok, err := verify(openssl, cacert, caResp.getCertificate()); ok && err == nil {
			log.Println("Generated certificates for worker node", v.GetHostname(), ": verified")
		}
		hostnames := v.GetHostname() + "," + v.GetIpAddress()
		caResp, err = MakeNodeCertificateReq(certRequest,
			cacert, cakey, config, v.GetHostname(), hostnames)
		if err != nil {
			return nil, err
		}
		certFiles[v.GetHostname()+".cert"] = caResp.getCertificate()
		certFiles[v.GetHostname()+".akey"] = caResp.getKey()
	}

	// create proxy cert
	caResp, err = MakeKubeCertificateReq(certRequest,
		cacert, cakey, config, "proxy", "")
	if err != nil {
		return nil, err
	}
	if ok, err := verify(openssl, cacert, caResp.getCertificate()); ok && err == nil {
		log.Println("Generated certificates for kupe-proxy : verified")
	}

	// create API cert
	addr, subnet, err := net.ParseCIDR(serviceCidr)
	if err != nil {
		return nil, err
	}
	// add entire service cidr to api certificate
	for i := 0; subnet.Contains(addr); i++ {
		additionalHost = append(additionalHost, addr.String())
		addr = internal.NextIP(addr, 1)
	}
	caResp, err = MakeKubeCertificateReq(certRequest,
		cacert, cakey, config, "api", hostAsList(certClient))
	if err != nil {
		return nil, err
	}
	if ok, err := verify(openssl, cacert, caResp.getCertificate()); ok && err == nil {
		log.Println("Generated certificates for kupe-api : verified")
	}

	return certFiles, nil
}
