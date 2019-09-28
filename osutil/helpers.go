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

OS Helper function

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package osutil

import (
	"github.com/spyroot/jettison/logging"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Parser struct {
	String string
	pos    int64
}

/**
  Function take dir that use used for chdir before executing cmd
*/
func ChangeAndExec(dir string, cmd string) error {

	err := os.Chdir(dir)
	if err != nil {
		logging.ErrorLogging(err)
		return err
	}

	_, err = exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println(" failed execute:", cmd)
		logging.ErrorLogging(err)
		return err
	}

	return nil
}

/**
  Function take dir that use used for chdir before executing cmd
*/
func ChangeAndExecWithSdout(dir string, cmd string) (string, error) {

	err := os.Chdir(dir)
	if err != nil {
		logging.ErrorLogging(err)
		return "", err
	}

	s, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		log.Println(" failed execute:", cmd)
		logging.ErrorLogging(err)
		return "", err
	}
	return string(s), nil
}

// Function check write/read access
func CheckWriteReadAccess(filename string) bool {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_RDONLY, 0666)
	if err != nil {
		if os.IsPermission(err) {
			log.Println("Error: Write permission denied.")
		}
		return false
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Println("failed to close", filename, err)
		}
	}()

	return true
}

// function check if file exist or not
func CheckIfExist(filename string) bool {
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return true
	}
	return true
}

// getEnvString returns string from environment variable.
func GetEnvString(v string, def string) string {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	return r
}

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

// getEnvBool returns boolean from environment variable.
func GetEnvBool(v string, def bool) bool {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	switch strings.ToLower(r[0:1]) {
	case "t", "y", "1":
		return true
	}

	return false
}
