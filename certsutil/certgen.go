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

Supplementally tool that main purpose generate certs

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package certsutil

import (
	"encoding/json"
	"fmt"
	"github.com/spyroot/jettison/config"
	"github.com/spyroot/jettison/consts"
	"github.com/spyroot/jettison/internal"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

type JsonCaConfig struct {
	Signing struct {
		Default struct {
			Expiry string `json:"expiry"`
		} `json:"default"`
		Profiles struct {
			Kubernetes struct {
				Usages []string `json:"usages"`
				Expiry string   `json:"expiry"`
			} `json:"kubernetes"`
		} `json:"profiles"`
	} `json:"signing"`
}

var additionalHost = []string{
	"172.16.88.100",
	"127.0.0.1",
	"localhost",
	"kubernetes.default",
}

func (node *JsonCaConfig) WriteToFile(filePath string) error {

	inputFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", filePath, err)
		}
	}()

	var p []byte
	p, err = json.MarshalIndent(node, "", "\t")
	if err != nil {
		return err
	}

	log.Println(" Writing cert", filePath, "file.")
	_, err = inputFile.Write(p)
	if err != nil {
		return err
	}

	return nil
}

// Generic Cert names
type CertRequestNames struct {
	C  string `json:"C"`
	L  string `json:"L"`
	O  string `json:"O"`
	OU string `json:"OU"`
	ST string `json:"ST"`
}

// Generic Cert Request
type CertRequest struct {
	CN  string `json:"CN", default:"admin"`
	Key struct {
		Algo string `json:"algo"`
		Size int    `json:"size"`
	} `json:"key"`
	Names []CertRequestNames `json:"names"`
}

/*
 *  Default constructor for ssl cert
 */
func NewCertRequest() CertRequest {
	something := CertRequest{}
	something.CN = "admin"
	something.Key.Algo = "rsa"
	something.Key.Size = 2048

	something.Names = []CertRequestNames{CertRequestNames{
		C:  "Auto",
		L:  "Auto",
		O:  "system:masters",
		OU: "VMware",
		ST: "VMware",
	}}

	return something
}

// serialize cert as json
func (node *CertRequest) PrintAsJson() {
	var p []byte
	p, err := json.MarshalIndent(node, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%s \n", p)
}

// serialize cert to a file
func (node *CertRequest) WriteToFile(filePath string) error {

	inputFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", filePath, err)
		}
	}()

	var p []byte
	p, err = json.MarshalIndent(node, "", "\t")
	if err != nil {
		return err
	}

	log.Println(" Writing cert", filePath, "file.")
	_, err = inputFile.Write(p)
	if err != nil {
		return err
	}

	return nil
}

/**
  Function iterate over all map that store path to each shell
  and if we find one set bool true
*/
func CheckIfExist() {

	// iterate over all shell path and set flags
	for k, _ := range consts.ExecDependency {
		_, err := exec.LookPath(k)
		if err != nil {
			consts.ExecDependency[k] = false
		}
	}
}

/**

 */
func buildTenantDirs(homeDir string, projectName string) error {

	// create home dir
	err := os.MkdirAll(homeDir, os.ModePerm)
	if err != nil {
		log.Println("failed create tenant project dir")
	}

	tenantDir := path.Join(homeDir, "/", projectName)
	err = os.MkdirAll(tenantDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed create tenant project dir %s", err)
	}

	for _, dir := range consts.TenantDirs {
		newDir := path.Join(tenantDir, "/", dir)
		err = os.MkdirAll(newDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed create tenant project dir %s", err)
		}
	}

	return nil
}

/**
  Function take dir that called need to take before calling exec
*/
func changeAndExec(dir string, cmd string) error {

	log.Println("Executing in dir", dir)

	err := os.Chdir(dir)
	if err != nil {
		return err
	}

	log.Println(cmd)
	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return err
	}

	return nil
}

/**
  Function create a list of host as comma separated list
*/
func hostAsList(nodes []*config.NodeTemplate) string {

	var list []string
	for _, v := range nodes {
		list = append(list, v.IPv4AddrStr)
	}

	for _, v := range nodes {
		additionalHost = append(additionalHost, v.IPv4AddrStr)
	}

	return strings.Join(list[:], ",")
}

/*
 *  Function generate certs
 */
