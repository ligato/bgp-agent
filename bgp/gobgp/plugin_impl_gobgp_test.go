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
	"errors"
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
	length                  uint8  = 24
	expectedReceivedAs             = uint32(65000)
	maxSessionEstablishment        = 2 * time.Minute
	timeoutForNotReceiving         = 5 * time.Second
)

var flavor = &local.FlavorLocal{}

// TestGoBGPPluginInfoPassing tests gobgp plugin for the ability of retrieving of ReachableIPRoutes using BGP protocol and passing this information to its registered watchers.
// Test is also testing the ability of watch unregistering.
func TestGoBGPPluginInfoPassing(t *testing.T) {
	// creating help variables/initialization
	RegisterTestingT(t)
	channel := make(chan bgp.ReachableIPRoute, 10)
	var lifecycleWG sync.WaitGroup

	// create components
	routeReflector, err := createAndStartRouteReflector(routeReflectorConf)
	Expect(err).To(BeNil(), "Can't create or start route reflector")
	goBGPPlugin := createGoBGPPlugin(serverConf)

	// watch -> start lifecycle of gobgp plugin -> send path to Route reflector -> check receiving on other end
	watchRegistration, registrationErr := goBGPPlugin.WatchIPRoutes("TestWatcher", bgp.ToChan(channel, logroot.StandardLogger()))
	Expect(registrationErr).To(BeNil(), "Can't properly register to watch IP routes")
	Expect(watchRegistration).NotTo(BeNil(), "WatchRegistration must be non-nil to be able to close registration later")
	lifecycleCloseChannel := startPluginLifecycle(goBGPPlugin, assertCorrectLifecycleEnd(&lifecycleWG))
	Expect(waitForSessionEstablishment(routeReflector)).To(BeNil(), "Session not established within timeout")
	Expect(addNewRoute(routeReflector, prefix1, nextHop1, length)).To(BeNil(), "Can't add new route")
	assertThatChannelReceivesCorrectRoute(channel)

	//unregister watching -> send another path to Route reflector -> check that nothing came to watcher
	Expect(watchRegistration.Close()).To(BeNil(), "Closing resitration failed")
	Expect(addNewRoute(routeReflector, prefix2, nextHop2, length)).To(BeNil(), "Can't add new route")
	assertThatChannelDidntReceiveAnything(channel, t)

	// stop all
	close(lifecycleCloseChannel) //gives command to stop gobgp lifecycle
	lifecycleWG.Wait()           //waiting for real gobgp lifecycle stop (that means also for assertion of correct closing (assertCorrectLifecycleEnd(...))
	Expect(routeReflector.Stop()).To(BeNil())
}

// assertCorrectLifecycleEnd asserts that lifecycle controlled by agent finished without error.
// This method is also used for synchronizing test run with lifecycle run, that means that above mentioned assertion happens within duration of test.
func assertCorrectLifecycleEnd(lifecycleWG *sync.WaitGroup) func(err error) {
	lifecycleWG.Add(1) // adding 1 to waitgroup before lifecycle loop starts in another go routine
	return func(err error) { // called insided lifecycle go routine just after ending the lifecycle
		Expect(err).To(BeNil())
		lifecycleWG.Done()
	}
}

// startPluginLifecycle creates cn-infra agent and use it to start lifecycle for passed gobgp plugin.
// Lambda parameter assertCorrectLifecycleEnd is used to verify correct lifecycle loop end and for synchronization
// purposes so that the verification happens in the duration of tests (another go routine can be kill without performing
// needed assertions). Returned channel can be use to stop agent's lifecycle loop.
func startPluginLifecycle(plugin *gobgp.Plugin, assertCorrectLifecycleEnd func(err error)) chan struct{} {
	agent := core.NewAgent(logroot.StandardLogger(), 1*time.Minute, []*core.NamedPlugin{{plugin.PluginName, plugin}}...)

	closeChannel := make(chan struct{}, 1)
	go func() {
		assertCorrectLifecycleEnd(core.EventLoopWithInterrupt(agent, closeChannel))
	}()
	return closeChannel
}

// createGoBGPPlugin creates basic gobgp plugin
func createGoBGPPlugin(bgpConfig *config.Bgp) *gobgp.Plugin {
	return gobgp.New(gobgp.Deps{
		PluginInfraDeps: *flavor.InfraDeps("TestGoBGP", local.WithConf()),
		SessionConfig:   bgpConfig})
}

// assertThatChannelReceivesCorrectRoute waits for received route and then checks it for correctness
func assertThatChannelReceivesCorrectRoute(channel chan bgp.ReachableIPRoute) {
	receivedRoute := <-channel
	logroot.StandardLogger().Println(receivedRoute)
	logroot.StandardLogger().Debug("Agent received new route ", receivedRoute)

	Expect(receivedRoute.As).To(Equal(expectedReceivedAs))
	Expect(receivedRoute.Nexthop.String()).To(Equal(nextHop1))
	Expect(receivedRoute.Prefix).To(Equal(prefix1 + "/24"))
}

// assertThatChannelDidntReceiveAnything waits for given time (timeoutForNotReceiving constant) and if in that time period data flows in data channel, then it fails the test.
func assertThatChannelDidntReceiveAnything(dataChan chan bgp.ReachableIPRoute, t *testing.T) {
	timeChan := time.NewTimer(timeoutForNotReceiving).C
	for {
		select {
		case <-timeChan:
			return
		case route := <-dataChan:
			t.Fatal("Channel did receive route even if it should not. Route received: ", route)
		}
	}
}

// createAndStartRouteReflector creates and starts Route Reflector
func createAndStartRouteReflector(bgpConfig *config.Bgp) (*server.BgpServer, error) {
	bgpServer := server.NewBgpServer()
	go bgpServer.Serve()

	if err := bgpServer.Start(&bgpConfig.Global); err != nil {
		stopFaultyBgpServer(bgpServer)
		return nil, err
	}
	if err := bgpServer.AddNeighbor(&bgpConfig.Neighbors[0]); err != nil {
		stopFaultyBgpServer(bgpServer)
		return nil, err
	}

	return bgpServer, nil
}

// stopFaultyBgpServer stops BgpServer that already does not correctly work. Possible error from stopping of server is
// not returned because root of the problem lies in previous detection of incorrect behaviour.
func stopFaultyBgpServer(bgpServer *server.BgpServer) {
	if err := bgpServer.Stop(); err != nil {
		logroot.StandardLogger().Error("error stoping server", err)
	}

}

// addNewRoute adds route to BgpServer
func addNewRoute(server *server.BgpServer, prefix string, nextHop string, length uint8) error {
	attrs := []bgpPacket.PathAttributeInterface{
		bgpPacket.NewPathAttributeOrigin(0),
		bgpPacket.NewPathAttributeNextHop(nextHop),
	}

	_, err := server.AddPath("",
		[]*table.Path{table.NewPath(
			nil,
			bgpPacket.NewIPAddrPrefix(length, prefix),
			false,
			attrs,
			time.Now(),
			false),
		},
	)
	return err
}

// waitForSessionEstablishment waits until it is possible to work with server correctly after start. Many commands depends on session being correctly established.
func waitForSessionEstablishment(server *server.BgpServer) error {
	timeChan := time.NewTimer(maxSessionEstablishment).C
	for {
		select {
		case <-timeChan:
			return errors.New("timer expired")
		default:
			time.Sleep(time.Second)
			if server.GetNeighbor("", false)[0].State.SessionState == config.SESSION_STATE_ESTABLISHED {
				return nil
			}
			logroot.StandardLogger().Debug("Waiting for session establishment")
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