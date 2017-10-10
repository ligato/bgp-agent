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
	"testing"
)

// TestGoBGPPluginInfoPassing tests gobgp plugin for the ability of retrieving of ReachableIPRoutes using BGP protocol and passing this information to its registered watchers.
// Test is also testing the ability of watch unregistering.
// Test creates beside goBGP plugin also route reflector so that goBGP plugin can receive data from its BGP-protocol communication partner. Each invocation of test will create also new route reflector.
func TestGoBGPPluginInfoPassing(x *testing.T) {
	t := TestHelper{golangTesting: x}
	t.DefaultSetup()
	defer t.Teardown()

	t.Given.RouteReflector()
	t.Given.GoBGPPluginWithWatcher() //connected to route reflector
	t.When.AddNewRoute()
	t.Then.WatcherReceivesAddedRoute()

	t.When.StopWatchingAndAddNewRoute()
	t.Then.WatcherReceivesNothing()
}
