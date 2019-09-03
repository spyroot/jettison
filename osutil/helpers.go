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
	"log"
	"os"
	"strings"
)

type Parser struct {
	String string
	pos    int64
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
	fileInfo, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("File does not exist.")
		}
		return true
	}
	log.Println("File does exist. File information:")
	log.Println(fileInfo)
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
