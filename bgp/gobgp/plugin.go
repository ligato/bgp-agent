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

//Package gobgp contains Ligato GoBGP BGP Plugin implementation
package gobgp

import (
	"fmt"
	"github.com/ligato/bgp-agent/bgp"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/server"
	"strconv"
)

// Plugin is GoBGP Ligato BGP Plugin implementation
type Plugin struct {
	server           *server.BgpServer
	watcherCallbacks []func(*bgp.ReachableIPRoute)
	stopWatch        chan bool
	PluginName       string
	Deps
}

// Deps combines all needed dependencies for Plugin struct. These dependencies should be injected into Plugin by using constructor's Deps parameter.
type Deps struct {
	local.PluginInfraDeps
	SessionConfig    *config.Bgp
}


func New(dependencies Deps) *Plugin {
	return &Plugin{Deps: dependencies, watcherCallbacks: []func(*bgp.ReachableIPRoute){}}
}

//Init uses plugin config and starts gobgp with dedicated goroutine for watching gobgp and forwarding best path reachable ip routes to registered watchers.
//After start of gobgp session, known neighbors from configuration are added to gobgp server.
func (plugin *Plugin) Init() error {
	plugin.Log.Debug("Init goBgp plugin")
	if plugin.SessionConfig == nil {
		//TODO add config load in case of missing config injection
		return fmt.Errorf("Can't init GoBGP plugin without configuration")
	}
	plugin.PluginName = plugin.SessionConfig.Global.Config.RouterId
	plugin.server = server.NewBgpServer()
	go plugin.server.Serve()

	if err := plugin.startSession(); err != nil {
		return err
	}
	if err := plugin.addKnownNeighbors(); err != nil {
		return err
	}
	plugin.stopWatch = make(chan bool, 1)
	w := plugin.server.Watch(server.WatchBestPath(true))
	go plugin.watchChanges(w)

	return nil
}

// watchChanges watches for events from goBGP server, translates them to bgp.ReachableIPRoute and sends them to registered watchers.
func (plugin *Plugin) watchChanges(watcher *server.Watcher) {
	for {
		select {
		case <-plugin.stopWatch:
			plugin.Log.Debug("Stop Watching ", plugin.PluginName)
			watcher.Stop()
			return
		case ev := <-watcher.Event():
			switch msg := ev.(type) {
			case *server.WatchEventBestPath:
				for _, path := range msg.PathList {
					as, err := strconv.ParseUint(path.GetAsPath().String(), 10, 32)
					if err != nil {
						plugin.Log.Fatal(err)
						return
					}
					pathInfo := bgp.ReachableIPRoute{
						As:      uint32(as),
						Prefix:  path.GetNlri().String(),
						Nexthop: path.GetNexthop(),
					}
					plugin.Log.Debug("Fill channel with new path", pathInfo)
					for _, callback := range plugin.watcherCallbacks {
						callback(&pathInfo)
					}
				}
			}
		}
	}
}

//Close performs clean up and final close
func (plugin *Plugin) Close() error {
	plugin.Log.Info("Closing goBgp plugin ", plugin.PluginName)
	close(plugin.stopWatch)
	return plugin.server.Stop()
}

//WatchIPRoutes subscribes consumer to notifications for any new learned IP-based routes
func (plugin *Plugin) WatchIPRoutes(subscriber string, callback func(*bgp.ReachableIPRoute)) error {
	plugin.Log.Infof("Subscriber %s registering for watching of IPRoutes in %s.", subscriber, plugin.PluginName)
	plugin.watcherCallbacks = append(plugin.watcherCallbacks, callback)
	return nil
}

//startSession starts session on already running goBGP server
func (plugin *Plugin) startSession() error {
	if err := plugin.server.Start(&plugin.SessionConfig.Global); err != nil {
		plugin.Log.Error("Failed to initialize go server", plugin.PluginName, err)
		return err
	}
	return nil
}

// addKnownNeighbors configures goBGP server for known neighbors from config
func (plugin *Plugin) addKnownNeighbors() error {
	for _, neighbor := range plugin.SessionConfig.Neighbors {
		if err := plugin.server.AddNeighbor(&neighbor); err != nil {
			plugin.Log.Error("Failed to add go neighbour", plugin.PluginName, err)
			return err
		}
	}

	return nil
}