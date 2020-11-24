package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
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

func (kt *keyTracker) snoopKeys() {
	wordChar, _ := regexp.Compile("[[:alnum:]']")
	wordDelim, _ := regexp.Compile("[[:punct:][:blank:]]")
	lineDelim, _ := regexp.Compile("[\n\r\f]")
	otherControlKey := []int{65360, 65361, 65362, 65363, 65364, 65367, 65365, 65366}
	kbdEvents := robotgo.EventStart()
	defer close(kbdEvents)

	log.Info("Listening for keypresses...")
	for e := range kbdEvents {
		log.Debug("Got keypress: ", e.Keychar, " : ", string(e.Keychar))
		switch {
		case wordChar.MatchString(string(e.Keychar)):
			kt.key <- e.Keychar
		case wordDelim.MatchString(string(e.Keychar)):
			kt.key <- e.Keychar
			kt.wordDelim <- true
		case lineDelim.MatchString(string(e.Keychar)):
			kt.lineDelim <- true
		case e.Keychar == 8:
			kt.backspace <- true
		case sort.SearchInts(otherControlKey, int(e.Rawcode)) > 0:
			kt.lineDelim <- true
		default:
			log.Debugf("Unknown key pressed: %v", e)
		}
	}
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

type wordStats struct {
	wordsChecked   int
	wordsCorrected int
}

func (w *wordStats) addChecked() {
	w.wordsChecked++
}

func (w *wordStats) addCorrected() {
	w.wordsCorrected++
}

func (w *wordStats) calcAccuracy() float32 {
	return (1 - float32(w.wordsCorrected)/float32(w.wordsChecked)) * 100
}

func newWordStats() *wordStats {
	w := wordStats{
		wordsChecked:   0,
		wordsCorrected: 0,
	}
	return &w
}

func main() {
	//log.SetLevel(log.DebugLevel)
	go systray.Run(systrayOnReady, systrayOnExit)
	log.Info("Reading config...")
	config := readConfig()
	kt := newKeyTracker()
	go slurpWords(kt, &config)
	kt.snoopKeys()
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
	stats := newWordStats()
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
			go checkWord(word, delim, replacements, stats)
			word = nil
		// got the line delim key, clear the current word
		case <-kt.lineDelim:
			word = nil
		}

	}

}

func checkWord(word []string, delim string, replacements *viper.Viper, stats *wordStats) {
	wordToCheck := strings.Join(word, "")
	stats.addChecked()
	replacement := replacements.GetString(wordToCheck)
	if replacement != "" {
		log.Debug("Found replacement for ", wordToCheck, ": ", replacement)
		stats.addCorrected()
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
