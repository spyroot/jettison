package ansibleutil

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path"
)

type AnsibleHostVars struct {
	Podnet          string `yaml:"podnet"`
	CaCertificate   string `yaml:"ca-cert"`
	KubePrivateKey  string `yaml:"kube-pem"`
	KubeCertificate string `yaml:"kube-cert"`
	Hostname        string
	HomePath        string
}

/**
  Function serialize itself to a ansible group var
*/
func (a *AnsibleHostVars) WriteToFile() error {

	dir := path.Join(a.HomePath, "/host_vars")
	// create dir for groups_var
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed create %s", dir)
	}

	// create a new file
	file := path.Join(a.HomePath, "/host_vars/"+a.Hostname)
	inputFile, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
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
		return fmt.Errorf("failed marshal yaml file file: %s", err)
	}

	_, err = inputFile.Write(yamlBuffer)
	if err != nil {
		return err
	}

	return nil
}
