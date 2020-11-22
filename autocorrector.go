package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/go-vgo/robotgo"
	"github.com/spf13/viper"
)

const configName = "autocorrector"
const configType = "toml"

func main() {
	// log.SetLevel(log.DebugLevel)
	go systray.Run(systrayOnReady, systrayOnExit)
	log.Info("Reading config...")
	config := readConfig()
	keycode := make(chan rune)
	space := make(chan bool)
	control := make(chan bool)
	backspace := make(chan bool)
	go slurpWords(keycode, space, control, backspace, &config)
	snoopKeys(keycode, space, control, backspace)
}

func readConfig() viper.Viper {
	c := viper.New()
	c.SetConfigName(configName)
	c.SetConfigType(configType)
	c.AddConfigPath("$HOME/.config/autocorrector")
	c.AddConfigPath(".")
	err := c.ReadInConfig()
	if err != nil {
		log.Fatal(fmt.Errorf("fatal error config file: %s", err))
	}
	c.WatchConfig()
	return *c
}

func slurpWords(keychar chan rune, space chan bool, control chan bool, backspace chan bool, replacements *viper.Viper) {
	var word []string
	for {
		select {
		// got a regular key, append to create a word
		case key := <-keychar:
			word = append(word, string(key))
		case <-backspace:
			if len(word) > 0 {
				word = word[:len(word)-1]
			}
		// got the space key, we've got a word, find a replacement
		case <-space:
			go checkWord(word, replacements)
			word = nil
		// got the enter key, clear the current word
		case <-control:
			word = nil
		}

	}

}

func checkWord(word []string, replacements *viper.Viper) {
	wordToCheck := strings.Join(word, "")
	replacement := replacements.GetString(wordToCheck)
	if replacement != "" {
		log.Debug("Found replacement for ", wordToCheck, ": ", replacement)
		eraseWord(len(word))
		replaceWord(replacement)
	}
}

func eraseWord(wordLen int) {
	for i := 0; i <= wordLen; i++ {
		robotgo.KeyTap("backspace")
	}
}

func replaceWord(word string) {
	robotgo.TypeStr(word)
	robotgo.KeyTap("space")
}

func snoopKeys(keycode chan rune, space chan bool, control chan bool, backspace chan bool) {

	nonSpace, _ := regexp.Compile("[[:graph:]]+")
	evChan := robotgo.EventStart()
	defer close(evChan)
	log.Info("Listening for keypresses...")
	for e := range evChan {
		log.Debug("Got keypress: ", e.Keychar)
		// any regular key pressed, slurp that up to form a word
		if nonSpace.MatchString(string(e.Keychar)) {
			keycode <- e.Keychar
		}
		// space pressed, triggers replacement lookup
		if e.Keychar == 32 {
			space <- true
		}
		// enter pressed, clear the current slurping
		if e.Keychar == 13 {
			control <- true
		}
		if e.Keychar == 8 {
			backspace <- true
		}
	}
}

func systrayOnReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Autocorrector")
	systray.SetTooltip("Autocorrect words you type")
	mQuit := systray.AddMenuItem("Quit", "Quit the Autocorrector")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func systrayOnExit() {
	os.Exit(0)
}
