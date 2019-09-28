package test

import (
	"fmt"
	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/spyroot/jettison/nsxtapi"
	nsxt "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// helper
func createDhcpBindingRequest() *nsxtapi.DhcpBindingRequest {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	rr := r1.Intn(254)

	h := fmt.Sprintf("%02x", rr)

	req := &nsxtapi.DhcpBindingRequest{}
	req.ServerUuid = "459bdabc-9452-465e-b132-d452e8e2a266"
	req.Mac = "00:50:56:93:f4:" + h
	req.Ipaddr = "172.16.84." + strconv.FormatInt(int64(rr), 10)
	req.Hostname = "test"
	req.Gateway = "172.16.81.254"
	req.TenantId = "test"

	return req
}

func createDhcpProfile() {

	req := &nsxtapi.DhcpServerCreateReq{}

	// name for a server
	req.ServerName = ""
	req.DhcpServerIp = "172.16.22.22"
	req.DnsNameservers = []string{"8.8.8.8"}
	req.DomainName = "test"
	req.GatewayIp = "172.16.22.1"

	req.LogicalPortID = ""
	req.DhcpProfileId = ""
	req.ClusterId = ""
	req.SwitchId = ""
	req.TenantId = ""
	req.Segment = ""
}

func TestCreateDhcpProfile(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient   *nsxt.APIClient
		clusterId   string
		profileName string
		tags        []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test nil",
			args: args{&c,
				"133fe9a7-2e87-409a-b1b3-406ab5833986",
				"test-profile",
				nsxtapi.MakeDhcpProfileTag("test", "91c9a86e-20f2-410f-9561-55e5161c1842"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.CreateDhcpProfile(tt.args.nsxClient, tt.args.clusterId, tt.args.profileName, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDhcpProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				assert.NilError(t, err)
			}

			if err == nil {
				assert.NotNil(t, got)
				_, _, err := nsxtapi.DeleteDhcpProfile(tt.args.nsxClient, got)
				assert.NilError(t, err)
			}
		})
	}
}

