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

Ansible json.  Work in progress still trying to figure do I want parse everything
or only status

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package ansibleutil

import "time"

type HostStatus struct {
	Changed     int `json:"changed"`
	Failures    int `json:"failures"`
	Ignored     int `json:"ignored"`
	Ok          int `json:"ok"`
	Rescued     int `json:"rescued"`
	Skipped     int `json:"skipped"`
	Unreachable int `json:"unreachable"`
}

type HostTask struct {
	AnsibleNoLog bool   `json:"_ansible_no_log"`
	Action       string `json:"action"`
	AnsibleFacts struct {
		DiscoveredInterpreterPython string `json:"discovered_interpreter_python"`
	} `json:"ansible_facts"`
	Changed    bool   `json:"changed"`
	Cmd        string `json:"cmd"`
	Delta      string `json:"delta"`
	End        string `json:"end"`
	Invocation struct {
		ModuleArgs struct {
			Data            string `json:"data"`
			RawParams       string `json:"_raw_params"`
			UsesShell       string `json:"_uses_shell"`
			Argv            string `json:"argv"`
			Chdir           string `json:"chdir"`
			Creates         string `json:"creates"`
			Executable      string `json:"executable"`
			Removes         string `json:"removes"`
			Stdin           string `json:"stdin"`
			StdinAddNewline bool   `json:"stdin_add_newline"`
			StripEmptyEnds  bool   `json:"strip_empty_ends"`
			Warn            bool   `json:"warn"`
		} `json:"module_args"`
	} `json:"invocation"`
	Rc          int           `json:"rc"`
	Start       string        `json:"start"`
	Stderr      string        `json:"stderr"`
	StderrLines []interface{} `json:"stderr_lines"`
	Stdout      string        `json:"stdout"`
	StdoutLines []string      `json:"stdout_lines"`
	Ping        string        `json:"ping"`
}

type AnsiblePing struct {
	CustomStats struct {
	} `json:"custom_stats"`
	GlobalCustomStats struct {
	} `json:"global_custom_stats"`
	Plays []struct {
		Play struct {
			Duration struct {
				End   time.Time `json:"end"`
				Start time.Time `json:"start"`
			} `json:"duration"`
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"play"`
		Tasks []struct {
			Hosts map[string]HostTask `json:"hosts"`
			Task  struct {
				Duration struct {
					End   time.Time `json:"end"`
					Start time.Time `json:"start"`
				} `json:"duration"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"task"`
		} `json:"tasks"`
	} `json:"plays"`
	Stats map[string]HostStatus `json:"stats"`
}
