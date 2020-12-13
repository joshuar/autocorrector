package main

import (
	"strings"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/joshuar/autocorrector/cmd"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/joshuar/autocorrector/internal/wordstats"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/go-vgo/robotgo"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	cmd.Execute()
	keyTracker := keytracker.NewKeyTracker()
	wordStats := wordstats.NewWordStats()

	systray.SetIcon(icon.Data)
	systray.SetTitle("Autocorrector")
	systray.SetTooltip("Autocorrector corrects your typos")
	mQuit := systray.AddMenuItem("Quit", "Quit Autocorrector")
	mEnabled := systray.AddMenuItemCheckbox("Enabled", "Enable Autocorrector", true)

	go slurpWords(keyTracker, wordStats)
	go keyTracker.SnoopKeys()

	for {
		select {
		case <-mEnabled.ClickedCh:
			if mEnabled.Checked() {
				mEnabled.Uncheck()
				log.Info("Disabling Autocorrector")
				keyTracker.Disabled = true
			} else {
				mEnabled.Check()
				log.Info("Enabling Autocorrector")
				keyTracker.Disabled = false
			}
		case <-mQuit.ClickedCh:
			log.Info("Requesting quit")
			systray.Quit()
		}
	}
}

func onExit() {
	// clean up here
}

// SlurpWords listens for key press events and handles appropriately
// func slurpWords(kt *keyTracker, replacements *viper.Viper) {
func slurpWords(kt *keytracker.KeyTracker, st *wordstats.WordStats) {
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
			delim := word[len(word)-1]
			word = word[:len(word)-1]
			go processWord(word, delim, st)
			word = nil
		// got the line delim or navigational key, clear the current word
		case <-kt.LineDelim:
			word = nil
		}

	}

}

// checkWord takes a typed word and looks up whether there is a replacement for it
// func checkWord(word []string, delim string, replacements *viper.Viper, stats *wordStats) {
func processWord(word []string, delim string, stats *wordstats.WordStats) {
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
	}
}
