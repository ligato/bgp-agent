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

package bgp

import (
	"github.com/ligato/cn-infra/logging/logrus"
	. "github.com/onsi/gomega"
	"net"
	"reflect"
	"testing"
)

// TestToChan tests ability of ToChan(...) function to create wrapping function that wraps given channel and forwards
// all ReachableIPRoute data (given to the wrapping function by parameter) to the channel inside.
// Test doesn't check logging capabilities of wrapping function.
func TestToChan(t *testing.T) {
	// prepare of tested instances/helper instances
	RegisterTestingT(t)
	channel := make(chan ReachableIPRoute, 1)
	wrappingFunc := ToChan(channel, logrus.DefaultLogger())

	// testing ToChan's returned function
	Expect(channel).To(BeEmpty())
	sent := ReachableIPRoute{As: 1, Prefix: "1.2.3.4/32", Nexthop: net.IPv4(192, 168, 1, 1)}
	wrappingFunc(&sent)
	Expect(channel).ToNot(BeEmpty())
	received := <-channel
	reflect.DeepEqual(sent, received)
	Expect(channel).To(BeEmpty())
}
