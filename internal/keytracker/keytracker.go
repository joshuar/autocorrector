package keytracker

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gen2brain/beeep"
	"github.com/go-vgo/robotgo"
	"github.com/joshuar/autocorrector/internal/wordstats"
	hook "github.com/robotn/gohook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	Key             chan rune
	WordDelim       chan bool
	LineDelim       chan bool
	Backspace       chan bool
	Disabled        bool
	ShowCorrections bool
	events          chan hook.Event
}

// SnoopKeys listens for key presses and fires on the appropriate channel
func (kt *KeyTracker) SnoopKeys() {
	// wordChar represents any standard character that would make up part of a word
	wordChar, _ := regexp.Compile("[[:alnum:]']")
	// wordDelim represents punctauation and space characters that indicate the end of a word
	wordDelim, _ := regexp.Compile("[[:punct:][:blank:]]")
	// lineDeline are linefeed/return characters indicating a new line was started
	lineDelim, _ := regexp.Compile("[\n\r\f]")
	// otherControlKey are the raw keycodes for various navigational keys like home, end, pgup, pgdown
	// and the arrow keys.
	otherControlKey := []int{65360, 65361, 65362, 65363, 65364, 65367, 65365, 65366}

	kt.events = robotgo.EventStart()

	// here we listen for key presses and match the key pressed against the regex patterns or raw keycodes above
	// depending on what key was pressed, we fire on the appropriate channel to do something about it
	for e := range kt.events {
		if !kt.Disabled {
			log.Debug("Got keypress: ", e.Keychar, " : ", string(e.Keychar))
			switch {
			case wordChar.MatchString(string(e.Keychar)):
				kt.Key <- e.Keychar
			case wordDelim.MatchString(string(e.Keychar)):
				kt.Key <- e.Keychar
				kt.WordDelim <- true
			case lineDelim.MatchString(string(e.Keychar)):
				kt.LineDelim <- true
			case e.Keychar == 8:
				kt.Backspace <- true
			case sort.SearchInts(otherControlKey, int(e.Rawcode)) > 0:
				kt.LineDelim <- true
			default:
				log.Debugf("Unknown key pressed: %v", e)
			}
		}
	}
}

// SlurpWords listens for key press events and handles appropriately
// func slurpWords(kt *keyTracker, replacements *viper.Viper) {
func (kt *KeyTracker) SlurpWords(st *wordstats.WordStats) {
	var word []string
	for {
		select {
		// got a letter or apostrophe key, append to create a word
		case key := <-kt.Key:
			word = append(word, string(key))
		case <-kt.Backspace:
			if len(word) > 0 {
				word = word[:len(word)-1]
			}
		// got a word delim key, we've got a word, find a replacement
		case <-kt.WordDelim:
			if len(word) > 0 {
				delim := word[len(word)-1]
				word = word[:len(word)-1]
				go kt.processWord(word, delim, st)
			}
			word = nil
		// got the line delim or navigational key, clear the current word
		case <-kt.LineDelim:
			word = nil
		}
	}
}

// checkWord takes a typed word and looks up whether there is a replacement for it
// func checkWord(word []string, delim string, replacements *viper.Viper, stats *wordStats) {
func (kt *KeyTracker) processWord(word []string, delim string, stats *wordstats.WordStats) {
	wordToCheck := strings.Join(word, "")
	stats.AddChecked(wordToCheck)
	replacement := viper.GetString(wordToCheck)
	if replacement != "" {
		// A replacement was found!
		log.Debug("Found replacement for ", wordToCheck, ": ", replacement)
		// Update our stats.
		stats.AddCorrected(wordToCheck, replacement)
		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word.
		for i := 0; i <= len(word); i++ {
			robotgo.KeyTap("backspace")
		}
		// Insert the replacement.
		// Type out the replacement and whatever delimiter was after it.
		robotgo.TypeStr(replacement)
		robotgo.KeyTap(delim)
		if kt.ShowCorrections {
			beeep.Alert("Correction!", fmt.Sprintf("Replaced %s with %s", wordToCheck, replacement), "")
		}
	}
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	k := make(chan rune)
	w := make(chan bool)
	l := make(chan bool)
	b := make(chan bool)
	d := false
	kt := KeyTracker{
		Key:       k,
		WordDelim: w,
		LineDelim: l,
		Backspace: b,
		Disabled:  d,
	}
	return &kt
}

func (k *KeyTracker) CloseKeyTracker() {
	close(k.Key)
	close(k.WordDelim)
	close(k.LineDelim)
	close(k.Backspace)
}
