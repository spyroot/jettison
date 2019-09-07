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

Test for vcenter utils

Author Mustafa Bayramov
mbaraymov@vmware.com
*/

package tests

import (
	"context"
	"github.com/spyroot/jettison/internal"
	"github.com/spyroot/jettison/testutils"
	"github.com/spyroot/jettison/vcenter"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"reflect"
	"testing"
)

var (
	TestVim       *internal.Vim
	TestImageName = ""
	TestVmUuid    = ""
)

/**
     main test setup routine
     create vm in target cluster
	 in case cluster / vm name needs to be change change variable in testutil package
*/
func setupTestCase(t *testing.T) func(t *testing.T) {
	v, err := internal.InitEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	TestVim = v

	uuid, err := testutils.CreateVmifneeded(v.Ctx, v.GetVimClient())
	if err != nil {
		t.Fatal(err)
	}

	TestImageName = testutils.TestImageName
	TestVmUuid = uuid

	return func(t *testing.T) {
		//		t.Log("teardown test case")
	}
}

func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Log("setup sub test")
	return func(t *testing.T) {
		t.Log("teardown sub test")
	}
}

func TestChangePowerState(t *testing.T) {
	type args struct {
		ctx  context.Context
		c    *vim25.Client
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		want    *mo.VirtualMachine
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.ChangePowerState(tt.args.ctx, tt.args.c, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangePowerState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChangePowerState() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConnect(t *testing.T) {
	type args struct {
		ctx      context.Context
		hostname string
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    *govmomi.Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.Connect(tt.args.ctx, tt.args.hostname, tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Connect() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetNetworkAttr(t *testing.T) {
	type args struct {
		ctx       context.Context
		c         *vim25.Client
		vmVimName string
	}
	tests := []struct {
		name    string
		args    args
		want    *[]mo.Network
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.GetNetworkAttr(tt.args.ctx, tt.args.c, tt.args.vmVimName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetworkAttr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNetworkAttr() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSwitchUuid(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		ctx        context.Context
		c          *vim25.Client
		switchName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil test1",
			args: args{
				ctx:        TestVim.Ctx,
				c:          TestVim.GetVimClient(),
				switchName: "test-segment",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.GetSwitchUuid(tt.args.ctx, tt.args.c, tt.args.switchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSwitchUuid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != true && got != nil {
				log.Println(got)
			}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("GetNetworkAttr() got = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestGetNetworks(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		ctx  context.Context
		c    *vim25.Client
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil test1",
			args: args{
				ctx:  nil,
				c:    TestVim.GetVimClient(),
				uuid: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "wrong name",
			args: args{
				ctx:  TestVim.Ctx,
				c:    TestVim.GetVimClient(),
				uuid: TestVmUuid,
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "correct name",
			args: args{
				ctx:  TestVim.Ctx,
				c:    TestVim.GetVimClient(),
				uuid: "test-segment",
			},
			wantErr: false,
			want:    "test-segment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.GetNetworks(tt.args.ctx, tt.args.c, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetworks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != true && got != nil {
				if !reflect.DeepEqual(got.Name, tt.want) {
					t.Errorf("GetNetworks() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGetSummaryVm(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		ctx  context.Context
		c    *vim25.Client
		name string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "nil test1",
			args: args{
				ctx:  nil,
				c:    TestVim.GetVimClient(),
				name: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "nil test2",
			args: args{
				ctx:  TestVim.Ctx,
				c:    nil,
				name: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "wrong name",
			args: args{
				ctx:  TestVim.Ctx,
				c:    TestVim.GetVimClient(),
				name: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "Correct name",
			args: args{
				ctx:  TestVim.Ctx,
				c:    TestVim.GetVimClient(),
				name: TestVmUuid,
			},
			wantErr: false,
			want:    TestVmUuid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.GetSummaryVm(tt.args.ctx, tt.args.c, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVM() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != true && got != nil {
				gotUuid := got.Summary.Config.Uuid
				if !reflect.DeepEqual(gotUuid, tt.want) {
					t.Errorf("GetNetworks() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGetVmAttr(t *testing.T) {

	cmpName := vcenter.VmLookupMap["name"]
	cmpUuid := vcenter.VmLookupMap["uuid"]

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		ctx          context.Context
		c            *vim25.Client
		vmComparator vcenter.VmComparator
		attrValue    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil callback",
			args: args{ctx: TestVim.Ctx,
				c:            TestVim.GetVimClient(),
				vmComparator: nil,
				attrValue:    TestImageName,
			},
			wantErr: true,
		},
		{
			name: "lookup by name",
			args: args{ctx: TestVim.Ctx,
				c:            TestVim.GetVimClient(),
				vmComparator: cmpName,
				attrValue:    TestImageName,
			},
			wantErr: false,
		},
		{
			name: "lookup by vm uuid",
			args: args{ctx: TestVim.Ctx,
				c:            TestVim.GetVimClient(),
				vmComparator: cmpUuid,
				attrValue:    TestVmUuid,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcenter.GetVmAttr(tt.args.ctx, tt.args.c, tt.args.vmComparator, tt.args.attrValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVmAttr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				vm := *got
				for _, v := range vm.Network {
					log.Println(v.Reference())
				}
			}
			//			t.Log(got)
		})
	}
}

func TestVmFromCluster(t *testing.T) {

	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	type args struct {
		ctx         context.Context
		c           *vim25.Client
		vmName      string
		clusterName string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"get vm",
			args{TestVim.Ctx,
				TestVim.GetVimClient(),
				TestImageName,
				testutils.TestImageCluster,
			},
			false,
		},
		{
			"empty names",
			args{TestVim.Ctx,
				TestVim.GetVimClient(),
				"", "",
			},
			true,
		},

		{
			"nil client",
			args{TestVim.Ctx,
				nil,
				"",
				"",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, vm, err := vcenter.VmFromCluster(tt.args.ctx, tt.args.c, tt.args.vmName, tt.args.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("VmFromCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Println(vm)
		})
	}
}
