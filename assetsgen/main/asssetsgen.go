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

very simply tool to generate assets for a package.
jettison already contains asset in assets.go

Author Mustafa Bayramov
mbaraymov@vmware.com
*/
package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const AssetsFile = "/Users/spyroot/go/src/github.com/spyroot/jettison/src/consts/assets.go"
const Base = "/Users/spyroot/go/src/github.com/spyroot/jettison/src/templates"
const PackageName = " consts"

func generateAssets() error {

	var subDirs []string
	var files map[string]string

	err := filepath.Walk(Base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			rel, err := filepath.Rel(Base, path)
			if err != nil {
				return err
			}
			if rel != "." {
				quoted := fmt.Sprintf("\"%s\"", rel)
				subDirs = append(subDirs, quoted)
			}
		}

		if !info.IsDir() {
			rel, err := filepath.Rel(Base, path)
			if err != nil {
				return err
			}

			if files == nil {
				files = make(map[string]string)
			}

			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			files[rel] = base64.StdEncoding.EncodeToString(data)
		}

		return nil
	})

	if err != nil {
		return err
	}

	assetsDir := fmt.Sprintf("var dirs = []string{%s}", strings.Join(subDirs, ",\n"))

	var content []string
	for k, v := range files {
		content = append(content, fmt.Sprintf("\"%s\" : \"%s\"", k, v))
	}
	assetsContent := fmt.Sprintf("var assets = map[string]string {%s}", strings.Join(content, ",\n"))

	fi, err := os.OpenFile(AssetsFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()

	pkgname := fmt.Sprintf("package %s\n", PackageName)
	if _, err := fi.Write([]byte(pkgname)); err != nil {
		return err
	}
	if _, err := fi.Write([]byte(assetsDir)); err != nil {
		return err
	}
	if _, err := fi.Write([]byte("\n")); err != nil {
		return err
	}
	if _, err := fi.Write([]byte(assetsContent)); err != nil {
		return err
	}

	return nil
}

func main() {

	generateAssets()
}
