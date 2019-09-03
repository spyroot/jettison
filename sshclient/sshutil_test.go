package sshclient

import (
	"github.com/spyroot/jettison/config"
	"strings"
	"testing"
)

// read global config.
var appConfig, _ = config.ReadConfig()

func Test_runRemoteCommand(t *testing.T) {
	type args struct {
		sshenv config.SshGlobalEnvironments
		host   string
		cmd    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"empty config",
			args{
				sshenv: config.SshGlobalEnvironments{},
				host:   "172.16.81.1",
				cmd:    "ls",
			},
			"",
			true,
		},
		{
			"valid host, wrong credentials",
			args{
				sshenv: config.SshGlobalEnvironments{},
				host:   "172.16.149.222",
				cmd:    "/bin/ls",
			},
			"",
			true,
		},
		{
			"create file",
			args{
				sshenv: appConfig.Infra.SshDefaults,
				host:   "172.16.149.222",
				cmd:    "/bin/touch test",
			}, "",
			false,
		},
		{
			"echo message",
			args{
				sshenv: appConfig.Infra.SshDefaults,
				host:   "172.16.149.222",
				cmd:    "/bin/echo test123",
			},
			"test123",
			false,
		},
		{
			"create file ls file",
			args{
				sshenv: appConfig.Infra.SshDefaults,
				host:   "172.16.149.222",
				cmd:    "/bin/touch test123 | ls test123",
			},
			"test123",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RunRemoteCommand(tt.args.sshenv, tt.args.host, tt.args.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("runRemoteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got = strings.TrimRight(got, "\r\n")
			if got != tt.want {
				t.Errorf("runRemoteCommand() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// test ssh copy
func Test_sshCopyid(t *testing.T) {
	type args struct {
		sshenv config.SshGlobalEnvironments
		host   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// First test
		{
			"empty config",
			args{
				sshenv: config.SshGlobalEnvironments{},
				host:   "172.16.81.1",
			}, true,
		},

		// Second test valid property but wrong globals
		{
			"wrong password",
			args{
				sshenv: appConfig.Infra.SshDefaults,
				host:   "172.16.81.1",
			}, true,
		},
		// Second test valid property but wrong globals
		{
			"valid config",
			args{
				sshenv: appConfig.Infra.SshDefaults,
				host:   "172.16.149.222",
			}, false,
		},
		// bogus host
		{
			"bogus host",
			args{
				sshenv: appConfig.Infra.SshDefaults,
				host:   "224.1.1.1",
			}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SshCopyId(tt.args.sshenv, tt.args.host); (err != nil) != tt.wantErr {
				t.Errorf("sshCopyid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
