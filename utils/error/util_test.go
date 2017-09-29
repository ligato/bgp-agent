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

//Package error_test contains tests for a set of error related utilities
package error_test

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"pantheon.tech/ligato-bgp/agent/utils/error"
	"testing"
)

func TestUtilsPanic(t *testing.T) {
	assert := assert.New(t)
	if !assert.Panics(func() {
		error.PanicIfError(errors.New("test error"))
	}) {
		t.Error("Panics should return true")
	}
}

func TestUtilsNoPanic(t *testing.T) {
	assert := assert.New(t)
	if !assert.NotPanics(func() {
		error.PanicIfError(nil)
	}) {
		t.Error("Not Panics should return true")
	}
}
