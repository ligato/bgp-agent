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

//Package gobgp contains Ligato GoBGP Plugin implementation
package gobgp

import (
	"fmt"
	"github.com/ligato/bgp-agent/bgp"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/osrg/gobgp/config"
	"github.com/osrg/gobgp/server"
	"strconv"
	"sync"
)

// Plugin is GoBGP Ligato BGP Plugin implementation
type Plugin struct {
	Deps
	server                *server.BgpServer
	serverWatcher         *server.Watcher
	watchersWithCallbacks map[watcherName]func(*bgp.ReachableIPRoute)
	stopWatch             chan bool
	watchWG               sync.WaitGroup // wait group that allows to wait until Watch loop is ended
}

// Deps combines all needed dependencies for Plugin struct. These dependencies should be injected into Plugin by using constructor's Deps parameter.
type Deps struct {
	local.PluginInfraDeps             // inject
	SessionConfig         *config.Bgp // optional inject (if not injected, it must be set using external config file)
}

// watcherName is by-name identification of registered watcher
type watcherName string

//New creates a GoBGP Ligato BGP Plugin implementation. Needed dependencies are injected into plugin implementation.
func New(dependencies Deps) *Plugin {
	return &Plugin{Deps: dependencies, watchersWithCallbacks: map[watcherName]func(*bgp.ReachableIPRoute){}}
}

//Init creates the gobgp server and checks if needed SessionConfig was injected and fails if it is not.
func (plugin *Plugin) Init() error {
	plugin.Log.Debug("Init goBgp plugin")
	plugin.applyExternalConfig()
	if plugin.SessionConfig == nil {
		return fmt.Errorf("Can't init GoBGP plugin without configuration")
	}
	plugin.server = server.NewBgpServer()

	return nil
}

// applyExternalConfig tries to find and load configuration from external filesystem and change it for injected configuration, because external configuration has higher priority.
// If external configuration is not found or can't be loaded or other problem occur, plugin.SessionConfig is not changed. This means that previous injection of plugin.SessionConfig
// variable can be still used.
func (plugin *Plugin) applyExternalConfig() {
	var externalCfg *config.Bgp
	found, err := plugin.PluginConfig.GetValue(externalCfg)	// It tries to lookup `PluginName + "-config"` in go run command flags.
	if err != nil {
		plugin.Log.Debug("External GoBGP plugin configuration could not load or other problem happened", err)
		return
	}
	if !found {
		plugin.Log.Debug("External GoBGP plugin configuration was not found")
		return
	}
	plugin.SessionConfig = externalCfg
}

// AfterInit starts gobgp with dedicated goroutine for watching gobgp and forwarding best path reachable ip routes to registered watchers.
// After start of gobgp session, known neighbors from configuration are added to gobgp server.
// Due to fact that AfterInit is called once Init() of all plugins have returned without error, other plugins can be registered watchers
// from the start of gobgp server if they call this plugin's WatchIPRoutes() in their Init(). In this way they won't miss any information
// forwarded to registered watchers just because they registered too late.
func (plugin *Plugin) AfterInit() error {
	go plugin.server.Serve()
	if err := plugin.startSession(); err != nil {
		return err
	}
	if err := plugin.addKnownNeighbors(); err != nil {
		return err
	}
	plugin.stopWatch = make(chan bool, 1)
	plugin.serverWatcher = plugin.server.Watch(server.WatchBestPath(true))
	plugin.watchWG.Add(1)
	go plugin.watchChanges(plugin.serverWatcher)

	return nil
}

// watchChanges watches for events from goBGP server, translates them to bgp.ReachableIPRoute and sends them to registered watchers.
func (plugin *Plugin) watchChanges(watcher *server.Watcher) {
	defer plugin.watchWG.Done()

	for {
		select {
		case <-plugin.stopWatch:
			plugin.Log.Debug("Stop Watching ", plugin.PluginName)
			return
		case ev := <-watcher.Event():
			switch msg := ev.(type) {
			case *server.WatchEventBestPath:
				for _, path := range msg.PathList {
					asPath := path.GetAsPath().String()
					as, err := strconv.ParseUint(asPath, 10, 32)
					if err != nil {
						plugin.Log.Warnf("Ignoring Path '%s' due to parse error: %v", asPath, err)
						continue
					}
					pathInfo := bgp.ReachableIPRoute{
						As:      uint32(as),
						Prefix:  path.GetNlri().String(),
						Nexthop: path.GetNexthop(),
					}
					plugin.Log.Debug("Fill channel with new path", pathInfo)
					for _, callback := range plugin.watchersWithCallbacks {
						callback(&pathInfo)
					}
				}
			}
		}
	}
}

//Close stops dedicated goroutine for watching gobgp. Then stops watcher provider by gobgp server and finally stops that gobgp server itself.
func (plugin *Plugin) Close() error {
	plugin.Log.Info("Closing goBgp plugin ", plugin.PluginName)
	close(plugin.stopWatch) //command to stop watching
	plugin.watchWG.Wait()   //wait for actual stop of watching
	plugin.serverWatcher.Stop()
	return plugin.server.Stop()
}

//WatchIPRoutes register watcher to notifications for any new learned IP-based routes.
//WatchRegistration is not retroactive, that means that any IP-based routes learned in the past are not send to new watchers.
//This also means that if you want be notified of all learned IP-based routes, you must register before calling of
//AfterInit(). In case of external(=not other plugin started with this plugin) watchers this means before plugin start.
//However, late-registered watchers are permitted (no error will be returned), but they can miss some learned IP-based routes.
func (plugin *Plugin) WatchIPRoutes(watcher string, callback func(*bgp.ReachableIPRoute)) (bgp.WatchRegistration, error) {
	plugin.Log.Infof("Watcher %s registering for watching of IPRoutes in %s.", watcher, plugin.PluginName)
	plugin.watchersWithCallbacks[watcherName(watcher)] = callback
	return &watchRegistration{watcher: watcherName(watcher), plugin: plugin}, nil
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

// watchRegistration is Plugin's simple WatchRegistration implementation that is sent to watchers.
// This implementation is not thread-safe.
type watchRegistration struct {
	watcher watcherName
	plugin  *Plugin
}

//Close ends the agreement between Plugin and watcher. Plugin stops sending watcher any further notifications.
func (wr *watchRegistration) Close() error {
	delete(wr.plugin.watchersWithCallbacks, wr.watcher)
	return nil
}
