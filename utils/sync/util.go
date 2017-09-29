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

//Package sync contains a set of counters synchronizer related utilities
package sync

import (
	"errors"
	"fmt"
	log "github.com/ligato/cn-infra/logging/logrus"
	"sync/atomic"
	"time"
)

//WaitCounterMatch waits until counter matches
//an expected number during an specific time.
func WaitCounterMatch(duration time.Duration, expected uint32, counter *uint32) error {
	timeChan := time.After(duration)
	for {
		select {
		case <-timeChan:
			return errors.New("timer expired, expected: " + fmt.Sprint(expected) +
				"but actual: " + fmt.Sprint(atomic.LoadUint32(counter)))
		default:
			time.Sleep(time.Millisecond)
			actual := atomic.LoadUint32(counter)
			if expected == atomic.LoadUint32(counter) {
				return nil
			}
			log.DefaultLogger().Debugf("Waiting for %v number of advertized paths, actual %v", expected, actual)
		}
	}
}
