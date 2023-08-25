// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package stats

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
