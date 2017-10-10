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

//Package gobgp_test contains Ligato GoBGP Plugin implementation tests
package gobgp_test

import (
	"github.com/ligato/bgp-agent/bgp"
	"github.com/ligato/bgp-agent/bgp/gobgp"
	"github.com/ligato/cn-infra/core"
	"github.com/ligato/cn-infra/flavors/local"
	"github.com/ligato/cn-infra/logging/logroot"
	. "github.com/onsi/gomega"
	"github.com/osrg/gobgp/config"
	bgpPacket "github.com/osrg/gobgp/packet/bgp"
	"github.com/osrg/gobgp/server"
	"github.com/osrg/gobgp/table"
	"sync"
	"testing"
	"time"
)

const (
	nextHop1                string = "10.0.0.1"
	nextHop2                string = "10.0.0.3"
	prefix1                 string = "10.0.0.0"
	prefix2                 string = "10.0.0.2"
	prefixMaskLength        uint8  = 24
	expectedReceivedAs             = uint32(65000)
	maxSessionEstablishment        = 2 * time.Minute
	timeoutForReceiving            = 30 * time.Second
	timeoutForNotReceiving         = 5 * time.Second
)

// TestHelper allows tests to be written in given/when/then idiom of BDD
type TestHelper struct {
	vars          *Variables
	Given         Given
	When          When
	Then          Then
	golangTesting *testing.T
}

// Variables is container of variables that should be accessible from every BDD component
type Variables struct {
	golangT               *testing.T
	routeReflector        *server.BgpServer
	goBGPPlugin           *gobgp.Plugin
	dataChannel           chan bgp.ReachableIPRoute
	lifecycleCloseChannel chan struct{}
	lifecycleWG           sync.WaitGroup
	watchRegistration     bgp.WatchRegistration
}

// Given is composition of multiple test step methods (see BDD Given keyword)
type Given struct {
	vars *Variables
}

// When is composition of multiple test step methods (see BDD When keyword)
type When struct {
	vars *Variables
}

// Then is composition of multiple test step methods (see BDD Then keyword)
type Then struct {
	vars *Variables
}

// DefaultSetup setups needed variables and ensures that these variables are accessible from all test BDD components (given, when, then)
func (t *TestHelper) DefaultSetup() {
	// creating and linking variables to test parts
	t.vars = &Variables{}
	t.Given.vars = t.vars
	t.When.vars = t.vars
	t.Then.vars = t.vars
	t.vars.golangT = t.golangTesting

	// registering gomega
	RegisterTestingT(t.vars.golangT)

	// initialize data channel
	t.vars.dataChannel = make(chan bgp.ReachableIPRoute, 10)
}

// Teardown handles properly releasing of resources or stopping of components (route reflector, agent with plugins)
func (t *TestHelper) Teardown() {
	//if lifecycle for plugin is used, then we need to properly wait for its end
	if t.vars.lifecycleCloseChannel != nil {
		close(t.vars.lifecycleCloseChannel) //gives command to stop agent lifecycle
		t.vars.lifecycleWG.Wait()           //waiting for real agent lifecycle stop (that means also for assertion of correct closing (assertCorrectLifecycleEnd(...))
	}

	//if route reflector is used, then we need to stop it
	if t.vars.routeReflector != nil {
		Expect(t.vars.routeReflector.Stop()).To(BeNil())
	}
}

// RouteReflector creates and starts Route Reflector. The route reflector functionality is simulated by GoBGP server.
func (g *Given) RouteReflector() {
	bgpServer := server.NewBgpServer()
	go bgpServer.Serve()

	if err := bgpServer.Start(&routeReflectorConf.Global); err != nil {
		g.stopFaultyRouteReflector()
		g.vars.golangT.Fatal(err, "Can't create or start route reflector")
	}
	if err := bgpServer.AddNeighbor(&routeReflectorConf.Neighbors[0]); err != nil {
		g.stopFaultyRouteReflector()
		g.vars.golangT.Fatal(err, "Can't create or start route reflector")
	}
	g.vars.routeReflector = bgpServer
}

// stopFaultyRouteReflector stops route reflector's BGP Server that already does not correctly work. Possible error from stopping of server is
// not returned because root of the problem lies in previous detection of incorrect behaviour.
func (g *Given) stopFaultyRouteReflector() {
	if err := g.vars.routeReflector.Stop(); err != nil {
		logroot.StandardLogger().Error("error stoping route reflector BGP server", err)
	}

}

// GoBGPPluginWithWatcher creates GoBGPPlugin (with Watcher registered in it) and prepares it for usage.
// As part of preparing, GoBGPplugin is started inside cn-infra agent. If route reflector was previously created, then
// we are also waiting for proper session establishment between route reflector and GoBGP plugin.
func (g *Given) GoBGPPluginWithWatcher() {
	//plugin creation
	g.createGoBGPPlugin()

	//adding watcher
	var registrationErr error
	g.vars.watchRegistration, registrationErr = g.vars.goBGPPlugin.WatchIPRoutes("TestWatcher", bgp.ToChan(g.vars.dataChannel, logroot.StandardLogger()))
	Expect(registrationErr).To(BeNil(), "Can't properly register to watch IP routes")
	Expect(g.vars.watchRegistration).NotTo(BeNil(), "WatchRegistration must be non-nil to be able to close registration later")

	//starting it using cn-infra agent
	g.startPluginLifecycle()

	//if route reflector used, then waiting for creating session between route reflector and GoBGP plugin
	if g.vars.routeReflector != nil {
		g.waitForSessionEstablishment()
	}
}

