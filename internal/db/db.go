// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package db

import (
	"context"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Correction struct {
	Word, Correction string
}

type Counter struct {
	Value uint64
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.Value, 1)
}

func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.Value)
}

type Counters struct {
	WordsChecked     Counter
	WordsCorrected   Counter
	KeysPressed      Counter
	BackspacePressed Counter
}

func (c *Counters) Efficiency() float64 {
	if c.KeysPressed.Get() == 0 {
		return 0
	}
	return float64(c.BackspacePressed.Get()) / float64(c.KeysPressed.Get()) * 100
}

func (c *Counters) Accuracy() float64 {
	return (1 - float64(c.WordsCorrected.Get())/float64(c.WordsChecked.Get())) * 100
}

func (c *Counters) Write(file string) error {
	fs, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return err
	}

	enc := gob.NewEncoder(fs)
	err = enc.Encode(&c)
	if err != nil {
		return err

	}
	log.Debug().Msg("Wrote counters to disk.")
	return nil
}

func OpenCounters(file string) (*Counters, error) {
	log.Debug().Msgf("Using counters file %s", file)
	fs, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return nil, err
	}
	var counters Counters
	dec := gob.NewDecoder(fs)
	err = dec.Decode(&counters)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return &counters, nil
}

func OpenCorrections(file string) (zerolog.Logger, error) {
	log.Debug().Msgf("Using corrections log %s", file)
	fs, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0640)
	if err != nil {
		return zerolog.New(os.Stderr).With().Timestamp().Logger(), nil
	}
	return zerolog.New(fs).With().Timestamp().Logger(), nil
}

type Stats struct {
	counters                      *Counters
	corrections                   zerolog.Logger
	countersFile, correctionsFile string
	Corrected                     chan Correction
	Done                          chan struct{}
}

func (s *Stats) addCorrected() {
	for c := range s.Corrected {
		s.counters.WordsCorrected.Inc()
		s.corrections.Info().
			Str("word", c.Word).Str("correction", c.Correction).
			Msg("Corrected.")
	}
}

func (s *Stats) IncCheckedCounter() {
	s.counters.WordsChecked.Inc()
}

func (s *Stats) IncKeyCounter() {
	s.counters.KeysPressed.Inc()
}

func (s *Stats) IncBackspaceCounter() {
	s.counters.BackspacePressed.Inc()
}

func (s *Stats) GetCheckedTotal() uint64 {
	return s.counters.WordsChecked.Get()
}

func (s *Stats) GetCorrectedTotal() uint64 {
	return s.counters.WordsCorrected.Get()
}

func (s *Stats) GetKeysPressed() uint64 {
	return s.counters.KeysPressed.Get()
}

func (s *Stats) GetBackspacePressed() uint64 {
	return s.counters.BackspacePressed.Get()
}

func (s *Stats) GetAccuracy() float64 {
	return s.counters.Accuracy()
}

func (s *Stats) GetEfficiency() float64 {
	return s.counters.Efficiency()
}

func (s *Stats) Save() {
	if err := s.counters.Write(s.countersFile); err != nil {
		log.Warn().Err(err).Msg("Error saving stats.")
	}
}

func RunStats(ctx context.Context, path string) (*Stats, error) {
	s := &Stats{
		Corrected:       make(chan Correction, 1),
		Done:            make(chan struct{}),
		countersFile:    filepath.Join(path, "counters"),
		correctionsFile: filepath.Join(path, "corrections.log"),
	}
	c, err := OpenCounters(s.countersFile)
	if err != nil {
		return nil, errors.Join(errors.New("could not open counters file"), err)
	}
	s.counters = c
	l, err := OpenCorrections(s.correctionsFile)
	if err != nil {
		return nil, errors.Join(errors.New("could not open corrections file"), err)
	}
	s.corrections = l
	go s.addCorrected()
	return s, nil
}
