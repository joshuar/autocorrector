package main

import (
	"strings"

	"github.com/joshuar/autocorrector/cmd"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/joshuar/autocorrector/internal/wordstats"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-vgo/robotgo"
)

func main() {
	cmd.Execute()
	kt := keytracker.NewKeyTracker()
	go slurpWords(kt)
	kt.SnoopKeys()
}

// SlurpWords listens for key press events and handles appropriately
// func slurpWords(kt *keyTracker, replacements *viper.Viper) {
func slurpWords(kt *keytracker.KeyTracker) {
	var word []string
	stats := wordstats.NewWordStats()
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
			delim := word[len(word)-1]
			word = word[:len(word)-1]
			// go checkWord(word, delim, replacements, stats)
			go checkWord(word, delim, stats)
			word = nil
		// got the line delim or navigational key, clear the current word
		case <-kt.LineDelim:
			word = nil
		}

	}

}

// checkWord takes a typed word and looks up whether there is a replacement for it
// func checkWord(word []string, delim string, replacements *viper.Viper, stats *wordStats) {
func checkWord(word []string, delim string, stats *wordstats.WordStats) {
	wordToCheck := strings.Join(word, "")
	stats.AddChecked()
	replacement := viper.GetString(wordToCheck)
	if replacement != "" {
		// A replacement was found!
		log.Debug("Found replacement for ", wordToCheck, ": ", replacement)
		// Update our stats.
		stats.AddCorrected()
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
