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
	"github.com/stretchr/testify/assert"
	"net"
	"reflect"
	"testing"
)

func TestToChan(t *testing.T) {
	// prepare of tested instances/helper instances
	assertThat := assert.New(t)
	channel := make(chan ReachableIPRoute, 1)
	wrappingFunc := ToChan(channel, logrus.DefaultLogger())

	// testing ToChan's returned function
	assertThat.Empty(channel)
	sent := ReachableIPRoute{As: 1, Prefix: "1.2.3.4/32", Nexthop: net.IPv4(192, 168, 1, 1)}
	wrappingFunc(&sent)
	assertThat.NotEmpty(channel)
	received := <-channel
	reflect.DeepEqual(sent, received)
	assertThat.Empty(channel)
}
