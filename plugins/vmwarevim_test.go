package main

import (
	"context"
	"github.com/spyroot/jettison/jettypes"
	"github.com/spyroot/jettison/vcenter"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"reflect"
	"testing"
)

func setupTest(t *testing.T) (*vcenter.TestingEnv, func(t *testing.T)) {

	vimHelper, _, err := vcenter.VimSetupHelper()
	if err != nil {
		log.Fatal("Failed VimSetupHelper()", err)
	}

	return vimHelper, func(t *testing.T) {
		t.Log("teardown test")
	}
}

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		want    jettypes.VimPlugin
		wantErr bool
	}{
		{
			name:    "empty types",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Init()
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func TestVmwareVim_ComputeCleanup(t *testing.T) {
	type fields struct {
		VimApi     *govmomi.Client
		VimFinder  *find.Finder
		datacenter *object.Datacenter
		ctx        context.Context
		nsxApi     nsxt.APIClient
		nsxtConfig *NsxtConfig
	}
	type args struct {
		projectName string
		nodes       *[]*jettypes.NodeTemplate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &VmwareVim{
				vimApi:     tt.fields.VimApi,
				VimFinder:  tt.fields.VimFinder,
				datacenter: tt.fields.datacenter,
				ctx:        tt.fields.ctx,
				nsxApi:     tt.fields.nsxApi,
				nsxtConfig: tt.fields.nsxtConfig,
			}
			if err := p.ComputeCleanup(tt.args.projectName, tt.args.nodes); (err != nil) != tt.wantErr {
				t.Errorf("ComputeCleanup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVmwareVim_CreateVms(t *testing.T) {

	type fields struct {
		VimApi     *govmomi.Client
		VimFinder  *find.Finder
		datacenter *object.Datacenter
		ctx        context.Context
		nsxApi     nsxt.APIClient
		nsxtConfig *NsxtConfig
	}
	type args struct {
		projectName string
		nodes       []*jettypes.NodeTemplate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &VmwareVim{
				vimApi:     tt.fields.VimApi,
				VimFinder:  tt.fields.VimFinder,
				datacenter: tt.fields.datacenter,
				ctx:        tt.fields.ctx,
				nsxApi:     tt.fields.nsxApi,
				nsxtConfig: tt.fields.nsxtConfig,
			}
			if err := p.CreateVms(tt.args.projectName, tt.args.nodes); (err != nil) != tt.wantErr {
				t.Errorf("CreateVms() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVmwareVim_GetNsx(t *testing.T) {

	env, teardown := setupTest(t)
	defer teardown(t)

	type args struct {
		client *nsxt.APIClient
		vim    *VmwareVim
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				env.TestVim.GetNsx(),
				env.TestVim,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.args.vim.GetNsx())
		})
	}
}

func TestVmwareVim_InitPlugin(t *testing.T) {

	type args struct {
		vimEndpoint jettypes.VimEndpoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "init with nothing",
			args:    args{nil},
			wantErr: true,
		},
		{
			name:    "valid credentials",
			args:    args{&vcenter.MinVimConfig{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := &VmwareVim{}
			err := p.InitPlugin(tt.args.vimEndpoint)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				assert.NotNil(t, p)
				assert.NotNil(t, p.GetNsx())
				assert.Contains(t, p.VimClient().Version, ".")
				assert.NotNil(t, p.VimClient())
				assert.NotNil(t, p.nsxtConfig)
				assert.NotNil(t, p.VimFinder)
				assert.NotNil(t, p.ctx)
			}
		})
	}
}

func TestVmwareVim_VimClient(t *testing.T) {

	env, teardown := setupTest(t)
	defer teardown(t)

	type args struct {
		client *vim25.Client
		vim    *VmwareVim
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				env.TestVim.VimClient(),
				env.TestVim,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//
			//got := tt.args
			//
			//if got := p.VimClient(); !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("VimClient() = %v, want %v", got, tt.want)
			//}
		})
	}
}

func TestVmwareVim_discoverNetwork(t *testing.T) {

	env, teardown := setupTest(t)
	defer teardown(t)

	type args struct {
		vim *VmwareVim
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "discovery",
			args: args{
				env.TestVim,
			},
			wantErr: false,
		},
		{
			name: "should fail with empty tz",
			args: args{
				func() *VmwareVim {
					// reset overlay and discover
					e, _ := setupTest(t)
					e.TestVim.nsxtConfig.SetOverlayTzName("")
					return e.TestVim
				}(),
			},
			wantErr: true,
		},
		{
			name: "should fail with empty edge",
			args: args{
				func() *VmwareVim {
					// reset overlay and discover
					e, _ := setupTest(t)
					e.TestVim.nsxtConfig.SetEdgeCluster("")
					return e.TestVim
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong transport name",
			args: args{
				func() *VmwareVim {
					// reset overlay and discover
					e, _ := setupTest(t)
					e.TestVim.nsxtConfig.SetOverlayTzName("test123")
					return e.TestVim
				}(),
			},
			wantErr: true,
		},
		{
			name: "wrong edge transport name",
			args: args{
				func() *VmwareVim {
					// reset overlay and discover
					e, _ := setupTest(t)
					e.TestVim.nsxtConfig.SetEdgeCluster("test123")
					return e.TestVim
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.vim.discoverNetwork(); (err != nil) != tt.wantErr {
				t.Errorf("discoverNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestVmwareVim_findTemplate(t *testing.T) {

	//env, teardown := setupTest(t)
	//defer teardown(t)

	type args struct {
		vim  *VmwareVim
		node *jettypes.NodeTemplate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should fail with nil arg",
			args: args{
				func() *VmwareVim {
					e, _ := setupTest(t)
					return e.TestVim
				}(), nil,
			},
			wantErr: true,
		},
		{
			name: "should set correct vim uuid",
			args: args{
				func() *VmwareVim {
					e, _ := setupTest(t)
					return e.TestVim
				}(), vcenter.TestTemplate(),
			},
			wantErr: false,
		},
		{
			name: "should not find it",
			args: args{
				func() *VmwareVim {
					// reset overlay and discover
					e, _ := setupTest(t)
					//					e.TestVim.nsxtConfig.SetOverlayTzName("")
					return e.TestVim
				}(), vcenter.TestTemplateBogus01(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, got1, err := tt.args.vim.findTemplate(tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("findTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
				assert.Nil(t, got1)
			}

			if err == nil {
				assert.NotNil(t, got)
				assert.NotNil(t, got1)

				t.Log(tt.args.node.UUID)
			}

			//if !reflect.DeepEqual(got, tt.want) {
			//	t.Errorf("findTemplate() got = %v, want %v", got, tt.want)
			//}
			//if !reflect.DeepEqual(got1, tt.want1) {
			//	t.Errorf("findTemplate() got1 = %v, want %v", got1, tt.want1)
			//}

		})
	}
}

func TestVmwareVim_findVmObject(t *testing.T) {
	type fields struct {
		VimApi     *govmomi.Client
		VimFinder  *find.Finder
		datacenter *object.Datacenter
		ctx        context.Context
		nsxApi     nsxt.APIClient
		nsxtConfig *NsxtConfig
	}
	type args struct {
		vmName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.VirtualMachine
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &VmwareVim{
				vimApi:     tt.fields.VimApi,
				VimFinder:  tt.fields.VimFinder,
				datacenter: tt.fields.datacenter,
				ctx:        tt.fields.ctx,
				nsxApi:     tt.fields.nsxApi,
				nsxtConfig: tt.fields.nsxtConfig,
			}
			got, err := p.findVmObject(tt.args.vmName)
			if (err != nil) != tt.wantErr {
				t.Errorf("findVmObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findVmObject() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVmwareVim_getTemplateData(t *testing.T) {
	type fields struct {
		VimApi     *govmomi.Client
		VimFinder  *find.Finder
		datacenter *object.Datacenter
		ctx        context.Context
		nsxApi     nsxt.APIClient
		nsxtConfig *NsxtConfig
	}
	type args struct {
		ctx  context.Context
		node *jettypes.NodeTemplate
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *object.VirtualMachine
		want1   *[]mo.Network
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &VmwareVim{
				vimApi:     tt.fields.VimApi,
				VimFinder:  tt.fields.VimFinder,
				datacenter: tt.fields.datacenter,
				ctx:        tt.fields.ctx,
				nsxApi:     tt.fields.nsxApi,
				nsxtConfig: tt.fields.nsxtConfig,
			}
			got, got1, err := p.getTemplateData(tt.args.ctx, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("getTemplateData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTemplateData() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getTemplateData() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestVmwareVim_isAttached(t *testing.T) {
	type fields struct {
		VimApi     *govmomi.Client
		VimFinder  *find.Finder
		datacenter *object.Datacenter
		ctx        context.Context
		nsxApi     nsxt.APIClient
		nsxtConfig *NsxtConfig
	}
	type args struct {
		vmName     string
		switchUuid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &VmwareVim{
				vimApi:     tt.fields.VimApi,
				VimFinder:  tt.fields.VimFinder,
				datacenter: tt.fields.datacenter,
				ctx:        tt.fields.ctx,
				nsxApi:     tt.fields.nsxApi,
				nsxtConfig: tt.fields.nsxtConfig,
			}
			got, err := p.isAttached(tt.args.vmName, tt.args.switchUuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("isAttached() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isAttached() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkNetworkAttachment(t *testing.T) {
	type args struct {
		networks   *[]mo.Network
		switchName string
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
			if got := checkNetworkAttachment(tt.args.networks, tt.args.switchName); got != tt.want {
				t.Errorf("checkNetworkAttachment() = %v, want %v", got, tt.want)
			}
		})
	}
}
