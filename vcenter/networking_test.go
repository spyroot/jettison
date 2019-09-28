package vcenter

import (
	"context"
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25"
	"log"
	"net"
	"os"
	"reflect"
	"testing"
)

func setupTest(t *testing.T) (*govmomi.Client, context.Context, func(t *testing.T)) {

	ctx := context.Background()
	vsphereClient, err := Connect(ctx,
		os.Getenv("VCENTER_HOSTNAME"),
		os.Getenv("VCENTER_USERNAME"),
		os.Getenv("VCENTER_PASSWORD"))

	if err != nil {
		log.Fatal("Failed VimSetupHelper()", err)
	}

	return vsphereClient, ctx, func(t *testing.T) {
		t.Log("teardown test")
	}
}

func TestAddNetworkAdapter(t *testing.T) {

	client, ctx, teardown := setupTest(t)
	_ = teardown

	type args struct {
		ctx         context.Context
		c           *vim25.Client
		networkName string
		vmName      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "valid nsxt-t uuid and vm name",
			args:    args{ctx, client.Client, "91c9a86e-20f2-410f-9561-55e5161c1842", "jettison-test"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, nics, err := AddNetworkAdapter(tt.args.ctx, tt.args.c, tt.args.networkName, tt.args.vmName)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNetworkAdapter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
				assert.NotNil(t, nics)
			}
			nicList, err := NetworkAdapterList(tt.args.ctx, tt.args.c, tt.args.vmName)
			if err != nil {
				t.Errorf("failed accquire nic list")
				return
			}

			found := false
			for _, v := range nicList {
				m, err := net.ParseMAC(v.MacAddress)
				if err != nil {
					t.Errorf("failed parse mac address")
					return
				}
				if m.String() == got {
					found = true
				}
			}

			t.Log(got)
			assert.Equal(t, found, true)
		})
	}
}

func TestGetNetworkAttr(t *testing.T) {

	client, ctx, teardown := setupTest(t)
	_ = teardown

	vm, err := GetVmAttr(ctx, client.Client, VmSearchHandler["name"], "jettison-test")
	if err != nil {
		t.Errorf("Failed to find vm in order execute a test")
		return
	}

	type args struct {
		ctx       context.Context
		c         *vim25.Client
		vmVimName string
		vmName    string
		attach    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid nsxt-t uuid and vm name",
			args: args{ctx,
				client.Client,
				vm.Summary.Vm.Value,
				"jettison-test",
				"57f91737-2bc6-41b7-9b78-34d595135043"}, // we choose network that vm not attached
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := GetNetworkAttr(tt.args.ctx, tt.args.c, tt.args.vmVimName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetworkAttr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
				oldNumNetworks := len(*got)
				// add adapter to get network again
				_, _, err = AddNetworkAdapter(tt.args.ctx, tt.args.c, tt.args.attach, tt.args.vmName)
				if err != nil {
					t.Errorf("failed to add adapter")
					return
				}

				newGot, _ := GetNetworkAttr(tt.args.ctx, tt.args.c, tt.args.vmVimName)
				assert.Equal(t, oldNumNetworks+1, len(*newGot))

				//				DeleteNetworkAdapter(tt.args.ctx, tt.args.c, "57f91737-2bc6-41b7-9b78-34d595135043", "jettison-test")
			}
		})
	}
}

func TestGetNetworkRef(t *testing.T) {
	type args struct {
		ctx  context.Context
		c    *vim25.Client
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNetworkRef(tt.args.ctx, tt.args.c, tt.args.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNetworkRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNetworkRef() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOpaqueByUuid(t *testing.T) {

	client, ctx, teardown := setupTest(t)
	_ = teardown

	type args struct {
		ctx        context.Context
		c          *vim25.Client
		switchUuid string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			args:    args{},
			wantErr: true,
		},
		{
			name:    "empty uuid",
			args:    args{ctx, client.Client, ""},
			wantErr: true,
		},
		{
			name:    "valid nsxt-t uuid",
			args:    args{ctx, client.Client, "91c9a86e-20f2-410f-9561-55e5161c1842"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOpaqueByUuid(tt.args.ctx, tt.args.c, tt.args.switchUuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOpaqueByUuid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestGetOpaqueNetworks(t *testing.T) {

	client, ctx, teardown := setupTest(t)
	_ = teardown

	type args struct {
		ctx context.Context
		c   *vim25.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nil",
			args:    args{},
			wantErr: true,
		},
		{
			name:    "valid request",
			args:    args{ctx, client.Client},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOpaqueNetworks(tt.args.ctx, tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOpaqueNetworks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
				for k, v := range got {
					if !nsxtapi.IsUuid(k) {
						t.Errorf("GetOpaqueNetworks() map contains invalid uuid")
					}

					ok, err := isOpaqueNetwork(tt.args.ctx, tt.args.c, v)
					assert.Nil(t, err)
					assert.Equal(t, true, ok)
				}
			}
		})
	}
}

func TestGetSwitchUuid(t *testing.T) {
	type args struct {
		ctx    context.Context
		c      *vim25.Client
		vmName string
	}
	tests := []struct {
		name    string
		args    args
		want    []NetworkAdapter
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSwitchUuid(tt.args.ctx, tt.args.c, tt.args.vmName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSwitchUuid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSwitchUuid() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isOpaqueNetwork(t *testing.T) {

	client, ctx, teardown := setupTest(t)
	_ = teardown

	type args struct {
		ctx        context.Context
		c          *vim25.Client
		switchName string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "valid request",
			args:    args{ctx, client.Client, "test-segment"},
			wantErr: false,
		},
		{
			name:    "invalid request",
			args:    args{ctx, client.Client, "test-segment123"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isOpaqueNetwork(tt.args.ctx, tt.args.c, tt.args.switchName)
			if (err != nil) != tt.wantErr {
				t.Errorf("isOpaqueNetwork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.Nil(t, got)
			}

			if err == nil {
				assert.NotNil(t, got)
			}
		})
	}
}
