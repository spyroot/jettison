package ansibleutil

import (
	"reflect"
	"testing"
)

func TestAddSlaveHost(t *testing.T) {
	type args struct {
		filePath    string
		ansibleHost AnsibleHosts
		section     string
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
			if err := AddSlaveHost(tt.args.filePath, tt.args.ansibleHost, tt.args.section); (err != nil) != tt.wantErr {
				t.Errorf("AddSlaveHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//
//func TestCreateTenantPlaybook(t *testing.T) {
//	type args struct {
//		filePath     string
//		templatePath string
//		vars         map[string]string
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if err := CreateTenantPlaybook(tt.args.filePath, tt.args.templatePath, tt.args.vars); (err != nil) != tt.wantErr {
//				t.Errorf("CreateTenantPlaybook() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}

func TestPing(t *testing.T) {
	type args struct {
		ansibleCommand AnsibleCommand
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Ping(tt.args.ansibleCommand)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ping() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveSlaveHost(t *testing.T) {
	type args struct {
		filePath    string
		ansibleHost AnsibleHosts
		section     string
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
			if err := RemoveSlaveHost(tt.args.filePath, tt.args.ansibleHost, tt.args.section); (err != nil) != tt.wantErr {
				t.Errorf("RemoveSlaveHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_appendAtPosition(t *testing.T) {
	type args struct {
		filePath    string
		ansibleHost AnsibleHosts
		pos         int64
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
			if err := appendToExistingGroup(tt.args.filePath, tt.args.ansibleHost, tt.args.pos); (err != nil) != tt.wantErr {
				t.Errorf("appendAtPosition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_appendHost(t *testing.T) {
	type args struct {
		filePath    string
		ansibleHost AnsibleHosts
		section     string
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
			if err := appendHost(tt.args.filePath, tt.args.ansibleHost, tt.args.section); (err != nil) != tt.wantErr {
				t.Errorf("appendHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parse(t *testing.T) {
	type args struct {
		filePath string
		start    int64
		parse    func(string) (string, bool)
	}
	tests := []struct {
		name    string
		args    args
		want    []Parser
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(tt.args.filePath, tt.args.start, tt.args.parse)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeHost(t *testing.T) {
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
			if err := removeHost(tt.args.filePath, tt.args.ansibleHost); (err != nil) != tt.wantErr {
				t.Errorf("removeHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeHost2(t *testing.T) {
	//type args struct {
	//	filePath string
	//	remove   string
	//	pos      int64
	//}
	//tests := []struct {
	//	name    string
	//	args    args
	//	wantErr bool
	//}{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		if err := removeHost2(tt.args.filePath, tt.args.remove, tt.args.pos); (err != nil) != tt.wantErr {
	//			t.Errorf("removeHost2() error = %v, wantErr %v", err, tt.wantErr)
	//		}
	//	})
	//}

	_ = AddSlaveHost("/usr/local/etc/ansible/hosts",
		AnsibleHosts{
			Name:     "ingress.39cb06be-a089-452c-9525-fb2d2e9595b6.vmwarelab.edu",
			Hostname: "172.16.1.1",
			Port:     22,
			User:     "vmware",
		}, "Test-controllers")

	//_ = AddSlaveHost("/usr/local/etc/ansible/hosts",
	//	AnsibleHosts{
	//		Name:     "ingress.39cb06be-a089-452c-9525-fb2d2e9595b6.vmwarelab.edu",
	//		Hostname: "172.16.1.1",
	//		Port:     "22",
	//		User:     "vmware",
	//	}, "Test-controllers")

	//_ = AddSlaveHost("/usr/local/etc/ansible/hosts",
	//	AnsibleHosts{
	//		Name:     "ingress.abc.vmwarelab.edu",
	//		Hostname: "172.16.1.1",
	//		Port:     "22",
	//		User:     "vmware",
	//	}, "Test-ingress")

	//
	//_ = AddSlaveHost("/usr/local/etc/ansible/hosts",
	//	AnsibleHosts{
	//		Name:     "controller.abc.vmwarelab.edu",
	//		Hostname: "172.16.1.2",
	//		Port:     "22",
	//		User:     "vmware",
	//	}, "Test-controllers")

	//_ = AddSlaveHost("/usr/local/etc/ansible/hosts",
	//	AnsibleHosts{
	//		Name:     "ingress.39cb06be-a089-452c-9525-666.vmwarelab.edu",
	//		Hostname: "172.16.1.1",
	//		Port:     "22",
	//		User:     "vmware",
	//	}, "Test-ingress")
	//
	//_ = RemoveSlaveHost("/usr/local/etc/ansible/hosts",
	//	AnsibleHosts{
	//		Name:     "ingress.39cb06be-a089-452c-9525-fb2d2e9595b6.vmwarelab.edu",
	//		Hostname: "172.16.1.1",
	//		Port:     "22",
	//		User:     "vmware",
	//	}, "Test-ingress")

}
