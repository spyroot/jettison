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

main entry
spyroot
mbaraymov@vmware.com
*/

package main

import (
	"encoding/base64"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spyroot/jettison/consts"
	"github.com/spyroot/jettison/internal"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/logging"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

// unpack all ansible file to target dir
func SetupAnsible(homeDir string) error {

	if len(homeDir) == 0 {
		return fmt.Errorf("invalid path")
	}

	// create root role
	baseDir := homeDir
	logging.Notification("Generating ansible roles and templates ", baseDir)

	dirs := consts.AssetsDirs()
	for _, v := range dirs {
		subDir := path.Join(baseDir, v)
		err := os.MkdirAll(subDir, os.ModePerm)
		if err != nil {
			logging.ErrorLogging(err)
			return err
		}
	}

	// iterate for each assets a key a path from a base
	// decode a file and write to destination
	for f, v := range consts.AllAssets() {
		data, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		filePath := path.Join(baseDir, f)
		fi, err := os.OpenFile(filePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}

		_, err = fi.Write(data)
		if err != nil {
			if err := fi.Close(); err != nil {
				return err
			}
			return err
		}

		if err := fi.Close(); err != nil {
			return err
		}
	}

	return nil
}

// initialize jettison that involves initializing a vim,
// database and all dependency
func initJettison() (*internal.Vim, *internal.Deployment2, error) {

	jetConfig, err := internal.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}

	//build a template based on yaml on configuration file
	var templateList []*jettypes.NodeTemplate
	for k, v := range jetConfig.Infra.Scenario {
		v.Type = jettypes.GetNodeType(k)
		templateList = append(templateList, v)
	}

	// build a scenario based on template
	scenario, err := internal.CreateScenario(templateList, jetConfig.GetDeploymentName())
	if err != nil {
		return nil, nil, nil
	}

	// setup ansible environments
	err = SetupAnsible(jetConfig.GetAnsible().AnsibleConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed create project dir error: %v", err)
	}

	// init a vim
	vim, err := internal.NewVim()
	if err != nil {
		return nil, nil, err
	}

	return vim, scenario, nil
}

// Delete deployment
// TODO unfinished
func DeleteDeployment() *cobra.Command {

	cmd := &cobra.Command{
		Use: "kill",
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) == 0 {
				return fmt.Errorf("kill needs a project name")
			}

			jetConfig, err := internal.ReadConfig()
			if err != nil {
				log.Fatal(err)
			}

			if jetConfig.GetDeploymentName() == args[0] {
				log.Printf("Found exiting project")
			}

			return nil
		},
	}
	return cmd
}

func RegeneratePlaybook() *cobra.Command {

	cmd := &cobra.Command{
		Use: "ansible",
		RunE: func(cmd *cobra.Command, args []string) error {

			vim, scenario, err := initJettison()
			if err != nil {
				return err
			}
			defer vim.Database().Close()

			deployer := internal.NewDeployer(scenario, vim)

			_, err = deployer.Execute(jettypes.AnsiblePlaybook)

			return nil
		},
	}

	return cmd
}

func RegenerateInventory() *cobra.Command {

	cmd := &cobra.Command{
		Use: "inventory",
		RunE: func(cmd *cobra.Command, args []string) error {

			vim, scenario, err := initJettison()
			if err != nil {
				return err
			}
			defer vim.Database().Close()

			deployer := internal.NewDeployer(scenario, vim)

			_, err = deployer.Execute(jettypes.AnsibleInventory)

			return nil
		},
	}

	return cmd
}

var deployer *internal.Deployer

// Main root deploy command
// It passed vim to deployer that will start deployment routine
func Deploy() *cobra.Command {

	cmd := &cobra.Command{
		Use: "deploy",
		RunE: func(cmd *cobra.Command, args []string) error {

			vim, scenario, err := initJettison()
			if err != nil {
				return err
			}
			defer vim.Database().Close()

			deployer = internal.NewDeployer(scenario, vim)

			err = deployer.Deploy()

			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

// root command for build that by default regenerate
// all ansible files for a project
func Ansible() *cobra.Command {

	cmd := &cobra.Command{
		Use: "build",
		RunE: func(cmd *cobra.Command, args []string) error {

			vim, scenario, err := initJettison()
			if err != nil {
				return err
			}
			defer vim.Database().Close()

			deployer := internal.NewDeployer(scenario, vim)

			_, err = deployer.Execute(jettypes.AnsiblePlaybook)
			_, err = deployer.Execute(jettypes.AnsibleInventory)

			return nil
		},
	}

	cmd.AddCommand(RegeneratePlaybook())
	cmd.AddCommand(RegenerateInventory())

	return cmd
}

// root command for build that by default regenerate
// all ansible files for a project
func Build() *cobra.Command {

	cmd := &cobra.Command{
		Use: "build",
		RunE: func(cmd *cobra.Command, args []string) error {

			vim, scenario, err := initJettison()
			if err != nil {
				return err
			}
			defer vim.Database().Close()

			deployer := internal.NewDeployer(scenario, vim)

			_, err = deployer.Execute(jettypes.AnsiblePlaybook)
			_, err = deployer.Execute(jettypes.AnsibleInventory)

			return nil
		},
	}

	cmd.AddCommand(RegeneratePlaybook())
	cmd.AddCommand(RegenerateInventory())

	return cmd
}

//func handleSignals(*internal.Vim) {
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
//	for sig := range c {
//		switch sig {
//		case syscall.SIGINT, syscall.SIGTERM:
//			//	Cleanup(Cleanup)
//			return
//		case syscall.SIGHUP:
//			log.Println("do nothing")
//		}
//	}
//}

func signalHandler(deployer *internal.Deployer) {
	log.Println("Cleaning environment")
	deployer.Stop()
}

// main entry to jettison
func main() {

	cmd := &cobra.Command{
		Use:          "jettison",
		Short:        "jettison",
		SilenceUsage: true,
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		signalHandler(deployer)
		os.Exit(1)
	}()

	cmd.AddCommand(DeleteDeployment())
	cmd.AddCommand(Build())
	cmd.AddCommand(Deploy())
	cmd.AddCommand(Ansible())

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
