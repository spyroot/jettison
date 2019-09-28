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
	"encoding/json"
	"fmt"

	"log"
	"os"
	"path"
	"strings"

	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/osutil"
)

/*
  Interface that client need implement in order cert gen be able generate certificates
*/
type CertClient interface {
	// ip address of host bounded to a cert
	GetIpAddress() string
	// hostname of host bounded to a cert
	GetHostname() string
	// a host type
	GetNodeType() jettypes.NodeType
}

type CertificateFileRespond interface {
	getCertificateReq() string
	getCertificate() string
	getConfig() string
	getKey() string
}

// Generic Cert Request
type CertRespond struct {
	// full path to csr file  ca.csr
	cacsr string

	// full path to key file ca.pem
	cakey string

	// full path to key file ca.pem
	cacert string

	// full path to key file ca.pem

	caconfig string
}

func (c *CertRespond) getCertificateReq() string {
	if c != nil {
		return c.cacsr
	}
	return ""
}

func (c *CertRespond) getCertificate() string {
	if c != nil {
		return c.cacert
	}
	return ""
}

func (c *CertRespond) getKey() string {
	if c != nil {
		return c.cakey
	}
	return ""
}

func (c *CertRespond) getConfig() string {
	if c != nil {
		return c.caconfig
	}
	return ""
}

/**
  json ca config file request
*/
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

/**
  Writes a a json config to a file
*/
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
	path      string
	tenant    string
	cfssl     string
	cfssljson string

	CN  string `json:"CN"`
	Key struct {
		Algo string `json:"algo"`
		Size int    `json:"size"`
	} `json:"key"`
	Names []CertRequestNames `json:"names"`
}

/*
 *  Default constructor for ssl cert
 */
func NewCertRequest() *CertRequest {

	newCert := &CertRequest{}
	newCert.CN = "admin"
	newCert.Key.Algo = "rsa"
	newCert.Key.Size = 2048

	newCert.Names = []CertRequestNames{
		{
			C:  "Auto",
			L:  "Auto",
			O:  "system:masters",
			OU: "VMware",
			ST: "VMware",
		},
	}
	return newCert
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

	_, err = inputFile.Write(p)
	if err != nil {
		return err
	}

	return nil
}

/* string format printer */
func Tprintf(format string, params map[string]string) string {

	for key, val := range params {
		format = strings.Replace(format, "%{"+key+"}s", fmt.Sprintf("%s", val), -1)
	}

	return format
}

// list of mandatory attrs for makeCert
var certRequestFields = []string{"cacert", "cakey", "config",
	"profile", "jsonfile", "bare", "cfssljson", "cfssl"}

/**
  Initialize all CA certificate and key
*/
func InitCa(certReq *CertRequest, dir string) (CertificateFileRespond, error) {

	jsonFile := path.Join(dir, "/", certReq.tenant+".ca-csr.json")
	if !osutil.CheckIfExist(jsonFile) {
		err := fmt.Errorf("failed create certificates no ca config present %s", jsonFile)
		logging.ErrorLogging(err)
		return nil, err
	}

	cmd := fmt.Sprintf("%s gencert -initca %s | %s -bare ca",
		certReq.cfssl, jsonFile, certReq.cfssljson)

	err := osutil.ChangeAndExec(dir, cmd)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	csr := path.Join(dir, "/", "ca.csr")
	cert := path.Join(dir, "/", "ca.pem")
	privatekey := path.Join(dir, "/", "ca-key.pem")

	if !osutil.CheckIfExist(csr) {
		err := fmt.Errorf("failed create certificates")
		logging.ErrorLogging(err)
		return nil, err
	}
	if !osutil.CheckIfExist(cert) {
		err := fmt.Errorf("failed create certificates")
		logging.ErrorLogging(err)
		return nil, err
	}
	if !osutil.CheckIfExist(privatekey) {
		err := fmt.Errorf("failed create certificates")
		logging.ErrorLogging(err)
		return nil, err
	}

	resp := &CertRespond{
		csr,
		privatekey,
		cert,
		jsonFile,
	}

	return resp, nil
}

/**

 */
func MakeCert(args map[string]string, dir string) (*CertRespond, error) {

	// check all mandatory fields
	for _, field := range certRequestFields {
		if _, ok := args[field]; !ok {
			err := fmt.Errorf("%s mandatory key", field)
			logging.ErrorLogging(err)
			return nil, err
		}
	}

	var s string
	if _, ok := args["hostname"]; ok {
		s = "%{cfssl}s gencert -ca=%{cacert}s -ca-key=%{cakey}s " +
			"-config=%{config}s -hostname=%{hostname}s -profile=%{profile}s " +
			"%{jsonfile}s | %{cfssljson}s -bare %{bare}s"
	} else {
		s = "%{cfssl}s gencert -ca=%{cacert}s -ca-key=%{cakey}s " +
			"-config=%{config}s -profile=%{profile}s " +
			"%{jsonfile}s | %{cfssljson}s -bare %{bare}s"
	}

	// build cmd
	cmd := Tprintf(s, args)

	// execute cfssl tool
	err := osutil.ChangeAndExec(dir, cmd)
	if err != nil {
		logging.ErrorLogging(err)
		return nil, err
	}

	bare, _ := args["bare"]
	csr := path.Join(dir, "/", bare+".csr")
	cert := path.Join(dir, "/", bare+".pem")
	privatekey := path.Join(dir, "/", bare+"-key.pem")

	if !osutil.CheckIfExist(csr) {
		err := fmt.Errorf("failed create certificates")
		logging.ErrorLogging(err)
		return nil, err
	}
	if !osutil.CheckIfExist(cert) {
		err := fmt.Errorf("failed create certificates")
		logging.ErrorLogging(err)
		return nil, err
	}
	if !osutil.CheckIfExist(privatekey) {
		err := fmt.Errorf("failed create certificates")
		logging.ErrorLogging(err)
		return nil, err
	}

	// return all paths to all files
	return &CertRespond{
		csr,
		privatekey,
		cert,
		args["config"],
	}, nil
}
