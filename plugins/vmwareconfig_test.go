package main

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/go-vmware-nsxt/manager"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_NewNsxtConfig(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create new",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNsxtConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNsxtConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_ReadConfig(t *testing.T) {

	tests := []struct {
		name    string
		args    io.Reader
		wantErr bool
	}{
		{
			name:    "should fail",
			args:    strings.NewReader("test"),
			wantErr: true,
		},
		{
			name:    "should fail empty string",
			args:    strings.NewReader(""),
			wantErr: true,
		},
		{
			name: "should pass",
			args: strings.NewReader(`
nsxt:
    hostname: 172.16.254.205
    username: admin
    password: "VMware1!"
    logicalSwitch: "test-segment"
    edgeCluster: "edge-cluster"
    overlayTransport:  "overlay-trasport-zone"
`),
			wantErr: false,
		},
		{
			name: "should pass",
			args: strings.NewReader(`
nsxt:
    hostname: 172.16.254.205
    username: admin
    password: "VMware1!"
    overlayTransport:  "overlay-trasport-zone"
`),
			wantErr: false,
		},
		{
			name: "should pass",
			args: strings.NewReader(`
nsxt:
    hostname: 172.16.254.205
    username: admin
    password: "VMware1!"
    overlayTransport:  "overlay-trasport-zone"
`),
			wantErr: false,
		},
		{
			name: "no username",
			args: strings.NewReader(`
nsxt:
    hostname: 172.16.254.205
    password: "VMware1!"
    overlayTransport:  "overlay-trasport-zone"
`),
			wantErr: true,
		},
		{
			name: "no password",
			args: strings.NewReader(`
nsxt:
    hostname: 172.16.254.205
    username: Admin
    overlayTransport:  "overlay-trasport-zone"
`),
			wantErr: true,
		},
		{
			name: "no hostname",
			args: strings.NewReader(`
nsxt:
    username: admin
    password: "VMware1!"
    overlayTransport:  "overlay-trasport-zone"
`),
			wantErr: true,
		},
		{
			name:    "nil reader",
			args:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadConfig(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Test_ReadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				assert.Equal(t, got.Hostname(), "172.16.254.205")
				assert.Equal(t, got.Username(), "admin")
				assert.Equal(t, got.Password(), "VMware1!")
				assert.Equal(t, got.OverlayTransportName(), "overlay-trasport-zone")

				got.SetOverlayTzName("test")
				assert.Equal(t, got.OverlayTransportName(), "test")

				got.SetOverlayTzId("12345")
				assert.Equal(t, got.OverlayTransportUuid(), "12345")

				got.SetEdgeClusterUuid("123456789")
				assert.Equal(t, got.EdgeClusterUuid(), "123456789")

				got.SetEdgeCluster("edge")
				assert.Equal(t, got.EdgeCluster(), "edge")

				got.SetDhcpUuid("dhcpid")
				assert.Equal(t, got.DhcpServerUuid(), "dhcpid")

				r1 := &manager.LogicalRouter{}
				r1.DisplayName = "test123"
				r1.Id = "1"

				r2 := &manager.LogicalRouter{}
				r2.DisplayName = "test123"
				r2.Id = "2"

				got.AddTierZero(r1)
				got.AddTierZero(r2)

				assert.Equal(t, 2, len(got.TierZero()))

				r11 := &manager.LogicalRouter{}
				r11.DisplayName = "test1234"
				r11.Id = "1"

				r22 := &manager.LogicalRouter{}
				r22.DisplayName = "test12356"
				r22.Id = "2"

				got.AddTierOne(r1)
				got.AddTierOne(r2)

				assert.Equal(t, 2, len(got.TierOne()))

			}
		})
	}
}

func Test_ReadFromFile(t *testing.T) {
	tests := []struct {
		name    string
		want    *os.File
		want1   *bufio.Reader
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {

		t.Log("Testing")
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ReadFromFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadFromFile() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ReadFromFile() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
