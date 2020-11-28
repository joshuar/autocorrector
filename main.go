package main

import (
	"regexp"
	"sort"
	"strings"

	"github.com/joshuar/autocorrector/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-vgo/robotgo"
)

const configName = "autocorrector"
const configType = "toml"

// keyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type keyTracker struct {
	key       chan rune
	wordDelim chan bool
	lineDelim chan bool
	backspace chan bool
}

// snoopKeys listens for key presses and fires on the appropriate channel
func (kt *keyTracker) snoopKeys() {
	// wordChar represents any standard character that would make up part of a word
	wordChar, _ := regexp.Compile("[[:alnum:]']")
	// wordDelim represents punctauation and space characters that indicate the end of a word
	wordDelim, _ := regexp.Compile("[[:punct:][:blank:]]")
	// lineDeline are linefeed/return characters indicating a new line was started
	lineDelim, _ := regexp.Compile("[\n\r\f]")
	// otherControlKey are the raw keycodes for various navigational keys like home, end, pgup, pgdown
	// and the arrow keys.
	otherControlKey := []int{65360, 65361, 65362, 65363, 65364, 65367, 65365, 65366}

	kbdEvents := robotgo.EventStart()
	defer close(kbdEvents)

	log.Info("Listening for keypresses...")
	// here we listen for key presses and match the key pressed against the regex patterns or raw keycodes above
	// depending on what key was pressed, we fire on the appropriate channel to do something about it
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

// newKeyTracker creates a new keyTracker struct
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

// wordStats stores counters for words checked and words corrected
type wordStats struct {
	wordsChecked   int
	wordsCorrected int
}

// addChecked will increment the words checked counter in a wordStats struct
func (w *wordStats) addChecked() {
	w.wordsChecked++
}

// addCorrected will increment the words corrected counter in a wordStats struct
func (w *wordStats) addCorrected() {
	w.wordsCorrected++
}

// calcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
func (w *wordStats) calcAccuracy() float32 {
	return (1 - float32(w.wordsCorrected)/float32(w.wordsChecked)) * 100
}

// newWordStats creates a new wordStats struct
func newWordStats() *wordStats {
	w := wordStats{
		wordsChecked:   0,
		wordsCorrected: 0,
	}
	return &w
}

func main() {
	cmd.Execute()
	kt := newKeyTracker()
	go slurpWords(kt)
	kt.snoopKeys()
}

// slurpWords listens for key press events and handles appropriately
// func slurpWords(kt *keyTracker, replacements *viper.Viper) {
func slurpWords(kt *keyTracker) {
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
			// go checkWord(word, delim, replacements, stats)
			go checkWord(word, delim, stats)
			word = nil
		// got the line delim or navigational key, clear the current word
		case <-kt.lineDelim:
			word = nil
		}

	}

}

// checkWord takes a typed word and looks up whether there is a replacement for it
// func checkWord(word []string, delim string, replacements *viper.Viper, stats *wordStats) {
func checkWord(word []string, delim string, stats *wordStats) {
	wordToCheck := strings.Join(word, "")
	stats.addChecked()
	replacement := viper.GetString(wordToCheck)
	if replacement != "" {
		// A replacement was found!
		log.Debug("Found replacement for ", wordToCheck, ": ", replacement)
		// Update our stats.
		stats.addCorrected()
		// Erase the existing word.
		eraseWord(len(word))
		// Insert the replacement.
		replaceWord(replacement, delim)
	}
}

// eraseWord removes a typed word
func eraseWord(wordLen int) {
	// Effectively, hit backspace key for the length of the word.
	for i := 0; i <= wordLen; i++ {
		robotgo.KeyTap("backspace")
	}
}

// replaceWord types the replacement word
func replaceWord(word string, delim string) {
	// Type out the replacement and whatever delimiter was after it.
	robotgo.TypeStr(word)
	robotgo.KeyTap(delim)
}
