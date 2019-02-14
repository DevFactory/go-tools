/* Copyright 2019 DevFactory FZ LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package time_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	etime "github.com/DevFactory/go-tools/pkg/extensions/time"
)

func TestRefresher(t *testing.T) {
	testChan := make(chan time.Time)
	actionCounter := 0
	actionDone := make(chan int, 1)
	action := func() {
		actionCounter++
		actionDone <- actionCounter
	}
	r := etime.NewRefresher(testChan, action, true, true)
	// check if call is idempotent - start should be already true
	r.StartRefreshing()

	// assert that onCreate action() was run
	cnt := <-actionDone
	assert.Equal(t, 1, r.GetRefreshCount())
	assert.Equal(t, 1, cnt)
	assert.True(t, r.IsRefreshing())

	// assert refresh on channel event
	testChan <- time.Now()
	cnt = <-actionDone
	assert.Equal(t, 2, r.GetRefreshCount())
	assert.Equal(t, 2, cnt)
	assert.True(t, r.IsRefreshing())

	// stop refreshing
	r.StopRefreshing()
	// check if call is idempotent
	r.StopRefreshing()
	assert.False(t, r.IsRefreshing())
}

func TestCreateTickerRefresher(t *testing.T) {
	actionCounter := 0
	actionDone := make(chan int, 1)
	action := func() {
		actionCounter++
		actionDone <- actionCounter
	}
	r := etime.NewTickerRefresher(action, true, true, time.Second)
	cnt := <-actionDone
	assert.NotNil(t, r)
	assert.Equal(t, 1, r.GetRefreshCount())
	assert.Equal(t, 1, cnt)
	assert.True(t, r.IsRefreshing())
}
