package ansibleutil

import (
	"encoding/json"
	"fmt"
	"github.com/spyroot/jettison/consts"
	"github.com/spyroot/jettison/logging"
	"github.com/spyroot/jettison/osutil"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
)

const (
	DefaultBecomePassword = "Vmware1!"
	DefaultCni            = "v0.8.2"
	DefaultEtcd           = "v3.4.0"
	DefaultCri            = "1.2.4"
)

// Ansible variable related to kubernetes
//
type AnsibleGlobalVars struct {
	HomePath   string
	Project    string `yaml:"project"`
	TenantHome string `yaml:"tenanthome"`

	Owner string `yaml:"k8sowner"`

	CniVersion  string `yaml:"cniversion"`
	EtcdVersion string `yaml:"etcdversion"`
	CriVersion  string `yaml:"criversion"`

	Downloads string `yaml:"downloads"`

	IngressIP       string   `yaml:"k8shaproxyip"`
	IngressHostname string   `yaml:"k8shaproxyhost"`
	MasterNode      []string `yaml:"masternode"`

	// cluster specific data
	ClusterDns string `yaml:"clusterdns"`
	ClusterIp  string `yaml:"clusterip"`

	// k8s cluster related data
	ServiceCidr    string `yaml:"servicenet"`
	ClusterCidr    string `yaml:"clustercidr"`
	EncyrptionKey  string `yaml:"encyrptionkey"`
	BecomePassword string
}

/*
 *  Default constructor for ssl cert
 */
func NewAnsibleGlobalVars() AnsibleGlobalVars {

	v := AnsibleGlobalVars{}
	v.CniVersion = DefaultCni
	v.EtcdVersion = DefaultEtcd
	v.CriVersion = DefaultCri
	v.BecomePassword = DefaultBecomePassword

	v.Downloads = path.Join(osutil.UserHomeDir(),
		consts.DefaultAnsibleLocation, "/templates/downloads")

	return v
}

//
//Function serialize itself to a ansible group var
//
func (a *AnsibleGlobalVars) WriteToFileJson() error {

	dir := path.Join(a.HomePath, "/", consts.DefaultGroupVarsPath)
	// create dir for groups_var
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed create %s", dir)
	}

	// create a new file
	file := path.Join(a.HomePath, "/", consts.DefaultGroupVarsPath, "/", consts.DefaultGroupVarFile)
	inputFile, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed create a file: %s", file)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", file, err)
		}
	}()

	var out []byte
	out, err = json.MarshalIndent(a, "", "\t")
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed marshal yaml file file: %s", err)
	}

	_, err = inputFile.Write(out)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}

	return nil
}

//Function serialize itself to a ansible group var
//
func (a *AnsibleGlobalVars) WriteToFile() error {

	dir := path.Join(a.HomePath, "/", consts.DefaultGroupVarsPath)
	// create dir for groups_var
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed create %s", dir)
	}

	// create a new file
	file := path.Join(a.HomePath, "/", consts.DefaultGroupVarsPath, "/", consts.DefaultGroupVarFile)
	inputFile, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed create a file: %s", file)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", file, err)
		}
	}()

	var yamlBuffer []byte
	yamlBuffer, err = yaml.Marshal(a)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed marshal yaml file file: %s", err)
	}

	_, err = inputFile.Write(yamlBuffer)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}

	return nil
}

//Function serialize itself to a ansible group var for a project.
func (a *AnsibleGlobalVars) WriteVars(vars map[string]string) error {

	dir := path.Join(a.HomePath, "/", consts.DefaultGroupVarsPath)
	// create dir for groups_var
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed create %s", dir)
	}

	// if we want add additional vars we open for append only
	file := path.Join(a.HomePath, "/", consts.DefaultGroupVarsPath, "/", consts.DefaultGroupVarFile)
	inputFile, err := os.OpenFile(file, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed create a file: %s", file)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", file, err)
		}
	}()

	for k, v := range vars {
		_, _ = inputFile.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}

	return nil
}
