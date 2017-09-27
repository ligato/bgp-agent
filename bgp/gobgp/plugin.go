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
	log "github.com/ligato/cn-infra/logging/logrus"
	goBgpConfig "github.com/osrg/gobgp/config"
	goBgpServer "github.com/osrg/gobgp/server"
	"strconv"
)

// Plugin is GoBGP Ligato BGP Plugin implementation
type Plugin struct {
	SessionConfig    *goBgpConfig.Bgp
	server           *goBgpServer.BgpServer
	watcherCallbacks []func(*bgp.ReachableRoute)
	stopWatch        chan bool
	PluginName       string
	Log              *log.Logger
}

//New creates a GoBGP Ligato BGP Plugin implementation
func New() *Plugin {
	return &Plugin{watcherCallbacks: []func(*bgp.ReachableRoute){}}
}

//Init uses plugin config and starts gobgp with dedicated goroutine for watching gobgp.
func (plugin *Plugin) Init() error {
	plugin.Log.Debug("Init goBgp plugin")
	if plugin.SessionConfig == nil {
		//TODO add config load in case of missing config injection
		return fmt.Errorf("Can't init GoBGP plugin without configuration.")
	}
	plugin.PluginName = plugin.SessionConfig.Global.Config.RouterId
	plugin.server = goBgpServer.NewBgpServer()
	go plugin.server.Serve()

	if err := plugin.startSession(); err != nil {
		return err
	}
	if err := plugin.addKnownNeighbors(); err != nil {
		return err
	}
	plugin.stopWatch = make(chan bool, 1)
	w := plugin.server.Watch(goBgpServer.WatchBestPath(true))
	go plugin.watchChanges(w)

	return nil
}

// watchChanges watches for events from goBGP server, translates them to bgp.ReachableRoute and sends them to registered watchers.
func (plugin *Plugin) watchChanges(watcher *goBgpServer.Watcher) {
	for {
		select {
		case <-plugin.stopWatch:
			plugin.Log.Debug("Stop Watching ", plugin.PluginName)
			watcher.Stop()
			return
		case ev := <-watcher.Event():
			switch msg := ev.(type) {
			case *goBgpServer.WatchEventBestPath:
				for _, path := range msg.PathList {
					as, err := strconv.ParseUint(path.GetAsPath().String(), 10, 32)
					if err != nil {
						plugin.Log.Fatal(err)
						return
					}
					pathInfo := bgp.ReachableRoute{
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
func (plugin *Plugin) WatchIPRoutes(subscriber string, callback func(*bgp.ReachableRoute)) error {
	plugin.Log.Infof("Subscriber %s registering for watching of IPRoutes in %s.", subscriber, plugin.PluginName)
	plugin.watcherCallbacks = append(plugin.watcherCallbacks, callback)
	return nil
}

//startSession starts session on already running goBGP server
func (plugin *Plugin) startSession() error {
	if err := plugin.server.Start(&plugin.SessionConfig.Global); err != nil {
		plugin.Log.Fatal("Failed to initialize go server", plugin.PluginName, err)
		return err
	}
	return nil
}

// addKnownNeighbors configures goBGP server for known neighbors from config
func (plugin *Plugin) addKnownNeighbors() error {
	for _, neighbor := range plugin.SessionConfig.Neighbors {
		if err := plugin.server.AddNeighbor(&neighbor); err != nil {
			plugin.Log.Fatal("Failed to add go neighbour", plugin.PluginName, err)
			return err
		}
	}

	return nil
}