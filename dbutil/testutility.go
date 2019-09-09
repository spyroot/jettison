package dbutil

import (
	"github.com/google/uuid"
	"github.com/spyroot/jettison/config"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"

	"math/rand"
	"net"
	"time"
)

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func CreateSyntheticValidNode(t *testing.T) *config.NodeTemplate {

	var node0 = &config.NodeTemplate{}

	node0.Name = uuid.New().String()
	node0.DesiredAddress = "172.16.81.0/24"
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.IPv4Addr = net.ParseIP("172.16.81.1")
	node0.VimCluster = uuid.New().String()
	node0.VmTemplateName = uuid.New().String()
	node0.Mac = append(node0.Mac, "abcd:abcd:abc")

	VimName := uuid.New().String()
	node0.SetVimName(VimName)
	assert.Equal(t, VimName, node0.GetVimName(), "switch uuid mismatch")

	node0.SetFolderPath("/Datacenter/vms")
	assert.Equal(t, "/Datacenter/vms", node0.GetFolderPath(), "folder mismatch")

	switchUuid := uuid.New().String()
	node0.SetSwitchUuid(switchUuid)
	assert.Equal(t, switchUuid, node0.GetSwitchUuid(), "switch uuid mismatch")
	log.Println(switchUuid)

	routerUuid := uuid.New().String()
	node0.SetRouterUuid(routerUuid)
	assert.Equal(t, routerUuid, node0.GetRouterUuid(), "router uuid mismatch")

	dhcpID := uuid.New().String()
	node0.SetDhcpId(dhcpID)
	assert.Equal(t, dhcpID, node0.GetDhcpId(), "dhcp id uuid mismatch")

	node0.VimCluster = "mgmt"

	node0.Type = config.ControlType

	return node0
}

func CreateSyntheticValidNodes(t *testing.T, numNodes int) *[]*config.NodeTemplate {

	var nodes []*config.NodeTemplate
	for i := 0; i < numNodes; i++ {
		n := CreateSyntheticValidNode(t)
		nodes = append(nodes, n)
	}

	return &nodes
}
