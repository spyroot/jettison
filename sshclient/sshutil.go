/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Tests for ssh utils

// Dependence
spyroot  ~ | .ansible  go get -v github.com/eugenmayer/go-sshclient/sshwrapper

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package sshclient

import (
	"fmt"
	"github.com/eugenmayer/go-sshclient/sshwrapper"
	"github.com/spyroot/jettison/consts"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type SshEnvironments interface {
	Username() string
	Password() string
	PublicKey() string
	PrivateKey() string
	SshpassTool() string
	SshCopyIdTool() string
	Port() int
}

/*
   Function execute command on remote server via ssh and output the result as return
*/
func RunRemoteCommand(ssh SshEnvironments, host string, cmd string) (string, error) {

	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed retrieve user home folder %s", err)
	}

	sshKeyPath := ssh.PrivateKey()
	if ssh.PrivateKey() == "" {
		sshKeyPath = filepath.Join(homePath, consts.DefaultPublicKey)
	}

	sshApi, err := sshwrapper.DefaultSshApiSetup(host, ssh.Port(), ssh.Username(), sshKeyPath)
	if err != nil {
		log.Println("ssh return error", err)
		return "", err
	}

	stdout, stderr, err := sshApi.Run(cmd)
	if err != nil {
		log.Print(stdout)
		log.Print(stderr)
		return "", err
	}

	return stdout, nil
}

// Function open an ssh session and issue ssh-copy-id.
// note since since ssh-copy-id never trust stdin for initial password authentication
// function uses sshpass to pass a password and username
// wget https://raw.githubusercontent.com/ansible/ansible/devel/examples/ansible.cfg
func SshCopyId(ssh SshEnvironments, host string) error {

	var (
		sshHost string
	)

	log.Println("Copy ssh key to a host", host, " using public key:", ssh.PublicKey())
	//osutil.CheckIfExist(sshenv.SshpassPath())

	sshHost = ssh.Username() + "@" + host

	//sshHost = sshenv.Username + "@" + host
	//subProcess := exec.Command(sshenv.SshpassPath,
	//	sshenv.SshpassPath, "-p", sshenv.Password, sshenv.SshCopyIdPath, "-p", strconv.Itoa(sshenv.Port),
	//	"-o", "StrictHostKeyChecking=no", "-i", sshenv.PublicKey, sshHost)

	subProcess := exec.Command(ssh.SshpassTool(),
		ssh.SshpassTool(), "-p", ssh.Password(), ssh.SshCopyIdTool(), "-p", strconv.Itoa(ssh.Port()),
		"-o", "StrictHostKeyChecking=no", "-i", ssh.PublicKey(), sshHost)

	// set new env before we call ssh
	additionalEnv := "SSHPASS=" + ssh.Password()
	newEnv := append(os.Environ(), additionalEnv)
	subProcess.Env = newEnv

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		log.Println("Failed create a pipe", err)
		return err
	}

	defer func() {
		if err := stdin.Close(); err != nil {
			log.Println("failed to close stdin", err)
		}
	}()

	//subProcess.Stdout = os.Stdout
	//subProcess.Stderr = os.Stderr

	if err = subProcess.Start(); err != nil {
		log.Println("failed create ssh sub-process")
		return err
	}

	err = subProcess.Wait()
	if err != nil {
		log.Println("Wait() failed", err)
		return err
	}

	return nil
}
