package corrections

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type corrections struct {
	correctionList    map[string]string
	updateCorrections chan bool
}

func (c *corrections) FindCorrection(misspelling string) string {
	return c.correctionList[misspelling]
}

func (c *corrections) checkConfig() {
	// check if any value is also a key
	// in this case, we'd end up with replacing the typo then replacing the replacement
	c.correctionList = make(map[string]string)
	viper.Unmarshal(&c.correctionList)
	for _, v := range c.correctionList {
		found := viper.GetString(v)
		if found != "" {
			log.Warnf("A replacement (%s) in the config is also listed as a typo. Deleting it to avoid recursive error.", v)
			delete(c.correctionList, found)
		}
	}
	log.Debug("Config looks okay.")
}

func NewCorrections() *corrections {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Could not find config file: ", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Errorf("fatal error config file: %s", err))
		}
	}
	log.Debugf("Using corrections config at %s", viper.ConfigFileUsed())
	corrections := &corrections{
		updateCorrections: make(chan bool),
	}
	corrections.checkConfig()
	go func() {
		for {
			switch {
			case <-corrections.updateCorrections:
				corrections.checkConfig()
			}
		}

	}()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debugf("Config file %s has changed, getting updates.", viper.ConfigFileUsed())
		corrections.updateCorrections <- true
	})
	viper.WatchConfig()
	return corrections
}