func TestCreateDhcpServer(t *testing.T) {
	type args struct {
		nsxClient *nsxt.APIClient
		req       *nsxtapi.DhcpServerCreateReq
		tags      []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		want    *manager.LogicalDhcpServer
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.CreateDhcpServer(tt.args.nsxClient, tt.args.req, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDhcpServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateDhcpServer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateStaticBinding(t *testing.T) {
	type args struct {
		nsxClient *nsxt.APIClient
		serverId  string
		macaddr   string
		ipaddr    string
		hostname  string
		gateway   string
		tenantId  string
	}
	tests := []struct {
		name    string
		args    args
		want    *manager.DhcpStaticBinding
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.CreateStaticBinding(tt.args.nsxClient, tt.args.serverId, tt.args.macaddr, tt.args.ipaddr, tt.args.hostname, tt.args.gateway, tt.args.tenantId)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateStaticBinding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateStaticBinding() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStaticBinding(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient *nsxt.APIClient
		serverId  string
		Entry     string
		fn        nsxtapi.DhcpHandler
	}

	var dhcpEntrySuccess manager.DhcpStaticBinding

	tests := []struct {
		name    string
		args    args
		want    manager.DhcpStaticBinding
		wantErr bool
	}{
		// test passing null value for connector
		{
			"test",
			args{
				nil,
				"test",
				"test",
				DhcpLoopCondition["ip"]},
			dhcpEntrySuccess,
			true,
		},

		{
			"test",
			args{
				nil,
				"test",
				"test",
				DhcpLoopCondition["ip"]},
			dhcpEntrySuccess,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := nsxtapi.GetStaticBinding(&c, tt.args.serverId, tt.args.Entry, tt.args.fn)

			if (err != nil) != tt.wantErr {
				t.Log("Say bye")
				t.Errorf("GetStaticBinding() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStaticBinding() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteStaticBinding(t *testing.T) {
	type args struct {
		nsxClient *nsxt.APIClient
		serverId  string
		ipAddr    string
		macAddr   string
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
			if err := nsxtapi.DeleteStaticBinding(tt.args.nsxClient, tt.args.serverId, tt.args.ipAddr, tt.args.macAddr); (err != nil) != tt.wantErr {
				t.Errorf("DeleteStaticBinding() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDhcpServerCreateReq_ClusterUuid(t *testing.T) {
	type fields struct {
		ServerName     string
		DhcpServerIp   string
		DnsNameservers []string
		DomainName     string
		GatewayIp      string
		Options        string
		LogicalPortID  string
		DhcpProfileId  string
		ClusterId      string
		SwitchId       string
		TenantId       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &nsxtapi.DhcpServerCreateReq{
				ServerName:     tt.fields.ServerName,
				DhcpServerIp:   tt.fields.DhcpServerIp,
				DnsNameservers: tt.fields.DnsNameservers,
				DomainName:     tt.fields.DomainName,
				GatewayIp:      tt.fields.GatewayIp,
				Options:        tt.fields.Options,
				LogicalPortID:  tt.fields.LogicalPortID,
				DhcpProfileId:  tt.fields.DhcpProfileId,
				ClusterId:      tt.fields.ClusterId,
				SwitchId:       tt.fields.SwitchId,
				TenantId:       tt.fields.TenantId,
			}
			if got := d.ClusterUuid(); got != tt.want {
				t.Errorf("ClusterUuid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDhcpServerCreateReq_SwitchUuid(t *testing.T) {
	type fields struct {
		ServerName     string
		DhcpServerIp   string
		DnsNameservers []string
		DomainName     string
		GatewayIp      string
		Options        string
		LogicalPortID  string
		DhcpProfileId  string
		ClusterId      string
		SwitchId       string
		TenantId       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &nsxtapi.DhcpServerCreateReq{
				ServerName:     tt.fields.ServerName,
				DhcpServerIp:   tt.fields.DhcpServerIp,
				DnsNameservers: tt.fields.DnsNameservers,
				DomainName:     tt.fields.DomainName,
				GatewayIp:      tt.fields.GatewayIp,
				Options:        tt.fields.Options,
				LogicalPortID:  tt.fields.LogicalPortID,
				DhcpProfileId:  tt.fields.DhcpProfileId,
				ClusterId:      tt.fields.ClusterId,
				SwitchId:       tt.fields.SwitchId,
				TenantId:       tt.fields.TenantId,
			}
			if got := d.SwitchUuid(); got != tt.want {
				t.Errorf("SwitchUuid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDhcpServerCreateReq_TenantUuid(t *testing.T) {
	type fields struct {
		ServerName     string
		DhcpServerIp   string
		DnsNameservers []string
		DomainName     string
		GatewayIp      string
		Options        string
		LogicalPortID  string
		DhcpProfileId  string
		ClusterId      string
		SwitchId       string
		TenantId       string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &nsxtapi.DhcpServerCreateReq{
				ServerName:     tt.fields.ServerName,
				DhcpServerIp:   tt.fields.DhcpServerIp,
				DnsNameservers: tt.fields.DnsNameservers,
				DomainName:     tt.fields.DomainName,
				GatewayIp:      tt.fields.GatewayIp,
				Options:        tt.fields.Options,
				LogicalPortID:  tt.fields.LogicalPortID,
				DhcpProfileId:  tt.fields.DhcpProfileId,
				ClusterId:      tt.fields.ClusterId,
				SwitchId:       tt.fields.SwitchId,
				TenantId:       tt.fields.TenantId,
			}
			if got := d.TenantUuid(); got != tt.want {
				t.Errorf("TenantUuid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindAttachedDhcpServerProfile(t *testing.T) {
	type args struct {
		nsxClient       *nsxt.APIClient
		logicalSwitchID string
	}
	tests := []struct {
		name    string
		args    args
		want    *manager.LogicalDhcpServer
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindAttachedDhcpServerProfile(tt.args.nsxClient, tt.args.logicalSwitchID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAttachedDhcpServerProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAttachedDhcpServerProfile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindDhcpServer(t *testing.T) {
	type args struct {
		nsxClient   *nsxt.APIClient
		searchValue string
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
			got, err := nsxtapi.FindDhcpServer(tt.args.nsxClient, tt.args.searchValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindDhcpServer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindDhcpServer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindDhcpServerByTag(t *testing.T) {
	type args struct {
		nsxClient *nsxt.APIClient
		tags      []common.Tag
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
			got, err := nsxtapi.FindDhcpServerByTag(tt.args.nsxClient, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindDhcpServerByTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindDhcpServerByTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindDhcpServerProfileByTag(t *testing.T) {
	type args struct {
		nsxClient *nsxt.APIClient
		tags      []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		want    *manager.DhcpProfile
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindDhcpServerProfileByTag(tt.args.nsxClient, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindDhcpServerProfileByTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindDhcpServerProfileByTag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//
func TestFindDhcpServerProfile(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient       *nsxt.APIClient
		logicalSwitchId string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// test passing null value for connector
		{
			"TestFindDhcpServerProfile nil connector",
			args{
				nil,
				""},
			"",
			true,
		},
		// test should pass with error since logicalSwitch empty
		{
			"TestFindDhcpServerProfile nil connector",
			args{
				&c,
				""},
			"",
			true,
		},
		// test should pass with error since logicalSwitch empty
		{
			"test3",
			args{
				&c,
				"91c9a86e-20f2-410f-9561-55e5161c1842",
			},
			"86622577-a94a-42f9-880c-80aa98a6e0ef",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindAttachedDhcpServerProfile(tt.args.nsxClient, tt.args.logicalSwitchId)
			log.Println(err != nil, tt.wantErr)
			if err != nil {
				if tt.wantErr {
					// if we want error we pass
					return
				} else {
					// if we dont want error but we got one, we failed.
					t.Errorf("FindDhcpServerProfile() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if got.Id != tt.want {
				t.Errorf("FindDhcpServerProfile() got = %v, want %v", got, tt.want)
			} else {
				t.Log(got.Id, tt.want)
			}
		})
	}
}

//
func TestGetStaticBindingByTag(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient *nsxt.APIClient
		dhcpUuid  string
		tags      []common.Tag
	}

	_, _ = nsxtapi.CreateStaticReq(&c, createDhcpBindingRequest())

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// test passing null value for connector
		{
			"TestGetStaticBindingByTag nil connector",
			args{
				nil,
				"", nil},
			true,
		},
		// test should pass with error since logicalSwitch empty
		{
			"TestGetStaticBindingByTag empty request",
			args{
				&c,
				"", nil},
			true,
		},
		// test should pass with error since logicalSwitch empty
		{
			"valid request",
			args{
				&c,
				"459bdabc-9452-465e-b132-d452e8e2a266", // test need dhcp define
				nsxtapi.MakeCustomTags("jettison", "SuperCluster"),
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.GetStaticBindingByTag(tt.args.nsxClient, tt.args.dhcpUuid, tt.args.tags)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetStaticBindingByTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				assert.NotNil(t, got)
				for _, v := range got {
					t.Log(v.DisplayName)
				}
			}

			if got != nil {
				assert.NilError(t, err)
			}
		})
	}
}
