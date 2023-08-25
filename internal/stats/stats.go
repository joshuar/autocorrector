// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package stats

import "sync/atomic"

type Stats struct {
	SessionStats
}

type counter struct {
	value uint64
}

func (c *counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c *counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}
