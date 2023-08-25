// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package wordstats

import "sync/atomic"

type SessionStats struct {
	KeysPressed      counter
	BackSpacePressed counter
}

func (s *SessionStats) Efficiency() float64 {
	if s.KeysPressed.Get() == 0 {
		return 0
	}
	keys := s.KeysPressed.Get()
	bs := s.BackSpacePressed.Get()
	return float64(bs) / float64(keys) * 100
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
