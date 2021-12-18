package corrections

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type corrections struct {
	correctionList map[string]string
	controlCh      chan interface{}
	correctionCh   chan string
}

func (c *corrections) findCorrection(misspelling string) string {
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

func (c *corrections) handler() {
	for data := range c.controlCh {
		switch d := data.(type) {
		case bool:
			c.checkConfig()
		case string:
			log.Debugf("Checking %s for correction", d)
			if found := c.findCorrection(d); found != "" {
				log.Debugf("Found! %s", found)
				c.correctionCh <- found
			}
		default:
			log.Debugf("Unknown data %T received: %v", d, d)
		}
	}
}

func NewCorrections() (chan interface{}, chan string) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Could not find config file: ", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Errorf("fatal error config file: %s", err))
		}
	}
	log.Debugf("Using corrections config at %s", viper.ConfigFileUsed())
	corrections := &corrections{
		controlCh:    make(chan interface{}),
		correctionCh: make(chan string),
	}
	corrections.checkConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debugf("Config file %s has changed, getting updates.", viper.ConfigFileUsed())
		corrections.controlCh <- true
	})
	viper.WatchConfig()
	go corrections.handler()
	return corrections.controlCh, corrections.correctionCh
}
