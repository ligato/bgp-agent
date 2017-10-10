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

//Package bgp_test contains tests for API helper functions
package bgp_test

import (
	"github.com/ligato/bgp-agent/bgp"
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
	"net"
	"testing"
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
	golangT      *testing.T
	channel      chan bgp.ReachableIPRoute
	wrappingFunc func(info *bgp.ReachableIPRoute)
	sentRoute    bgp.ReachableIPRoute
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

	// registering gomega
	RegisterTestingT(t.vars.golangT)

	// initialize used channel
	t.vars.channel = make(chan bgp.ReachableIPRoute, 1)
}

// WrappingFuncAsToChanResult adds Wrapping function (that is produced by bgp.ToChan) together with corresponding channel
// (input for bgp.ToChan) to Given things that will be used in test later (see BDD Given).
func (g *Given) WrappingFuncAsToChanResult() {
	g.vars.wrappingFunc = bgp.ToChan(g.vars.channel, logrus.DefaultLogger())
}

// SentRouteToWrappingFunc creates route and passes it to previously Given wrapping function.
func (w *When) SentRouteToWrappingFunc() {
	w.vars.sentRoute = bgp.ReachableIPRoute{As: 1, Prefix: "1.2.3.4/32", Nexthop: net.IPv4(192, 168, 1, 1)}
	w.vars.wrappingFunc(&w.vars.sentRoute)
}

// ChannelReceivesIt assert that previously sent route to wrapping function is forwarded (or its copy is forwarded) to
// channel that was input for wrapping function creation.
func (t *Then) ChannelReceivesIt() {
	Expect(t.vars.channel).ToNot(BeEmpty())
	received := <-t.vars.channel
	Expect(received).To(Equal(t.vars.sentRoute))
	Expect(t.vars.channel).To(BeEmpty())
}
