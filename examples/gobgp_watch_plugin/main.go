// Copyright (c) 2017 Pantheon technologies s.r.o.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//Package gobgp_watch_plugin contains executable main for agent workflow example
package main

import (
	"github.com/ligato/bgp-agent/bgp"
	"github.com/ligato/bgp-agent/bgp/gobgp"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	"github.com/osrg/gobgp/config"
	"os"
	"time"
)

const (
	goBgpPluginName   = "goBgpPlugin"
	examplePluginName = "pluginExample"
)

var (
	goBgpConfig = &config.Bgp{
		Global: config.Global{
			Config: config.GlobalConfig{
				As:       65000,
				RouterId: "172.18.0.254",
				Port:     -1,
			},
		},
		Neighbors: []config.Neighbor{
			{
				Config: config.NeighborConfig{
					PeerAs:          65001,
					NeighborAddress: "172.18.0.2",
				},
			},
		},
	}
	flavor = &local.FlavorLocal{}
)

func main() {
	goBgpPlugin := gobgp.New(gobgp.Deps{
		PluginInfraDeps: *flavor.InfraDeps(goBgpPluginName, local.WithConf()),
		SessionConfig:   goBgpConfig})

	exPlugin := newExamplePlugin(goBgpPlugin)

	namedPlugins := []*core.NamedPlugin{
		{goBgpPlugin.PluginName, goBgpPlugin},
		{exPlugin.PluginName, exPlugin},
	}

	agent := core.NewAgent(logroot.StandardLogger(), 1*time.Minute, namedPlugins...)

	err := core.EventLoopWithInterrupt(agent, exPlugin.closeCh)
	if err != nil {
		os.Exit(1)
	}
}

type pluginExample struct {
	local.PluginInfraDeps
	reg         bgp.WatchRegistration
	goBgpPlugin *gobgp.Plugin
	closeCh     chan struct{}
}

func newExamplePlugin(plugin *gobgp.Plugin) *pluginExample {
	closeCh := make(chan struct{})
	return &pluginExample{
		PluginInfraDeps: *flavor.InfraDeps(examplePluginName),
		goBgpPlugin:     plugin,
		closeCh:         closeCh}
}

func (plugin *pluginExample) Init() error {
	reg, err := plugin.goBgpPlugin.WatchIPRoutes("watcher", func(information *bgp.ReachableIPRoute) {
		plugin.Log.Infof("Agent received path %v", information)
		close(plugin.closeCh)
	})
	plugin.reg = reg
	return err
}

func (plugin *pluginExample) Close() error {
	return plugin.reg.Close()
}
