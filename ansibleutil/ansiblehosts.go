package ansibleutil

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spyroot/jettison/consts"
	"github.com/spyroot/jettison/logging"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type AnsibleInventoryHost struct {
	AnsibleConnection string `yaml:"ansible_connection,omitempty"`
	AnsibleHost       string `yaml:"ansible_host,omitempty"`
	AnsiblePort       int    `yaml:"ansible_port,omitempty"`
	AnsibleUser       string `yaml:"ansible_user,omitempty"`
}

//`yaml:"a,omitempty"`

type AnsibleGroups struct {
	Hosts map[string]*AnsibleInventoryHost `yaml:"hosts,omitempty"`
}

type AnsibleCluster struct {
	Members map[string]*AnsibleGroups `yaml:"children,omitempty"`
}

type AnsibleInventoryHosts struct {
	filePath string

	All struct {
		Cluster map[string]*AnsibleCluster `yaml:"children,omitempty"`
	} `yaml:"all,omitempty"`

	isEmpty bool
}

/**
  Function adds a ansible host to inventory.  It doesn't serialize it file.
  Caller must do it explicitly.
*/
func (a *AnsibleInventoryHosts) AddSlaveHost(ansibleHost *AnsibleHosts, project string) error {

	if len(ansibleHost.Hostname) == 0 {
		return fmt.Errorf("ansible host empty")
	}

	h := AnsibleInventoryHost{}

	if len(ansibleHost.Connection) == 0 {
		h.AnsibleConnection = consts.DefaultAnsibleProtocol
	} else {
		h.AnsibleConnection = ansibleHost.Connection
	}

	if len(ansibleHost.User) == 0 {
		h.AnsibleUser = consts.DefaultAnsibleUsername
	} else {
		h.AnsibleUser = ansibleHost.User
	}

	if ansibleHost.Port >= 65535 || ansibleHost.Port <= 0 {
		h.AnsiblePort = consts.DefaultAnsiblePort
	} else {
		h.AnsiblePort = ansibleHost.Port
	}

	h.AnsibleHost = ansibleHost.Hostname

	cluster := a.addClusterIfNeed(project)
	group := cluster.addGroupIfNeed(ansibleHost.Group)
	group.addHostIfNeed(ansibleHost.Name, &h)

	return nil
}

func (a *AnsibleInventoryHosts) DeleteProject(projectName string) bool {

	if a == nil {
		return false
	}
	if a.All.Cluster == nil || len(projectName) == 0 {
		return false
	}

	_, ok := a.All.Cluster[projectName]
	if ok {
		delete(a.All.Cluster, projectName)
		if len(a.All.Cluster) == 0 {
			a.isEmpty = true
		}
	}

	return ok
}

func (a *AnsibleInventoryHosts) DeleteSaveHost(hostname string) bool {

	if a == nil {
		return false
	}
	if a.All.Cluster == nil || len(hostname) == 0 {
		return false
	}

	for pkey, c := range a.All.Cluster {
		for ckey, g := range c.Members {
			for key := range g.Hosts {
				if key == hostname {

					// delete host
					delete(g.Hosts, key)

					// group if it last entry
					if len(g.Hosts) == 0 {
						delete(c.Members, ckey)
					}

					// project itself if no more group left
					if len(c.Members) == 0 {
						delete(a.All.Cluster, pkey)
					}

					if len(a.All.Cluster) == 0 {
						a.All.Cluster = nil
						a.isEmpty = true
					}
					return true
				}
			}
		}
	}
	return false
}

/*
   Function search in all project a given host.
*/
func (a *AnsibleInventoryHosts) FindSaveHost(hostname string) (*AnsibleInventoryHost, bool) {

	if a == nil {
		return nil, false
	}
	if a.All.Cluster == nil || len(hostname) == 0 {
		return nil, false
	}

	for _, v := range a.All.Cluster {
		for _, v := range v.Members {
			for k, v := range v.Hosts {
				if k == hostname {
					return v, true
				}
			}
		}
	}

	return nil, false
}

/**
  Function adds a new host to a given ansible group, if a host with given key already
  in a group it will replace it with new one.
*/
func (group *AnsibleGroups) addHostIfNeed(hostname string, host *AnsibleInventoryHost) {

	if group.Hosts == nil {
		group.Hosts = map[string]*AnsibleInventoryHost{}
		group.Hosts[hostname] = host
	}

	// if we found we update
	if _, ok := group.Hosts[hostname]; ok {
		group.Hosts[hostname] = host
		return
	}

	group.Hosts[hostname] = host
}

/**

 */
