// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package corrections

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
)

const (
	correctionsFilename = "corrections.toml"
)

type Corrections struct {
	correctionsList map[string]string
	mu              sync.Mutex
}

func (c *Corrections) CheckWord(word string) (string, bool) {
	c.mu.Lock()
	correction, ok := c.correctionsList[word]
	c.mu.Unlock()
	return correction, ok
}

// NewCorrections creates channels for sending words to check for corrections
// (and signalling a config file reload) as well as a channel for recieving
// corrected words
func NewCorrections() (*Corrections, error) {
	var correctionsFile string
	var c []byte
	var err error
	corrections := &Corrections{
		correctionsList: make(map[string]string),
	}

	correctionsFile = filepath.Join(os.Getenv("HOME"), ".config/autocorrector/", correctionsFilename)

	c, err = os.ReadFile(correctionsFile)
	if err != nil {
		log.Warn().Err(err).Msg("Could not open personal corrections file. Will try system-wide one.")
	}
	correctionsFile = filepath.Join("/usr/share/autocorrector", correctionsFilename)
	c, err = os.ReadFile(correctionsFile)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(c, &corrections.correctionsList)
	if err != nil {
		return nil, err
	}

	log.Info().Str("file", correctionsFile).Msg("Opened corrections file.")
	return corrections, nil
}
