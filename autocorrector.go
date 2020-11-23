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

type keyTracker struct {
	key       chan rune
	wordDelim chan bool
	lineDelim chan bool
	backspace chan bool
}

func newKeyTracker() *keyTracker {
	k := make(chan rune)
	w := make(chan bool)
	l := make(chan bool)
	b := make(chan bool)
	kt := keyTracker{
		key:       k,
		wordDelim: w,
		lineDelim: l,
		backspace: b,
	}
	return &kt
}

func main() {
	//log.SetLevel(log.DebugLevel)
	go systray.Run(systrayOnReady, systrayOnExit)
	log.Info("Reading config...")
	config := readConfig()
	kt := newKeyTracker()
	go slurpWords(kt, &config)
	snoopKeys(kt)
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

func slurpWords(kt *keyTracker, replacements *viper.Viper) {
	var word []string
	for {
		select {
		// got a letter or apostrophe key, append to create a word
		case key := <-kt.key:
			word = append(word, string(key))
		case <-kt.backspace:
			if len(word) > 0 {
				word = word[:len(word)-1]
			}
		// got a word delim key, we've got a word, find a replacement
		case <-kt.wordDelim:
			delim := word[len(word)-1]
			word = word[:len(word)-1]
			go checkWord(word, delim, replacements)
			word = nil
		// got the line delim key, clear the current word
		case <-kt.lineDelim:
			word = nil
		}

	}

}

func checkWord(word []string, delim string, replacements *viper.Viper) {
	wordToCheck := strings.Join(word, "")
	replacement := replacements.GetString(wordToCheck)
	if replacement != "" {
		log.Debug("Found replacement for ", wordToCheck, ": ", replacement)
		eraseWord(len(word))
		replaceWord(replacement, delim)
	}
}

func eraseWord(wordLen int) {
	for i := 0; i <= wordLen; i++ {
		robotgo.KeyTap("backspace")
	}
}

func replaceWord(word string, delim string) {
	robotgo.TypeStr(word)
	robotgo.KeyTap(delim)
}

func snoopKeys(kt *keyTracker) {
	wordChar, _ := regexp.Compile("[[:alnum:]']")
	wordDelim, _ := regexp.Compile("[[:punct:][:blank:]]")
	lineDelim, _ := regexp.Compile("[\n\r\f]")
	evChan := robotgo.EventStart()
	defer close(evChan)
	log.Info("Listening for keypresses...")
	for e := range evChan {
		log.Debug("Got keypress: ", e.Keychar, " : ", string(e.Keychar))
		// any regular key pressed, slurp that up to form a word
		if wordChar.MatchString(string(e.Keychar)) {
			kt.key <- e.Keychar
		}
		// word delim key pressed, check the word
		if wordDelim.MatchString(string(e.Keychar)) { //32
			kt.key <- e.Keychar
			kt.wordDelim <- true
		}
		// line delim key pressed, clear the current slurping
		if lineDelim.MatchString(string(e.Keychar)) { //13
			kt.lineDelim <- true
		}
		if e.Keychar == 8 {
			kt.backspace <- true
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
