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

AnsibleHosts utils.  The main purpose be able programmatically add bunch of hosts
to a ansible inventory and manage this host

After brew
create files
sudo mkdir -p /var/log/ansible
sudo touch /var/log/ansible/ansible.log
sudo chown -R ${USER}:wheel /var/log/ansible
sudo chmod 775 /var/log/ansible
sudo chmod 774 /var/log/ansible/ansible.log

wget https://raw.githubusercontent.com/ansible/ansible/devel/examples/ansible.cfg

hostfile       = inventories/hosts
wget https://raw.githubusercontent.com/cyverse/ansible-tips-and-tricks/master/ansible.cfg

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package ansibleutil

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

/*
   Struct that holds string that parser matched based and position
*/
type Parser struct {
	String string
	pos    int64
}

/*
   Struct that holds string that parser matched based and position
   Serialize inventory information for hosts.

   Example:
        localhost1 ansible_ssh_port=2222 ansible_ssh_user=root ansible_host=127.0.0.1
*/
type AnsibleHosts struct {
	Name     string
	Hostname string
	Port     string
	User     string
}

/*
   Struct that holds string that parser matched based and position
*/
type AnsibleCommand struct {
	Path   string   // path to ansible exec file
	CMD    []string // slice for ansible command "controller -m ping" etc"
	Config string   // path to ansible.cfg
}

// Function takes file path argument that must point to a valid ansible host file
// and start position from where parse need move file cursor. The last argument is
// function that must do custom token match.
func parse(filePath string, start int64, parse func(string) (string, bool)) ([]Parser, error) {

	var (
		lineCounter = 0
		results     []Parser
	)

	inputFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", filePath, err)
		}
	}()

	if _, err := inputFile.Seek(start, 0); err != nil {
		return results, err
	}

	scanner := bufio.NewScanner(inputFile)
	pos := start
	scanLines := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = bufio.ScanLines(data, atEOF)
		pos += int64(advance)
		return
	}

	scanner.Split(scanLines)

	for scanner.Scan() {
		if output, add := parse(scanner.Text()); add {
			results = append(results, Parser{String: output, pos: pos})
		}
		lineCounter++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Function appends a new section [controllers], [nodes] etc and
// appends at the end of a file
func appendHost(filePath string, ansibleHost AnsibleHosts, section string) error {

	inputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed opening file %s", err)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", filePath, err)
		}
	}()

	// create a new section
	newSection := "\n[" + section + "]\n"
	_, err = inputFile.WriteString(newSection)
	if err != nil {
		return fmt.Errorf("failed opening file %s", err)
	}

	// add host to a ansible section
	newHostEntry := ansibleHost.Name +
		" ansible_ssh_port=" + ansibleHost.Port +
		" ansible_ssh_user=" + ansibleHost.User +
		" ansible_host=" + ansibleHost.Hostname +
		"\n"

	_, err = inputFile.WriteString(newHostEntry)
	if err != nil {
		return fmt.Errorf("failed opening file %s", err)
	}

	return nil
}

// Function appends a new section [controllers], [nodes] etc and
// appends at the end of a file
func removeHost(filePath string, ansibleHost AnsibleHosts, pos int64) error {

	inputFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", filePath, err)
		}
	}()

	tmpFileName := filePath + ".tmp"
	tmpFile, err := os.OpenFile(tmpFileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed opening file %s", err)
	}

	defer func() {
		if err := tmpFile.Close(); err != nil {
			log.Println("failed to close", tmpFileName, err)
		}
	}()

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		// skip line and write to a temp file
		oldHostEntry := ansibleHost.Name +
			" ansible_ssh_port=" + ansibleHost.Port +
			" ansible_ssh_user=" + ansibleHost.User +
			" ansible_host=" + ansibleHost.Hostname
		if !strings.Contains(line, oldHostEntry) {
			_, err = tmpFile.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("failed to write to a %s file error : %s", tmpFileName, err)
			}
		}
	}

	if scanner.Err() != nil {
		return fmt.Errorf("failed opening file %s", scanner.Err())
	}

	// rename that will overwrite old file
	err = os.Rename(tmpFileName, filePath)
	if err != nil {
		return fmt.Errorf("failed opening file %s", scanner.Err())
	}

	return nil
}

// Function adds a host at the give position in ansible hosts file.
// position mainly used for a case when we want add host file in specific section
// in ansible hosts file
func appendAtPosition(filePath string, ansibleHost AnsibleHosts, pos int64) error {

	inputFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed opening file %s", err)
	}

	defer func() {
		if err := inputFile.Close(); err != nil {
			log.Println("failed to close", filePath, err)
		}
	}()

	if _, err := inputFile.Seek(pos, 0); err != nil {
		return err
	}

	// add host to a ansible section
	newHostEntry := ansibleHost.Name +
		" ansible_ssh_port=" + ansibleHost.Port +
		" ansible_ssh_user=" + ansibleHost.User +
		" ansible_host=" + ansibleHost.Hostname +
		"\n"

	_, err = inputFile.WriteString(newHostEntry)
	if err != nil {
		return fmt.Errorf("failed opening file %s", err)
	}

	return nil
}

