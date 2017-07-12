package qcache_inventory

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

func NewContainer(id, name string, ips map[string]string) types.ContainerJSON {
	cbase :=  &types.ContainerJSONBase{
		ID: id,
		Name: name,
	}

	netConfig := &types.NetworkSettings{}
	netConfig.Networks = map[string]*network.EndpointSettings{}
	for iface, ip := range ips {
		endpoint := &network.EndpointSettings{
			IPAddress: ip,
		}
		netConfig.Networks[iface] =  endpoint
	}
	cnt := types.ContainerJSON{
		ContainerJSONBase: cbase,
		NetworkSettings: netConfig,
	}
	return cnt
}

func TestContainer_Equal(t *testing.T) {
	cnt := NewContainer("CntID1", "CntName1", map[string]string{"eth0": "172.17.0.2"})
	checkIP := NewIPContainerRequest("src1", "172.17.0.2")
	assert.True(t, checkIP.Equal(cnt))
	checkName := ContainerRequest{Name: "CntName1"}
	assert.True(t, checkName.Equal(cnt))
	checkID := ContainerRequest{ID: "CntID1"}
	assert.True(t, checkID.Equal(cnt))
}

func TestContainerRequest_TimedOut(t *testing.T) {
	req := ContainerRequest{
		Source: "src1",
		Name: "CntName1",
	}
	req.IssuedAt = time.Now().AddDate(0,0,-1)
	assert.True(t, req.TimedOut(), "Should be timed out long ago")
}