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

//Package bgp contains definitions for Ligato BGP-Agent Plugins
package bgp

import (
	"github.com/ligato/cn-infra/logging"
	"net"
)

// ReachableIPRoute represents new learned IP-based route that could be used for route-based decisions.
type ReachableIPRoute struct {
	As      uint32
	Prefix  string
	Nexthop net.IP
}

// WatchRegistration represents both-side-agreed agreement between Plugin and watchers that binds Plugin to notify watchers
// about new learned IP-based routes.
// WatchRegistration implementation is meant for watcher side as evidence about agreement and way how to access watcher side
// control upon agreement (i.e. to close it). Implementations don't have to be thread-safe.
type WatchRegistration interface {
	//Close ends the agreement between Plugin and watcher. Plugin stops sending watcher any further notifications.
	Close() error
}

// ToChan creates a callback that can be passed to the Watch function in order to receive
// notifications through a channel.
func ToChan(ch chan ReachableIPRoute, logger logging.Logger) func(info *ReachableIPRoute) {
	return func(info *ReachableIPRoute) {
		ch <- *info
		logger.Debugf("Callback function sending info %v to channel", *info)
	}
}
