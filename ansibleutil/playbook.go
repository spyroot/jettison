package ansibleutil

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Task struct {
	Debug string `yaml:"debug"`
}

// Ansible playbook mapping to a structure
// It min set that playbook needs contains
type PlaybookSection struct {
	Hosts       string            `yaml:"hosts"`
	Connection  string            `yaml:"connection,omitempty"`
	Roles       []string          `yaml:"roles,omitempty"`
	Tasks       []Task            `yaml:"tasks,omitempty"`
	GatherFacts string            `yaml:"gather_facts,omitempty"`
	Become      bool              `yaml:"become,omitempty"`
	Vars        map[string]string `yaml:"vars,omitempty"`
}

// A playbook holds all entire playbook, each section is an array element
// the main purpose of object is to generate playbooks a the run time
// and be able add or remove variable from playbook section.  For example set of variables
// that required for a role ,  or generate roles and attach to a playbook
type Playbook struct {
	playbooks []*PlaybookSection `yaml:""`

	file     string
	template string
}

// Creates a playbook from io.reader.  A reader can be anything that implements
// io reader interface.  String , Buffer , file etc
func MakeNewPlaybook(reader io.Reader) (*Playbook, error) {
	p, err := makeTemplate(reader)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func MakeNewFromFile(file string) (*Playbook, error) {
	p, err := readFromFile(file)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Each ansible hosts section of playbook
// might contain comma separated template key
// that we need replace to a new key
func (p *Playbook) Transform(templateKey string, newKey string) {

	for i := range p.playbooks {
		s := strings.Split(p.playbooks[i].Hosts, ",")
		for j, key := range s {
			if strings.Contains(strings.ToLower(key), strings.ToLower(templateKey)) {
				s[j] = newKey
				p.playbooks[i].Hosts = strings.Join(s, ",")
				break
			}
		}
	}
}

// return playbook section
func (p *Playbook) GetSection(section string) *PlaybookSection {

	for i, v := range p.playbooks {
		if v.Hosts == section {
			return p.playbooks[i]
		}
	}

	return nil
}

// adds set of vars to a given section
func (p *Playbook) AddVars(section string, vars map[string]string) {

	playbook := p.GetSection(section)
	if playbook.Vars == nil {
		playbook.Vars = make(map[string]string)
	}

	for k, v := range vars {
		playbook.Vars[k] = v
	}
}

//a generate a templates from static defined
// ansible template
func GenerateTemplate() *strings.Reader {
	return strings.NewReader(playbookTemplate)
}

// write playbook to a writer
func (p *Playbook) Write(writer io.Writer) error {

	if p != nil {
		out, err := yaml.Marshal(p.playbooks)
		if err != nil {
			return err
		}

		_, err = writer.Write(out)
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new playbook from playbook template file.
// in jettison template stored in template folder and assets
// tool generate a assets stored in const package.
// during make phase assets regenerated
func readFromFile(file string) (*Playbook, error) {

	fi, r, err := createReadBuffer(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read tempate from file")
	}
	defer fi.Close()

	playbook, err := makeTemplate(r)
	if err != nil {
		return nil, fmt.Errorf("failed parse default configuration %v", err)
	}

	return playbook, nil
}

//
// Read a nsx-t configuration file to buffer io return open file and reader.
// caller must close a file
func createReadBuffer(file string) (*os.File, *bufio.Reader, error) {

	fi, err := os.Open(file)
	if err == nil {
		// return buffer caller need close file
		return nil, bufio.NewReader(fi), nil
	}
	return nil, nil, err
}

//  Function generate a playbook section for each playbooks
//  single playbook can contain multiply sections
func makeTemplate(reader io.Reader) (*Playbook, error) {

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var playbook = &Playbook{}
	var s []*PlaybookSection

	err = yaml.Unmarshal(data, &s)
	if err != nil {
		return nil, err
	}
	playbook.playbooks = s

	return playbook, nil
}

// default template
const playbookTemplate = `
- hosts: 127.0.0.1
  connection: local
  roles:
    - download

- hosts: Controller, Worker, Ingress
  gather_facts: yes
  become: true
  roles:
    - basic
    - hosts
    - sshforuser

- hosts: Ingress
  become: true
  roles:
    - ingress
    - k8s-client
    - kubeproxy-certs

- hosts: Controller
  become: true
  roles:
    - master-node
    - k8s-client
    - encryption
    - etcd
    - kubemasters
    - k8sroles

- hosts: Worker
  become: true
  roles:
    - k8s-client
    - workers
`