func GenerateTenantCerts(vim *internal.Vim, nodes []*config.NodeTemplate, vars map[string]string) error {

	var (
		name    = vim.AppConfig.GetDeploymentName()
		homeDir = vim.AppConfig.GetAnsible().AnsibleTemplates
	)

	err := buildTenantDirs(homeDir, name)
	if err != nil {
		return err
	}

	jsonCaConfig := JsonCaConfig{}
	jsonCaConfig.Signing.Default.Expiry = consts.DefaultExpire
	jsonCaConfig.Signing.Profiles.Kubernetes.Usages = []string{
		"signing",
		"key encipherment",
		"server auth",
		"client auth",
	}

	// generate ca config
	jsonCaConfig.Signing.Profiles.Kubernetes.Expiry = consts.DefaultExpire
	caConfig := name + ".ca-config.json"
	dir := path.Join(homeDir, "/", name, "/ca/")
	configPath := path.Join(dir, caConfig)
	err = jsonCaConfig.WriteToFile(configPath)
	if err != nil {
		return err
	}
	vars[consts.AnsibleCaConfigJson] = configPath

	// ca cert
	{
		certRequest := NewCertRequest()
		jsonReqFile := name + ".ca-csr.json"
		certRequest.Names[0].O = "Kubernetes"
		dir := path.Join(homeDir, "/", name, "/ca/")
		jsonReqPath := path.Join(dir, jsonReqFile)
		// write a cert request to file
		err := certRequest.WriteToFile(jsonReqPath)
		if err != nil {
			return err
		}
		// generate cert
		vars[consts.AnsibleKeyCaJson] = jsonReqPath
		cmd := fmt.Sprintf(consts.InitCertCa, jsonReqPath)
		err = changeAndExec(dir, cmd)
		if err != nil {
			return err
		}
	}

	{
		// admin cert
		adminCertRequest := NewCertRequest()
		adminCertRequest.Names[0].O = "system:masters"
		adminCertRequest.Names[0].OU = "Kubernetes"
		certRequest := name + ".admin-csr.json"
		dir = path.Join(homeDir, "/", name, "/admin/")
		jsonReqPath := path.Join(dir, certRequest)
		err := adminCertRequest.WriteToFile(jsonReqPath)
		if err != nil {
			return err
		}
		vars[consts.AnsibleJsonRequest] = jsonReqPath
		err = os.Chdir(dir)
		if err != nil {
			return err
		}
		cmd := fmt.Sprintf(consts.AdminCertGen, vars[consts.AnsibleCaConfigJson], certRequest)
		err = changeAndExec(dir, cmd)
		if err != nil {
			return err
		}
	}

	for _, node := range nodes {
		if node.Type == config.WorkerType {
			workerCert := NewCertRequest()
			workerCert.CN = "system:node:" + node.IPv4AddrStr
			workerCert.Names[0].O = "system:nodes"
			workerCert.Names[0].OU = "Kubernetes"

			fileName := node.IPv4AddrStr + "-csr.json"
			dir = path.Join(homeDir, "/", name, "/workers/")
			fullPath := path.Join(dir, fileName)
			err := workerCert.WriteToFile(fullPath)
			if err != nil {
				return err
			}
			cmd := fmt.Sprintf(consts.WorkerCertGen, vars[consts.AnsibleCaConfigJson],
				node.IPv4AddrStr, node.IPv4AddrStr, fileName, node.IPv4AddrStr)
			err = changeAndExec(dir, cmd)
			if err != nil {
				return err
			}
		}
	}

	// proxy cert
	{
		proxyCertRequest := NewCertRequest()
		proxyCertRequest.CN = "system:kube-proxy"
		proxyCertRequest.Names[0].O = "system:node-proxier"
		proxyCertRequest.Names[0].OU = "Kubernetes"
		proxyFileName := name + ".kube-proxy-csr.json"
		dir = path.Join(homeDir, "/", name, "/proxy/")
		proxyCertPath := path.Join(dir, proxyFileName)
		err := proxyCertRequest.WriteToFile(proxyCertPath)
		if err != nil {
			return err
		}
		vars[consts.AnsibleJsonProxyReq] = proxyCertPath
		cmd := fmt.Sprintf(consts.ProxyCertGen, vars[consts.AnsibleCaConfigJson], proxyFileName)
		err = changeAndExec(dir, cmd)
		if err != nil {
			return err
		}
	}

	// api cert
	{
		apiCertRequest := NewCertRequest()
		apiCertRequest.CN = "kubernetes"
		apiCertRequest.Names[0].O = "Kubernetes"
		apiCertRequest.Names[0].OU = "Kubernetes"
		apiFileName := name + ".kubernetes-csr.json"
		dir = path.Join(homeDir, "/", name, "/api/")
		apiCertPath := path.Join(dir, apiFileName)
		err := apiCertRequest.WriteToFile(apiCertPath)
		if err != nil {
			return err
		}
		vars[consts.AnsibleJsonApiReq] = apiCertPath
		cmd := fmt.Sprintf(consts.ApiCertGen, vars[consts.AnsibleCaConfigJson], hostAsList(nodes), apiFileName)
		err = changeAndExec(dir, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}