func (cluster *AnsibleCluster) addGroupIfNeed(name string) *AnsibleGroups {

	if cluster.Members == nil {
		cluster.Members = map[string]*AnsibleGroups{}
		group := &AnsibleGroups{}
		cluster.Members[name] = group
		return group
	}

	if group, ok := cluster.Members[name]; ok {
		return group
	}

	group := &AnsibleGroups{}
	cluster.Members[name] = group

	return group
}

/**

 */
func (a *AnsibleInventoryHosts) addClusterIfNeed(name string) *AnsibleCluster {

	if a.All.Cluster == nil {
		a.All.Cluster = map[string]*AnsibleCluster{}
		cluster := &AnsibleCluster{}
		a.All.Cluster[name] = cluster
		return cluster
	}

	if v, ok := a.All.Cluster[name]; ok {
		return v
	}

	cluster := &AnsibleCluster{}
	a.All.Cluster[name] = cluster

	return cluster
}

func isSafe(filePath string) bool {

	// path to host need store either in /tmp or /usr/local/etc/ansible or /etc/ansible etc
	if strings.Contains(filePath, "/tmp") ||
		strings.Contains(filePath, "/usr/local/etc/ansible") ||
		strings.Contains(filePath, "/etc/ansible") ||
		strings.Contains(filePath, "/.ansible/") {
		return true
	}

	return false
}

/**
  Function serialize itself to a ansible file in yaml format to a file.
*/
func (a *AnsibleInventoryHosts) WriteToFile() error {

	if a == nil {
		return nil
	}

	if len(a.filePath) == 0 {
		return nil
	}

	if isSafe(a.filePath) == false {
		logging.ErrorLogging(fmt.Errorf("unsafe path"))
		return nil
	}

	// if object is empty set we delete old file since no more projects left
	if a.isEmpty {
		tmpFile := path.Join(a.filePath, "/", consts.DefaultAnsibleTempFile)
		oldFile, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			nerr := fmt.Errorf("failed create a %s file: %s", tmpFile, err)
			logging.ErrorLogging(nerr)
			return nerr
		}

		oldFile.Close()
		newFile := path.Join(a.filePath, "/", consts.DefaultAnsibleHostFile)
		err = os.Rename(tmpFile, newFile)

		return nil
	}

	// nil not ok but file can be empty / test that.
	if a.All.Cluster == nil {
		return nil
	}

	// create dir if need
	if _, err := os.Stat(a.filePath); os.IsNotExist(err) {
		log.Println("Creating dir", a.filePath)
		err := os.MkdirAll(a.filePath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed create %s directory for ansible inventory file error %s ", a.filePath, err)
		}
	}

	// create a new file
	tmpFile := path.Join(a.filePath, "/", consts.DefaultAnsibleTempFile)
	inputFile, err := os.OpenFile(tmpFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed create a %s file: %s", tmpFile, err)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", tmpFile, err)
		}
	}()

	var yamlBuffer []byte
	yamlBuffer, err = yaml.Marshal(a)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed marshal %s yaml file file: %s", tmpFile, err)
	}

	_, err = inputFile.Write(yamlBuffer)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed marshal %s yaml file file: %s", tmpFile, err)
	}

	// rename that will overwrite old file
	newFile := path.Join(a.filePath, "/", consts.DefaultAnsibleHostFile)
	err = os.Rename(tmpFile, newFile)
	if err != nil {
		logging.ErrorLogging(err)
		return fmt.Errorf("failed opening file %s", err)
	}

	return nil
}

/**
  Function Reads existing yaml file
*/
func CreateFromInventory(filePath string) (*AnsibleInventoryHosts, error) {

	if len(filePath) == 0 {
		logging.ErrorLogging(fmt.Errorf("empty ansible path to a file"))
		return nil, errors.Wrap(fmt.Errorf("empty ansible path to a file"), "")
	}

	var ansibleHosts AnsibleInventoryHosts
	ansibleHosts.filePath = filePath

	file := path.Join(filePath, "/", consts.DefaultAnsibleHostFile)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		// check if we can create file / if we can return empty inventory
		inputFile, err2 := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)
		if err2 != nil {
			logging.ErrorLogging(err)
			return nil, fmt.Errorf("failed create a %s file: %s", file, err2)
		}
		defer func() {
			if err := inputFile.Close(); err != nil {
				log.Println("failed to close", file, err)
			}
		}()

		return &ansibleHosts, nil
	}

	// if we failed we will return empty inventory
	err = yaml.Unmarshal(data, &ansibleHosts)
	if err != nil {
		logging.ErrorLogging(err)
		return &ansibleHosts, err
	}

	return &ansibleHosts, nil
}
