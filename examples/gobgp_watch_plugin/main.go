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

// main is the main function for running the whole example
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

// pluginExample is cn-infra plugin that serves as example watcher registered in goBGPplugin.
// Purpose of this plugin is to show process of registration/unregistration of goBGPPlugin watchers and also how could be the data
// received by the registered watchers.
type pluginExample struct {
	local.PluginInfraDeps
	reg         bgp.WatchRegistration
	goBgpPlugin *gobgp.Plugin
	closeCh     chan struct{}
}

// newExamplePlugin creates new pluginExample with reference to goBGPPlugin(<plugin> param) so it can use it later to register itself to goBGPPlugin as watcher.
func newExamplePlugin(plugin *gobgp.Plugin) *pluginExample {
	closeCh := make(chan struct{})
	return &pluginExample{
		PluginInfraDeps: *flavor.InfraDeps(examplePluginName),
		goBgpPlugin:     plugin,
		closeCh:         closeCh}
}

// Init registers pluginExample as watcher in goBGPPlugin. Callback for registered pluginExample will receive first route
// from goBGPPlugin and then stops the whole example.
// In case of problems with watcher registration, error (forwarded from registration failure) is returned and initialization
// of pluginExample fails.
func (plugin *pluginExample) Init() error {
	reg, err := plugin.goBgpPlugin.WatchIPRoutes("watcher", func(information *bgp.ReachableIPRoute) {
		plugin.Log.Infof("Agent received path %v", information)
		close(plugin.closeCh)
	})
	plugin.reg = reg
	return err
}

// Close will unregister pluginExample as watcher in goBGPPlugin.
// If unregister fails, pluginExample will fail too (with the same error).
func (plugin *pluginExample) Close() error {
	return plugin.reg.Close()
}
