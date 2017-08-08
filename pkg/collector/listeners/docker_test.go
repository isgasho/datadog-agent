package listeners

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigIDFromPs(t *testing.T) {
	co := types.Container{
		ID:    "deadbeef",
		Image: "test",
	}
	dl := DockerListener{}

	id := dl.getConfigIDFromPs(co)
	assert.Equal(t, "test", id)

	labeledCo := types.Container{
		ID:     "deadbeef",
		Image:  "test",
		Labels: map[string]string{"io.datadog.check.id": "w00tw00t"},
	}
	id = dl.getConfigIDFromPs(labeledCo)
	assert.Equal(t, "w00tw00t", id)
}

func TestGetHostsFromPs(t *testing.T) {
	dl := DockerListener{}

	co := types.Container{
		ID:    "foo",
		Image: "test",
	}

	assert.Empty(t, dl.getHostsFromPs(co))

	nets := make(map[string]*network.EndpointSettings)
	nets["bridge"] = &network.EndpointSettings{IPAddress: "172.17.0.2"}
	nets["foo"] = &network.EndpointSettings{IPAddress: "172.17.0.3"}
	networkSettings := types.SummaryNetworkSettings{
		Networks: nets}

	co = types.Container{
		ID:              "deadbeef",
		Image:           "test",
		NetworkSettings: &networkSettings,
		Ports:           []types.Port{types.Port{PrivatePort: 1337}, types.Port{PrivatePort: 42}},
	}
	hosts := dl.getHostsFromPs(co)

	assert.Equal(t, "172.17.0.2", hosts["bridge"])
	assert.Equal(t, "172.17.0.3", hosts["foo"])
	assert.Equal(t, 2, len(hosts))
}

func TestGetPortsFromPs(t *testing.T) {
	dl := DockerListener{}

	co := types.Container{
		ID:    "foo",
		Image: "test",
	}
	assert.Empty(t, dl.getPortsFromPs(co))

	co.Ports = make([]types.Port, 0)
	assert.Empty(t, dl.getPortsFromPs(co))

	co.Ports = append(co.Ports, types.Port{PrivatePort: 1234})
	co.Ports = append(co.Ports, types.Port{PrivatePort: 4321})
	ports := dl.getPortsFromPs(co)
	assert.Equal(t, 2, len(ports))
	assert.Contains(t, ports, 1234)
	assert.Contains(t, ports, 4321)
}

func TestGetConfigIDFromInspect(t *testing.T) {
	co := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{ID: "deadbeef", Image: "test"},
		Mounts:            make([]types.MountPoint, 0),
		Config:            &container.Config{},
		NetworkSettings:   &types.NetworkSettings{},
	}
	dl := DockerListener{}

	id := dl.getConfigIDFromInspect(co)
	assert.Equal(t, "test", string(id))

	labeledCo := types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{ID: "deadbeef", Image: "test"},
		Mounts:            make([]types.MountPoint, 0),
		Config:            &container.Config{Labels: map[string]string{"io.datadog.check.id": "w00tw00t"}},
		NetworkSettings:   &types.NetworkSettings{},
	}
	id = dl.getConfigIDFromInspect(labeledCo)
	assert.Equal(t, "w00tw00t", string(id))
}

func TestGetHostsFromInspect(t *testing.T) {
	dl := DockerListener{}

	cBase := types.ContainerJSONBase{
		ID:    "foo",
		Image: "test",
	}
	co := types.ContainerJSON{
		ContainerJSONBase: &cBase,
		Mounts:            make([]types.MountPoint, 0),
		Config:            &container.Config{Labels: map[string]string{"io.datadog.check.id": "w00tw00t"}},
		NetworkSettings:   &types.NetworkSettings{},
	}

	assert.Empty(t, dl.getHostsFromInspect(co))

	nets := make(map[string]*network.EndpointSettings)
	nets["bridge"] = &network.EndpointSettings{IPAddress: "172.17.0.2"}
	nets["foo"] = &network.EndpointSettings{IPAddress: "172.17.0.3"}
	ports := make(nat.PortMap)
	p, _ := nat.NewPort("tcp", "1337")
	ports[p] = make([]nat.PortBinding, 0)
	p, _ = nat.NewPort("tcp", "42")
	ports[p] = make([]nat.PortBinding, 0)

	cBase = types.ContainerJSONBase{
		ID:    "deadbeef",
		Image: "test",
	}
	networkSettings := types.NetworkSettings{
		NetworkSettingsBase: types.NetworkSettingsBase{Ports: ports},
		Networks:            nets,
	}

	co = types.ContainerJSON{
		ContainerJSONBase: &cBase,
		Mounts:            make([]types.MountPoint, 0),
		Config:            &container.Config{},
		NetworkSettings:   &networkSettings,
	}
	hosts := dl.getHostsFromInspect(co)

	assert.Equal(t, "172.17.0.2", hosts["bridge"])
	assert.Equal(t, "172.17.0.3", hosts["foo"])
	assert.Equal(t, 2, len(hosts))
}

func TestGetPortsFromInspect(t *testing.T) {
	dl := DockerListener{}

	cBase := types.ContainerJSONBase{
		ID:    "deadbeef",
		Image: "test",
	}

	ports := make(nat.PortMap)
	networkSettings := types.NetworkSettings{
		NetworkSettingsBase: types.NetworkSettingsBase{Ports: ports},
		Networks:            make(map[string]*network.EndpointSettings),
	}

	co := types.ContainerJSON{
		ContainerJSONBase: &cBase,
		Mounts:            make([]types.MountPoint, 0),
		Config:            &container.Config{},
		NetworkSettings:   &networkSettings,
	}
	assert.Empty(t, dl.getPortsFromInspect(co))

	ports = make(nat.PortMap)
	p, _ := nat.NewPort("tcp", "1234")
	ports[p] = make([]nat.PortBinding, 0)
	p, _ = nat.NewPort("tcp", "4321")
	ports[p] = make([]nat.PortBinding, 0)

	co.NetworkSettings.Ports = ports
	pts := dl.getPortsFromInspect(co)
	assert.Equal(t, 2, len(pts))
	assert.Contains(t, pts, 1234)
	assert.Contains(t, pts, 4321)
}