// Function removes a host from a ansible hosts file
func RemoveSlaveHost(filePath string, ansibleHost AnsibleHosts, section string) error {

	sectionStart, err := parse(filePath, 0, func(s string) (string, bool) {
		r, _ := regexp.Compile("\\[" + section + "\\]")
		if r.MatchString(s) {
			return s, true
		}
		return s, false
	})

	if err != nil {
		return fmt.Errorf("error while parsing file: %s", err)
	}

	// nothing to remove
	if len(sectionStart) == 0 {
		return nil
	}

	regex := ansibleHost.Name +
		" ansible_ssh_port=" + ansibleHost.Port +
		" ansible_ssh_user=" + ansibleHost.User +
		" ansible_host=" + ansibleHost.Hostname

	result, err := parse(filePath, sectionStart[0].pos,
		func(s string) (string, bool) {
			r, _ := regexp.Compile(regex)
			if r.MatchString(s) {
				return s, true
			}
			return s, false
		})

	if len(result) > 0 {
		startPos := result[0].pos - int64(len(regex)) - 1
		err = removeHost(filePath, ansibleHost, startPos)
		if err != nil {
			return nil
		}
	}

	return nil
}

// Function adds a host to specific section in ansible hosts file.
// if section not present function will add a new section and add a host
func AddSlaveHost(filePath string, ansibleHost AnsibleHosts, section string) error {

	sectionStart, err := parse(filePath, 0, func(s string) (string, bool) {
		r, _ := regexp.Compile("\\[" + section + "\\]")
		if r.MatchString(s) {
			return s, true
		}
		return s, false
	})

	if err != nil {
		return fmt.Errorf("error while parsing file: %s", err)
	}
	if len(sectionStart) > 1 {
		return fmt.Errorf("potential duplicate section in ansible hosts file")
	}
	if len(sectionStart) == 0 {
		log.Println("Appending hosts", ansibleHost.Hostname, "to a section", section)
		err = appendHost(filePath, ansibleHost, section)
		if err != nil {
			return fmt.Errorf("failed adjust ansible hosts file %s", err)
		}
		return nil
	}

	result, err := parse(filePath, sectionStart[0].pos,
		func(s string) (string, bool) {
			regex := ansibleHost.Name +
				" ansible_ssh_port=" + ansibleHost.Port +
				" ansible_ssh_user=" + ansibleHost.User +
				" ansible_host=" + ansibleHost.Hostname
			r, _ := regexp.Compile(regex)
			if r.MatchString(s) {
				return s, true
			}
			return s, false
		})

	if len(result) == 0 {
		log.Println("Appending hosts", ansibleHost.Hostname, "to a section", section)
		err = appendAtPosition(filePath, ansibleHost, sectionStart[0].pos)
		return err
	}

	return nil
}

// Function executes ansible ping
func Ping(ansibleCommand AnsibleCommand) (string, error) {

	if len(ansibleCommand.CMD) == 0 {
		return "", fmt.Errorf("empty ansible command")
	}

	if ansibleCommand.Path == "" {
		return "", fmt.Errorf("empty path to ansible tool")
	}

	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed retrieve user home dir %s", err)
	}

	ansibleConfig := ansibleCommand.Config
	if ansibleCommand.Config == "" {
		ansibleConfig = filepath.Join(homePath, ".ansible")
	}

	err = os.Setenv("ANSIBLE_CONFIG", ansibleConfig)
	if err != nil {
		return "", fmt.Errorf("can't set enviroment variable ANSIBLE_CONFIG  %s", err)
	}

	subProcess := exec.Command(ansibleCommand.Path, ansibleCommand.CMD...)
	log.Println("Executing ansible cmd", subProcess.Args)

	// set new env before we call ssh
	additionalEnv := "ANSIBLE_CONFIG=" + ansibleConfig
	newEnv := append(os.Environ(), additionalEnv)
	subProcess.Env = newEnv

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed create stdin pipe:  %s", err)
	}

	defer func() {
		if err := stdin.Close(); err != nil {
			log.Println("failed to close stdin", err)
		}
	}()

	// wait will close stdout
	stdout, err := subProcess.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed create stout pipe:  %s", err)
	}

	if err = subProcess.Start(); err != nil {
		return "", fmt.Errorf("failed start ansible process:  %s", err)
	}

	//	read json output to array and return entire buffer
	b, err := ioutil.ReadAll(stdout)
	if err != nil {
		return "", fmt.Errorf("failed read from stdout:  %s", err)
	}

	err = subProcess.Wait()
	if err != nil {
		return "", fmt.Errorf("wait() failed:  %s", err)
	}

	return string(b), nil
}
