package test

import (
	"github.com/spyroot/jettison/nsxtapi"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"net"
	"testing"
)

/**
  Test create / delete invariant
*/
func TestCreateLogicalRouter(t *testing.T) {

	nsxClient, teardown := setupTest()
	defer teardown()

	clusters, err := nsxtapi.FindEdgeCluster(&nsxClient, nsxtapi.EdgeClusterCallback["name"], "edge-cluster")
	if err != nil {
		t.Fatal("No cluster defined")
	}

	type args struct {
		nsxClient   *nsxt.APIClient
		routerType  string
		name        string
		clusterUuid string
		tags        []common.Tag
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"test create tier0",
			args{
				&nsxClient,
				"TIER0",
				"test",
				clusters[0].Id,
				nil,
			},
			false,
		},
		{
			"test create tier1",
			args{
				&nsxClient,
				"TIER1",
				"test",
				clusters[0].Id,
				nil,
			},
			false,
		},
		{
			"wrong type",
			args{
				&nsxClient,
				"TIER2",
				"test",
				clusters[0].Id,
				nil,
			},
			true,
		},
		{
			"wrong cluster id",
			args{
				&nsxClient,
				"TIER0",
				"test",
				"1234",
				nil,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := nsxtapi.RouterCreateReq{}

			req.Name = tt.args.name
			req.RouterType = tt.args.routerType
			req.ClusterID = tt.args.clusterUuid
			req.Tags = tt.args.tags

			nsx := tt.args.nsxClient
			got, err := nsxtapi.CreateLogicalRouter(nsx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLogicalRouter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.wantErr == true {
				t.Log(" Expected error and go error")
				return
			}

			if err == nil && got == nil {
				t.Errorf("CreateLogicalRouter() error not nil but got back nil")
				return
			}

			if got != nil {

				// read / delete / read after addition
				_, _, err = tt.args.nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsx.Context, got.Id)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateLogicalRouter() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				_, err = nsxtapi.DeleteLogicalRouter(tt.args.nsxClient, got.Id)
				if (err != nil) != tt.wantErr {
					t.Errorf("CreateLogicalRouter() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				_, _, err = tt.args.nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsx.Context, got.Id)
				if err == nil {
					t.Errorf("Delete failed")
					return
				}
			}
		})
	}
}

func TestDeleteLogicalRouter(t *testing.T) {

	nsxClient, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient  *nsxt.APIClient
		routerName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"bogus name",
			args{
				&nsxClient,
				"bogus name",
			},
			true,
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := nsxtapi.DeleteLogicalRouter(tt.args.nsxClient, tt.args.routerName); (err != nil) != tt.wantErr {
				t.Errorf("DeleteLogicalRouter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

/*
  In order to pass this test you need have router defined.
*/
func TestFindLogicalRouter(t *testing.T) {

	nsxClient, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient *nsxt.APIClient
		callback  nsxtapi.RouterSearchHandler
		searchVal string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should return not found",
			args{
				&nsxClient,
				nsxtapi.RouterCallback["name"],
				"",
			},
			true,
		},
		{
			"test search by router name",
			args{
				&nsxClient,
				nsxtapi.RouterCallback["name"],
				"primary-t0",
			},
			false,
		},
		{
			"test search by uuid name",
			args{
				&nsxClient,
				nsxtapi.RouterCallback["uuid"],
				"ba95b780-3689-419b-8f20-c7179e05813f",
			},
			false,
		},
		{
			"test search by edge id",
			args{
				&nsxClient,
				nsxtapi.RouterCallback["edgeid"],
				"133fe9a7-2e87-409a-b1b3-406ab5833986",
			},
			false,
		},
		{
			"test search by type ",
			args{
				&nsxClient,
				nsxtapi.RouterCallback["type"],
				"TIER1",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindLogicalRouter(tt.args.nsxClient, tt.args.callback, tt.args.searchVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindLogicalRouter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, v := range got {
				if err == nil {
					t.Log("Found router id ", v.Id, " name", v.DisplayName, " type ", v.RouterType, "edge id", v.EdgeClusterId)
				}
			}
		})
	}
}

/*
  In order to pass this test you need have cluster must be defined.
*/
func TestFindEdgeCluster(t *testing.T) {

	nsxClient, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient *nsxt.APIClient
		callback  nsxtapi.EdgeSearchHandler
		searchVal string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should return not found",
			args{
				&nsxClient,
				nsxtapi.EdgeClusterCallback["name"],
				"",
			},
			true,
		},
		{
			"test search by cluster name",
			args{
				&nsxClient,
				nsxtapi.EdgeClusterCallback["name"],
				"edge-cluster",
			},
			false,
		},
		{
			"test search by cluster uuid",
			args{
				&nsxClient,
				nsxtapi.EdgeClusterCallback["uuid"],
				"133fe9a7-2e87-409a-b1b3-406ab5833986",
			},
			false,
		},
		{
			"test search by cluster uuid not found",
			args{
				&nsxClient,
				nsxtapi.EdgeClusterCallback["uuid"],
				"133fe9a7",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindEdgeCluster(tt.args.nsxClient, tt.args.callback, tt.args.searchVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindEdgeCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, v := range got {
				if err == nil {
					t.Log("found router id ", v.Id, " name", v.DisplayName)
				}
			}
		})
	}
}

//
// Add static routes
//
func TestAddStaticRoute(t *testing.T) {

	nsxClient, teardown := setupTest()
	defer teardown()

	// TODO PASS cluster name via ENV
	clusters, err := nsxtapi.FindEdgeCluster(&nsxClient, nsxtapi.EdgeClusterCallback["name"], "edge-cluster")
	if err != nil {
		t.Fatal("No cluster defined")
	}

	if len(clusters) == 0 {
		t.Fatal("No cluster defined")
	}

	cluster := clusters[0]

	type args struct {
		nsxClient *nsxt.APIClient
		network   string
		nexthop   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should return error",
			args{
				&nsxClient,
				"",
				"",
			},
			true,
		},
		{
			"invalid next hop",
			args{
				&nsxClient,
				"10.20.0.0/24",
				"",
			},
			true,
		},
		{
			"invalid next hop",
			args{
				&nsxClient,
				"10.20.0.0/24",
				"172.16.88.1",
			},
			false,
		},
	}

	for _, tt := range tests {

		//add router
		req := nsxtapi.RouterCreateReq{}
		req.Name = "test"
		req.RouterType = nsxtapi.RouteTypeTier1
		req.ClusterID = cluster.Id
		got, err := nsxtapi.CreateLogicalRouter(tt.args.nsxClient, req)
		assert.Nil(t, err)

		t.Run(tt.name, func(t *testing.T) {

			assert.NotNil(t, got)

			req2 := nsxtapi.AddStaticReq{}
			req2.RouterUuid = got.Id
			req2.Network = tt.args.network
			//			req.PortUuid =
			req2.NextHopAddr = net.ParseIP(tt.args.nexthop)

			_, err = nsxtapi.AddStaticRoute(tt.args.nsxClient, req2)
			if (err != nil) != tt.wantErr {
				t.Errorf("TestAddStaticRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})

		// delete
		_, err = nsxtapi.DeleteLogicalRouter(tt.args.nsxClient, got.Id)
		assert.Nil(t, err)
	}
}
