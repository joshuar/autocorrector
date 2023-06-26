// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package corrections

import (
	"sync"

	"github.com/adrg/xdg"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Corrections struct {
	correctionsList map[string]string
	mu              sync.Mutex
}

func (c *Corrections) updateCorrections() {
	c.mu.Lock()
	viper.Unmarshal(&c.correctionsList)
	// check if any value is also a key
	// in this case, we'd end up with replacing the typo then replacing the replacement
	for _, v := range c.correctionsList {
		found := viper.GetString(v)
		if found != "" {
			log.Warn().Msgf("A replacement (%s) in the config is also listed as a typo. Deleting it to avoid recursive error.", v)
			delete(c.correctionsList, found)
		}
	}
	c.mu.Unlock()
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
func NewCorrections() *Corrections {
	viper.SetConfigName("corrections")
	viper.SetConfigType("toml")
	viper.AddConfigPath(xdg.ConfigHome + "/autocorrector")
	viper.AddConfigPath("/usr/share/autocorrector")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal().Msgf("Could not find config file: %s", viper.ConfigFileUsed())
		} else {
			log.Fatal().Err(err).Msg("Fatal error config file.")
		}
	}
	log.Debug().Caller().
		Msgf("Using corrections config at %s", viper.ConfigFileUsed())
	corrections := &Corrections{
		correctionsList: make(map[string]string),
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debug().Caller().
			Msgf("Config file %s has changed, getting updates.", viper.ConfigFileUsed())
		corrections.updateCorrections()
	})
	viper.WatchConfig()
	corrections.updateCorrections()
	return corrections
}
