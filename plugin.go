package qcache_inventory

import (
	"fmt"
	"time"
	"github.com/zpatrick/go-config"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"github.com/deckarep/golang-set"

	"github.com/qframe/types/constants"
	"github.com/qframe/types/docker-events"
	"github.com/qframe/types/qchannel"
	"github.com/qframe/types/plugin"
	"reflect"
	"strings"
	"github.com/docker/docker/api/types"
)

const (
	version = "0.3.1"
	pluginTyp = qtypes_constants.CACHE
	pluginPkg = "inventory"
	dockerAPI = "v1.29"
)

var (
	ctx = context.Background()
)

type Plugin struct {
	*qtypes_plugin.Plugin
	Inventory Inventory
	engCli *client.Client

}

func New(qChan qtypes_qchannel.QChan, cfg *config.Config, name string) (Plugin, error) {
	return Plugin{
		Plugin: qtypes_plugin.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
		Inventory: NewInventory(),
	}, nil
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start inventory v%s", p.Version))
	dc := p.QChan.Data.Join()
	tickerTime := p.CfgIntOr("ticker-ms", 500)
	ticker := time.NewTicker(time.Millisecond * time.Duration(tickerTime)).C
	dockerHost := p.CfgStringOr("docker-host", "unix:///var/run/docker.sock")
	var err error
	p.engCli, err = client.NewClient(dockerHost, dockerAPI, nil, nil)
	if err != nil {
		p.Log("error", fmt.Sprintf("Could not connect docker/docker/client to '%s': %v", dockerHost, err))
		return
	}
	info, err := p.engCli.Info(ctx)
	if err != nil {
		p.Log("error", fmt.Sprintf("Error during Info(): %v >err> %s", info, err.Error()))
		return
	} else {
		p.Log("info", fmt.Sprintf("Connected to '%s' / v'%s'", info.Name, info.ServerVersion))
	}
	for {
		select {
		case val := <-dc.Read:
			switch val.(type) {
			case qtypes_docker_events.ContainerEvent:
				ce := val.(qtypes_docker_events.ContainerEvent)
				if ce.Event.Type == "container" && ce.Event.Action == "start" {
					ipSet := mapset.NewSet()
					for _,v := range ce.Container.NetworkSettings.Networks {
						ipSet.Add(v.IPAddress)
					}
					ips, err := p.AddNetworkIPs(ipSet, ce)
					if err != nil {
						p.Log("error", fmt.Sprintf("Error during AddNetworkIPs(): %s", err.Error()))
					}
					p.Log("debug", fmt.Sprintf("Add CntID:%s into Inventory (name:%s, IPs:%s)",ce.Container.ID[:13], ce.Container.Name, strings.Join(ips,",")))
					p.Inventory.SetItem(ce.Container.ID, &ce.Container, ips)
				}
			case ContainerRequest:
				req := val.(ContainerRequest)
				p.Log("debug", fmt.Sprintf("Received InventoryRequest for %v", req))
				p.Inventory.ServeRequest(req)
			default:
				p.Log("trace", fmt.Sprintf("Dunno type '%s': %v", reflect.TypeOf(val), val))

			}
		case <- ticker:
			p.Log("trace", "Ticker came along: p.Inventory.CheckRequests()")
			p.Inventory.CheckRequests()
			continue
		}
	}
}

func (p *Plugin) AddNetworkIPs(ips  mapset.Set, ce qtypes_docker_events.ContainerEvent) (res []string, err error) {
	nets, err := p.engCli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		p.Log("error", err.Error())
		return GetList(ips), err
	}
	for _, net := range nets {
		p.Log("trace", fmt.Sprintf(">> Network: %s", net.Name))
		netInspect, err := p.engCli.NetworkInspect(ctx, net.ID, types.NetworkInspectOptions{})
		if err != nil {
			p.Log("error", fmt.Sprintf("Error during NetworkList(): %s", err.Error()))
			continue
		}
		for cntID, cnt := range netInspect.Containers {
			if cntID == ce.Container.ID {
				ip4 := strings.Split(cnt.IPv4Address, "/")
				if len(ip4) != 2 {
					p.Log("error", fmt.Sprintf("Container '%s' has unexpected IP '%s'... skip", cntID, cnt.IPv4Address))
					continue
				}
				p.Log("trace", fmt.Sprintf("   Name:%-15s || %s==%s: %s", cnt.Name, cntID[:12], ce.Container.ID[:12], ip4[0]))
				ips.Add(ip4[0])
			}
		}
	}
	return GetList(ips), err
}

func GetList(s mapset.Set) (res []string) {
	it := s.Iterator()
	for elem := range it.C {
		res = append(res, elem.(string))
	}
	return
}