// waitForSessionEstablishment waits until it is possible to work with server correctly after start. Many commands depends on session being correctly established.
func (g *Given) waitForSessionEstablishment() {
	timeChan := time.NewTimer(maxSessionEstablishment).C
	for {
		select {
		case <-timeChan:
			g.vars.golangT.Fatal("Session not established within timeout")
		default:
			time.Sleep(time.Second)
			if g.vars.routeReflector.GetNeighbor("", false)[0].State.SessionState == config.SESSION_STATE_ESTABLISHED {
				return
			}
			logroot.StandardLogger().Debug("Waiting for session establishment")
		}
	}
}

// createGoBGPPlugin creates properly filled structure of gobgp plugin
func (g *Given) createGoBGPPlugin() {
	flavor := &local.FlavorLocal{}
	g.vars.goBGPPlugin = gobgp.New(gobgp.Deps{
		PluginInfraDeps: *flavor.InfraDeps("TestGoBGP", local.WithConf()),
		SessionConfig:   serverConf})
}

// startPluginLifecycle creates cn-infra agent and uses it to start lifecycle for gobgp plugin.
// For future agent lifecycle closing purpose, lifecycleCloseChannel in Given.Vars is used.
// For proper wait for lifecycle end, lifecycleWG in Given.Vars is used.
func (g *Given) startPluginLifecycle() {
	agent := core.NewAgent(logroot.StandardLogger(), 1*time.Minute, []*core.NamedPlugin{{g.vars.goBGPPlugin.PluginName, g.vars.goBGPPlugin}}...)
	g.vars.lifecycleCloseChannel = make(chan struct{}, 1)
	g.vars.lifecycleWG.Add(1) // adding 1 to waitgroup before lifecycle loop starts in another go routine
	go func() {
		Expect(core.EventLoopWithInterrupt(agent, g.vars.lifecycleCloseChannel)).To(BeNil(), "Agent's lifecycle didn't ended properly")
		g.vars.lifecycleWG.Done() // ending waiting for lifecycle loop end
	}()
}

// AddNewRoute adds first constant-based route to route reflector and asserts success.
func (w *When) AddNewRoute() {
	Expect(w.addNewRoute(prefix1, nextHop1, prefixMaskLength)).To(BeNil(), "Can't add new route")
}

// StopWatchingAndAddNewRoute closes watch registration and adds second constant-based route to route reflector. For both actions success is asserted.
func (w *When) StopWatchingAndAddNewRoute() {
	Expect(w.vars.watchRegistration.Close()).To(BeNil(), "Closing registration failed")
	Expect(w.addNewRoute(prefix2, nextHop2, prefixMaskLength)).To(BeNil(), "Can't add new route")
}

// addNewRoute adds route to Route reflector.
// New route is defined by <prefix>(full IPv4 address as string) with
// <prefixMaskLength>(count of bit of network mask for prefix) and <nextHop>(full IPv4 address as string).
// It returns nonnill error if addition was not successful.
func (w *When) addNewRoute(prefix string, nextHop string, prefixMaskLength uint8) error {
	attrs := []bgpPacket.PathAttributeInterface{
		bgpPacket.NewPathAttributeOrigin(0),
		bgpPacket.NewPathAttributeNextHop(nextHop),
	}

	_, err := w.vars.routeReflector.AddPath("",
		[]*table.Path{table.NewPath(
			nil,
			bgpPacket.NewIPAddrPrefix(prefixMaskLength, prefix),
			false,
			attrs,
			time.Now(),
			false),
		},
	)
	return err
}

// WatcherReceivesNothing is timeout-based wait to assert that nothing comes to watcher by watcher data flow source. Timeout can be changed by changing the timeoutForNotReceiving constant.
func (t *Then) WatcherReceivesNothing() {
	timeChan := time.NewTimer(timeoutForNotReceiving).C
	for {
		select {
		case <-timeChan:
			return
		case route := <-t.vars.dataChannel:
			t.vars.golangT.Fatal("Channel did receive route even if it should not. Route received: ", route)
		}
	}
}

// WatcherReceivesAddedRoute waits for received route and then checks it for correctness. If nothing comes in timeout
// (timeoutForReceiving constant), test fails.
func (t *Then) WatcherReceivesAddedRoute() {
	timeChan := time.NewTimer(timeoutForReceiving).C
	for {
		select {
		case <-timeChan:
			t.vars.golangT.Fatal("Channel didn't received any root, but it should have.")
			return
		case receivedRoute := <-t.vars.dataChannel:
			logroot.StandardLogger().Println(receivedRoute)
			logroot.StandardLogger().Debug("Agent received new route ", receivedRoute)

			Expect(receivedRoute.As).To(Equal(expectedReceivedAs))
			Expect(receivedRoute.Nexthop.String()).To(Equal(nextHop1))
			Expect(receivedRoute.Prefix).To(Equal(prefix1 + "/24"))
			return
		}
	}
}

var (
	serverConf = &config.Bgp{
		Global: config.Global{
			Config: config.GlobalConfig{
				As:       65001,
				RouterId: "172.18.0.2",
				Port:     -1,
			},
		},
		Neighbors: []config.Neighbor{
			config.Neighbor{
				Config: config.NeighborConfig{
					PeerAs:          65000,
					NeighborAddress: "127.0.0.1",
				},
				Transport: config.Transport{
					Config: config.TransportConfig{
						RemotePort: 10179,
					},
				},
			},
		},
	}
	routeReflectorConf = &config.Bgp{
		Global: config.Global{
			Config: config.GlobalConfig{
				As:       65000,
				RouterId: "172.18.0.254",
				Port:     10179,
			},
		},
		Neighbors: []config.Neighbor{
			config.Neighbor{
				Config: config.NeighborConfig{
					PeerAs:          65001,
					NeighborAddress: "127.0.0.1",
				},
				Transport: config.Transport{
					Config: config.TransportConfig{
						PassiveMode: true,
					},
				},
			},
		},
	}
)
