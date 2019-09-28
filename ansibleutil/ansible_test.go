package ansibleutil

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestRunAnsible(t *testing.T) {

	type args struct {
		ansibleCommand AnsibleCommand
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"invalid playbook",
			args{AnsibleCommand{
				Path:   "/usr/local/bin/ansible-playbook",
				CMD:    []string{"-K", "current.yml"},
				Config: "",
			},
			},
			true,
		},
		{
			"invalid cmd",
			args{AnsibleCommand{
				Path:   "/usr/local/bin/ansible-playbook2",
				CMD:    []string{"-K", "current.yml"},
				Config: "",
			},
			},
			true,
		},
		{
			"valid file", // note folder needs contain test.yml
			args{AnsibleCommand{
				Path:   "/usr/local/bin/ansible-playbook",
				CMD:    []string{"--extra-vars", "\"ansible_become_pass=VMware1!\"", "test.yml"},
				Config: "",
			},
			},
			false,
		},
		{
			"valid file long scenario", // note folder needs contain test.yml
			args{AnsibleCommand{
				Path:   "/usr/local/bin/ansible-playbook",
				CMD:    []string{"-b", "test1.yml"},
				Config: "",
			},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunAnsible(tt.args.ansibleCommand)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunAnsible() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Equal(t, got, "")
			}

			if err == nil {
				assert.NotNil(t, got)
				log.Println(got)
			}
		})
	}
}

func Test_notExist(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notExist(tt.args.file); got != tt.want {
				t.Errorf("notExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeHostInpoition(t *testing.T) {
	type args struct {
		filePath    string
		ansibleHost AnsibleHosts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := removeHostInpoition(tt.args.filePath, tt.args.ansibleHost); (err != nil) != tt.wantErr {
				t.Errorf("removeHostInpoition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_setConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		want1   string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := setConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("setConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("setConfig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("setConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
