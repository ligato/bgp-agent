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
	log "github.com/ligato/cn-infra/logging/logrus"
	goBgpConfig "github.com/osrg/gobgp/config"
	goBgpServer "github.com/osrg/gobgp/server"
	"github.com/fgschwan/bgp-agent/bgp"
	"github.com/fgschwan/bgp-agent/bgp/nlri"
	"github.com/fgschwan/bgp-agent/utils/gobgp"
	"github.com/fgschwan/bgp-agent/utils/parse"
	"fmt"
)

type pluginImpl struct {
	SessionConfig *goBgpConfig.Bgp
	server        *goBgpServer.BgpServer
	agentChannel  chan<- nlri.ReachableRoute
	bgp.PluginCore
	bgp.Session
	stopWatch  chan bool
	pluginID   string
}

//BGPPluginDebug is the public implementation of  GoBGP Ligato BGP Plugin and PluginDebug
type BGPPluginDebug struct {
	bgp.Plugin
	bgp.Session
}

//New creates a GoBGP Ligato BGP Plugin implementation
func New() BGPPluginDebug {
	plugin := pluginImpl{}
	return BGPPluginDebug{
		Plugin:  bgp.Plugin{PluginCore: &plugin},
		Session: &plugin,
	}
}

func (plugin *pluginImpl) WaitForSessionEstablishment() error {
	return gobgp.WaitForSessionEstablishment(plugin.server)
}

//Init defines initial status of the plugin using given config.
func (plugin *pluginImpl) Init() error {
	log.DefaultLogger().Debug("Init goBgp plugin")
	if plugin.SessionConfig == nil {
		//TODO add config load in case of missing config injection
		return fmt.Errorf("Can't init GoBGP plugin without configuration.")
	}
	plugin.pluginID = plugin.SessionConfig.Global.Config.RouterId

	server, err := plugin.startSession()
	if err != nil {
		return err
	}
	plugin.server = server
	plugin.stopWatch = make(chan bool, 1)
	w := server.Watch(goBgpServer.WatchBestPath(true))
	go plugin.watchChanges(w)

	return nil
}

func (plugin *pluginImpl) watchChanges(watcher *goBgpServer.Watcher) {
	for {
		select {
		case <-plugin.stopWatch:
			log.DefaultLogger().Debug("Stop Watching ", plugin.pluginID)
			watcher.Stop()
			return
		case ev := <-watcher.Event():
			switch msg := ev.(type) {
			case *goBgpServer.WatchEventBestPath:
				for _, path := range msg.PathList {
					pathInfo := nlri.ReachableRoute{}
					as, err := parse.ToUint32(path.GetAsPath().String())
					if err != nil {
						log.DefaultLogger().Fatal(err)
						if sessionErr := plugin.closeSession(); sessionErr != nil {
							log.DefaultLogger().Fatal("Failed to close session after AS path convertion failure", plugin.pluginID, sessionErr)
						}
						return
					}
					pathInfo.As = as
					pathInfo.Prefix = path.GetNlri().String()
					pathInfo.Nexthop = path.GetNexthop()
					log.DefaultLogger().Debug("Fill channel with new path", pathInfo)
					plugin.agentChannel <- pathInfo
				}
			}
		}
	}
}

func (plugin *pluginImpl) closeSession() error {
	log.DefaultLogger().Info("Close server ", plugin.pluginID)
	close(plugin.stopWatch)
	return plugin.server.Stop()
}

//Close performs clean up and final close
func (plugin *pluginImpl) Close() error {
	log.DefaultLogger().Info("Closing goBgp plugin ", plugin.pluginID)
	err := plugin.closeSession()
	if err != nil {
		return err
	}
	return nil
}

//Register registers plugin, linking pluginChannel to agentChannel
func (plugin *pluginImpl) Register(channel chan nlri.ReachableRoute) error {
	log.DefaultLogger().Info("Registering goBgp plugin channel ", plugin.pluginID)
	plugin.agentChannel = channel
	return nil
}

//startSession configures go BGP Server and starts the server
func (plugin *pluginImpl) startSession() (*goBgpServer.BgpServer, error) {
	server := goBgpServer.NewBgpServer()
	go server.Serve()

	if err := server.Start(&plugin.SessionConfig.Global); err != nil {
		log.DefaultLogger().Fatal("Failed to initialize go server", plugin.pluginID, err)
		return nil, err
	}

	for _, neighbor := range plugin.SessionConfig.Neighbors {
		if err := server.AddNeighbor(&neighbor); err != nil {
			log.DefaultLogger().Fatal("Failed to add go neighbour", plugin.pluginID, err)
			if sessionErr := plugin.closeSession(); sessionErr != nil {
				log.DefaultLogger().Fatal("Failed to close session after failure of add go neighbour", plugin.pluginID, sessionErr)
			}
			return nil, err
		}
	}

	return server, nil
}
