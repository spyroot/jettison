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
	"github.com/eugenmayer/go-sshclient/sshwrapper"
	"github.com/spyroot/jettison/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// Function execute command on remote server via ssh and output the result as return
func RunRemoteCommand(sshenv config.SshGlobalEnvironments, host string, cmd string) (string, error) {

	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Println("failed retrieve user home sshKeyPath ", err)
		return "", err
	}

	sshKeyPath := filepath.Join(homePath, ".ssh/id_rsa")
	log.Println("Reading key from sshKeyPath", sshKeyPath)

	port, err := strconv.Atoi(sshenv.Port)
	if err != nil {
		return "", err
	}

	sshApi, err := sshwrapper.DefaultSshApiSetup(host, port, sshenv.Username, sshKeyPath)
	if err != nil {
		log.Println(err)
		log.Println("return error")
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
func SshCopyId(sshenv config.SshGlobalEnvironments, host string) error {

	var (
		sshHost string
	)

	log.Println("Copy ssh key to a host", host, " using ssh env", sshenv)

	sshHost = sshenv.Username + "@" + host
	subProcess := exec.Command(sshenv.SshpassPath,
		sshenv.SshpassPath, "-p", sshenv.Password, sshenv.SshCopyIdPath, "-p", sshenv.Port,
		"-o", "StrictHostKeyChecking=no", "-i", sshenv.PublicKey, sshHost)

	// set new env before we call ssh
	additionalEnv := "SSHPASS=" + sshenv.Password
	newEnv := append(os.Environ(), additionalEnv)
	subProcess.Env = newEnv

	stdin, err := subProcess.StdinPipe()
	if err != nil {
		log.Println("Failed create a pipe", err)
		return err
	}

	defer stdin.Close()

	subProcess.Stdout = os.Stdout
	subProcess.Stderr = os.Stderr

	if err = subProcess.Start(); err != nil {
		log.Println("Failed swan ssh process")
		return err
	}

	err = subProcess.Wait()
	if err != nil {
		log.Println("Wait() failed", err)
		return err
	}

	return nil
}
