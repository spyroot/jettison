package test

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"

	"github.com/spyroot/jettison/nsxtapi"
)

func setupTest() (nsxt.APIClient, func()) {

	nsxtClient, err := nsxtapi.Connect(
		os.Getenv("NSXHOST"),
		os.Getenv("NSXUSERNAME"),
		os.Getenv("NSXPASSWORD"))

	if err != nil {
		log.Println("setupTest() error = ", err)
		log.Fatal("Failed to connect to nsx-t manager")
	}

	return nsxtClient, func() {
		log.Println("teardown test")
	}
}

var DhcpLoopCondition = map[string]func(dhcpEntry manager.DhcpStaticBinding, val string) bool{
	"mac": func(dhcpEntry manager.DhcpStaticBinding, macAddr string) bool {
		if dhcpEntry.MacAddress == macAddr {
			return true
		}
		return false
	},
	"ip": func(dhcpEntry manager.DhcpStaticBinding, ipAddr string) bool {
		if dhcpEntry.IpAddress == ipAddr {
			return true
		}
		return false
	},
}

func TestConnect(t *testing.T) {
	type args struct {
		managerHost string
		user        string
		password    string
	}
	tests := []struct {
		name    string
		args    args
		want    nsxt.APIClient
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.Connect(tt.args.managerHost, tt.args.user, tt.args.password)
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

//
func TestFindTransportZone(t *testing.T) {

	c, teardown := setupTest()
	defer teardown()

	type args struct {
		nsxClient     *nsxt.APIClient
		transportName string
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
		{
			"test valid id",
			args{
				&c,
				"overlay-trasport-zone",
			},
			"86622577-a94a-42f9-880c-80aa98a6e0ef",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := nsxtapi.FindTransportZone(tt.args.nsxClient, tt.args.transportName)
			log.Println(err != nil, tt.wantErr)
			if err != nil {
				if tt.wantErr {
					// if we want error we pass
					return
				} else {
					// if we dont want error but we got one, we failed.
					t.Errorf("TestFindTransportZone() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			_ = got
		})
	}
}
