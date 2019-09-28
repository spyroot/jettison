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

package vcenter

import (
	"context"
	"fmt"
	"github.com/spyroot/jettison/internal"
	"github.com/spyroot/jettison/logging"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"reflect"
	"testing"
)

//var (
//	TestVim       *internal.Vim
//	TestImageName = ""
//	TestVmUuid    = ""
//)

//
//
func setupSubTest(t *testing.T) func(t *testing.T) {
	t.Log("setup sub test")
	return func(t *testing.T) {
		t.Log("teardown sub test")
	}
}

func TestDeleteFolder(t *testing.T) {

	client, ctx, teardown := setupTest(t)
	_ = teardown

	folder, err := find.NewFinder(client.Client).DefaultFolder(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, folder)

	folder.CreateFolder(ctx, "jettest")

	type args struct {
		ctx        context.Context
		c          *vim25.Client
		folderName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Folder test",
			args{
				ctx:        ctx,
				c:          client.Client,
				folderName: "jettest",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := DeleteFolder(tt.args.ctx, tt.args.c, tt.args.folderName)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

/**

 */
func TestChangePowerState(t *testing.T) {

	env, c := VimSetupHelper()

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
			got, err := ChangePowerState(tt.args.ctx, tt.args.c, tt.args.uuid)
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

	env, c := VimSetupHelper()

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
			got, err := Connect(tt.args.ctx, tt.args.hostname, tt.args.username, tt.args.password)
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

	env, c := VimSetupHelper()

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
			got, err := GetNetworkAttr(tt.args.ctx, tt.args.c, tt.args.vmVimName)
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

	env, c := VimSetupHelper()

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
			got, err := GetSwitchUuid(tt.args.ctx, tt.args.c, tt.args.switchName)
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

	env, c := VimSetupHelper()

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
				ctx:  env.ctx,
				c:    TestVim.GetVimClient(),
				uuid: TestVmUuid,
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "correct name",
			args: args{
				ctx:  env.ctx,
				c:    TestVim.GetVimClient(),
				uuid: "test-segment",
			},
			wantErr: false,
			want:    "test-segment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNetworks(tt.args.ctx, tt.args.c, tt.args.uuid)
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

	env, c := VimSetupHelper()

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
				c:    c.Client,
				name: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "nil test2",
			args: args{
				ctx:  env.ctx,
				c:    nil,
				name: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "wrong name",
			args: args{
				ctx:  env.ctx,
				c:    c.Client,
				name: "edge01.vmwarelab.edu",
			},
			wantErr: true,
			want:    "",
		},
		{
			name: "Correct name",
			args: args{
				ctx:  env.ctx,
				c:    c.Client,
				name: TestVmUuid,
			},
			wantErr: false,
			want:    TestVmUuid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSummaryVm(tt.args.ctx, tt.args.c, tt.args.name)
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

//
//  gett vm attributes test
func TestGetVmAttr(t *testing.T) {

	env, c := VimSetupHelper()

	cmpName := VmSearchHandler["name"]
	cmpUuid := VmSearchHandler["uuid"]

	type args struct {
		ctx          context.Context
		c            *vim25.Client
		vmComparator VmComparator
		attrValue    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil callback",
			args: args{ctx: env.ctx,
				c:            c.Client,
				vmComparator: nil,
				attrValue:    TestImageName,
			},
			wantErr: true,
		},
		{
			name: "lookup by name",
			args: args{ctx: env.ctx,
				c:            c.Client,
				vmComparator: cmpName,
				attrValue:    TestImageName,
			},
			wantErr: false,
		},
		{
			name: "lookup by vm uuid",
			args: args{ctx: env.ctx,
				c:            c.Client,
				vmComparator: cmpUuid,
				attrValue:    TestVmUuid,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVmAttr(tt.args.ctx, tt.args.c, tt.args.vmComparator, tt.args.attrValue)
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

	client, ctx, teardown := setupTest(t)
	_ = teardown

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
			args{ctx,
				client.Client,
				TestImageName,
				"mgmt",
			},
			false,
		},
		{
			"empty names",
			args{ctx,
				client.Client,
				"", "",
			},
			true,
		},

		{
			"nil client",
			args{ctx,
				nil,
				"",
				"",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, vm, err := VmFromCluster(tt.args.ctx, tt.args.c, tt.args.vmName, tt.args.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("VmFromCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Println(vm)
		})
	}
}

func TestGetOpaqueNetwork(t *testing.T) {

	env, c := VimSetupHelper()

	type args struct {
		ctx        context.Context
		c          *vim25.Client
		folderName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Folder test",
			args{
				ctx:        env.ctx,
				c:          c.Client,
				folderName: "Test-61eb201c-420f-4c5a-b5f0-4f1dada5a015",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := GetOpaqueNetworks(tt.args.ctx, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteFolder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
