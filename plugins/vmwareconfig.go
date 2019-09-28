package main

import (
	"bufio"
	"fmt"

	"github.com/spyroot/jettison/logging"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spyroot/jettison/nsxtapi"
	"github.com/vmware/go-vmware-nsxt/manager"
)

const (
	NsxtConfigFile    = "../config.yml"
	DefaultConfigPath = "/usr/local/etc/jettison"
)

type Config struct {
	Hostname      string `yaml:"hostname"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	LogicalSwitch string `yaml:"logicalSwitch"`
	EdgeCluster   string `yaml:"edgeCluster"`
	OverlayTzName string `yaml:"overlayTransport"`
}

// Edge Cluster field can be name or UUID, in case client passed a name internally jettison keep note uuid.
type NsxtConfig struct {
	NsxtConfig Config `yaml:"nsxt"`

	overlayTzId   string // overlay transport zone
	edgeClusterId string // edge cluster id
	dhcpServerID  string

	tier0 map[string]*manager.LogicalRouter // all tier0 attached to cluster - populated during discovery
	tier1 map[string]*manager.LogicalRouter // all tier1 attached to cluster - populate during  discovery
}

func (n *NsxtConfig) Hostname() string {
	if n != nil {
		return n.NsxtConfig.Hostname
	}
	return ""
}

func (n *NsxtConfig) Username() string {
	if n != nil {
		return n.NsxtConfig.Username
	}
	return ""
}

func (n *NsxtConfig) Password() string {
	if n != nil {
		return n.NsxtConfig.Password
	}
	return ""
}

func (n *NsxtConfig) OverlayTransportName() string {
	if n != nil {
		return n.NsxtConfig.OverlayTzName
	}
	return ""
}

func (n *NsxtConfig) EdgeCluster() string {
	if n != nil {
		return n.NsxtConfig.EdgeCluster
	}
	return ""
}

func (n *NsxtConfig) SetEdgeCluster(s string) {
	if n != nil {
		n.NsxtConfig.EdgeCluster = s
	}
}

func (n *NsxtConfig) LogicalSwitch() string {
	if n != nil {
		return n.NsxtConfig.LogicalSwitch
	}
	return ""
}

// set target overlay transport zone uuid
func (n *NsxtConfig) SetOverlayTzId(uuid string) {
	n.overlayTzId = uuid
}

// return transport zone that plugin will use
func (n *NsxtConfig) OverlayTransportUuid() string {
	if n != nil {
		return n.overlayTzId
	}
	return ""
}

//
func (n *NsxtConfig) SetOverlayTzName(name string) {
	n.NsxtConfig.OverlayTzName = name
}

//
func (n *NsxtConfig) SetEdgeClusterUuid(uuid string) {
	n.edgeClusterId = uuid
}

//
func (n *NsxtConfig) EdgeClusterUuid() string {
	if n != nil {
		return n.edgeClusterId
	}
	return ""
}

//
func (n *NsxtConfig) SetDhcpUuid(s string) {
	if n != nil {
		n.dhcpServerID = s
	}
}

func (n *NsxtConfig) DhcpServerUuid() string {
	if n != nil {
		return n.dhcpServerID
	}

	return ""
}

func (n *NsxtConfig) GetActiveTierZero() (string, error) {

	if n != nil {
		for _, v := range n.tier0 {
			v.HighAvailabilityMode = nsxtapi.HaActiveStandby
			return v.Id, nil
		}
	}

	return "", fmt.Errorf("tier zero not found")
}

func (n *NsxtConfig) TierZero() map[string]*manager.LogicalRouter {
	if n != nil {
		return n.tier0
	}
	return map[string]*manager.LogicalRouter{}
}

func (n *NsxtConfig) TierOne() map[string]*manager.LogicalRouter {
	if n != nil {
		return n.tier1
	}
	return map[string]*manager.LogicalRouter{}
}

// Add tier zero to a list
func (n *NsxtConfig) AddTierZero(t0 *manager.LogicalRouter) {
	if n != nil {
		if n.tier0 == nil {
			n.tier0 = make(map[string]*manager.LogicalRouter, 0)
		}
		n.tier0[t0.Id] = t0
	}
}

// Add tier tier one to a list
func (n *NsxtConfig) AddTierOne(t1 *manager.LogicalRouter) {
	if n != nil {
		if n.tier1 == nil {
			n.tier1 = make(map[string]*manager.LogicalRouter, 0)
		}
		n.tier1[t1.Id] = t1
	}
}

func validate(nsxt *NsxtConfig) (bool, error) {

	if nsxt.NsxtConfig.Hostname == "" {
		return false, fmt.Errorf("missing NSX-T Manager hostname")
	}
	if nsxt.NsxtConfig.Username == "" {
		return false, fmt.Errorf("missing NSX-T Manager username")
	}
	if nsxt.NsxtConfig.Password == "" {
		return false, fmt.Errorf("missing NSX-T Manager password")
	}
	if nsxt.NsxtConfig.OverlayTzName == "" {
		return false, fmt.Errorf("missing NSX-T Manager overlay transport zone")
	}

	return true, nil
}

//
// Creates a new NsxtConfig from configuration file.
//
func NewNsxtConfig() (*NsxtConfig, error) {
	file, r, err := ReadFromFile()
	if err != nil {
		return nil, fmt.Errorf("failed to read default location configuration")
	}
	defer file.Close()

	nsxtConfig, err := ReadConfig(r)
	if err != nil {
		return nil, fmt.Errorf("failed parse default configuration %v", err)
	}

	return nsxtConfig, nil
}

//
// Read a nsx-t configuration file to buffer io return open file and reader.
// caller must close a file
func ReadFromFile() (*os.File, *bufio.Reader, error) {

	pwd, _ := os.Getwd()

	p := filepath.Join(pwd, NsxtConfigFile)
	log.Println("Reading config ", p)
	file, err := os.Open(p)
	if err == nil {
		// return buffer caller need close file
		return nil, bufio.NewReader(file), nil
	}

	p = filepath.Join(pwd, "./config.yml")
	log.Println("Reading config ", p)
	file, err = os.Open(p)
	if err == nil {
		return nil, bufio.NewReader(file), nil
	}

	log.Println("Reading default config ", DefaultConfigPath+NsxtConfigFile)
	p = filepath.Join(DefaultConfigPath, NsxtConfigFile)
	file, err = os.Open(p)
	if err != nil {
		logging.CriticalMessage("failed to read from default location")
		return nil, nil, fmt.Errorf("failed to read from default location")
	}

	return file, bufio.NewReader(file), nil
}

//
//  Reads yaml configuration file and parse it and returns
//  NsxtConfig, it does internal validation for mandatory
//  fields.
//
func ReadConfig(reader io.Reader) (*NsxtConfig, error) {

	if reader == nil {
		return nil, fmt.Errorf("nil reader")
	}

	nsxConfig := &NsxtConfig{}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed read nsx-t configuration")
	}

	err = yaml.Unmarshal(data, nsxConfig)
	if err != nil {
		return nil, err
	}

	var r bool
	r, err = validate(nsxConfig)
	if r == false {
		return nil, err
	}

	if nsxConfig.NsxtConfig.Password == "*" {
		for ok := true; ok; ok = !(len(nsxConfig.NsxtConfig.Password) > 1) {
			fmt.Print("NSX-T password: ")
			_, err = fmt.Scanln(nsxConfig.NsxtConfig.Password)
			if err != nil {
				ok = false
			}
		}
	}

	return nsxConfig, nil
}
