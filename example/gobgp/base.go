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

//Package gobgp contains executable main for agent workflow example
package main

import (
	"github.com/ligato/bgp-agent/bgp"
	"github.com/ligato/bgp-agent/bgp/gobgp"
	errorUtil "github.com/ligato/bgp-agent/utils/error"
	"github.com/ligato/bgp-agent/utils/sync"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	log "github.com/ligato/cn-infra/logging/logrus"
	"github.com/osrg/gobgp/config"
	"sync/atomic"
	"time"
)

var (
	counter     = uint32(0)
	goBgpConfig = &config.Bgp{
		Global: config.Global{
			Config: config.GlobalConfig{
				As:       65000,
				RouterId: "172.18.0.254",
				Port:     -1,
			},
		},
		Neighbors: []config.Neighbor{
			config.Neighbor{
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
		PluginInfraDeps: *flavor.InfraDeps("example"),
		SessionConfig:   goBgpConfig})

	reg, err := goBgpPlugin.WatchIPRoutes("watcher", func(information *bgp.ReachableIPRoute) {
		log.DefaultLogger().Infof("Agent received new path %v", information)
		atomic.AddUint32(&counter, 1)
	})
	errorUtil.PanicIfError(err)

	namedPlugins := []*core.NamedPlugin{{"ExampleNamedPlugin", goBgpPlugin}}
	agent := core.NewAgent(logroot.StandardLogger(), 1*time.Minute, namedPlugins...)

	errorUtil.PanicIfError(agent.Start())
	errorUtil.PanicIfError(sync.WaitCounterMatch(2*time.Minute, 1, &counter))
	errorUtil.PanicIfError(agent.Stop())
	reg.Close()
}
