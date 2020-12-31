package keytracker

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

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
	// otherControlKey are the raw keycodes for various navigational keys like home, end, pgup, pgdown
	// and the arrow keys.
	otherControlKey := []int{65360, 65361, 65362, 65363, 65364, 65367, 65365, 65366}

	kt.events = robotgo.EventStart()

	// here we listen for key presses and match the key pressed against the regex patterns or raw keycodes above
	// depending on what key was pressed, we fire on the appropriate channel to do something about it
	for e := range kt.events {
		if !kt.Disabled {
			switch {
			case e.Keychar == 8:
				kt.Backspace <- true
			case unicode.IsDigit(e.Keychar) || unicode.IsLetter(e.Keychar):
				kt.Key <- e.Keychar
			case unicode.IsPunct(e.Keychar) || unicode.IsSpace(e.Keychar):
				kt.Key <- e.Keychar
				kt.WordDelim <- true
			case unicode.IsControl(e.Keychar):
				kt.LineDelim <- true
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
	w := newWord()
	for {
		select {
		// got a letter or apostrophe key, append to create a word
		case key := <-kt.Key:
			w.appendBuf(string(key))
		// got the backspace key, remove last character from the buffer
		case <-kt.Backspace:
			w.removeBuf()
		// got a word delim key, we've got a word, find a replacement
		case <-kt.WordDelim:
			w.processWord(st, kt.ShowCorrections)
		// got the line delim or navigational key, clear the current word
		case <-kt.LineDelim:
			w.clearBuf()
		}
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	close(kt.Key)
	close(kt.WordDelim)
	close(kt.LineDelim)
	close(kt.Backspace)
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

type word struct {
	charBuf  []string
	asString string
	length   int
	delim    string
}

func (w *word) clearBuf() {
	w.charBuf = nil
}

func (w *word) appendBuf(char string) {
	w.charBuf = append(w.charBuf, char)
}

func (w *word) removeBuf() {
	if len(w.charBuf) > 0 {
		w.charBuf = w.charBuf[:len(w.charBuf)-1]
	}
}

func (w *word) extract() {
	w.delim = w.charBuf[len(w.charBuf)-1]
	w.asString = strings.Join(w.charBuf[:len(w.charBuf)-1], "")
	w.length = len(w.asString)
	w.clearBuf()
}

func (w *word) processWord(stats *wordstats.WordStats, showCorrections bool) {
	if len(w.charBuf) > 0 {
		w.extract()
		replacement := viper.GetString(w.asString)
		if replacement != "" {
			// A replacement was found!
			log.Debug("Found replacement for ", w.asString, ": ", replacement)
			// Update our stats.
			go stats.AddCorrected(w.asString, replacement)
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word.
			for i := 0; i <= w.length; i++ {
				robotgo.KeyTap("backspace")
			}
			// Insert the replacement.
			// Type out the replacement and whatever delimiter was after it.
			robotgo.TypeStr(replacement)
			robotgo.KeyTap(w.delim)
			if showCorrections {
				beeep.Alert("Correction!", fmt.Sprintf("Replaced %s with %s", w.asString, replacement), "")
			}
		}
	}
}

func newWord() *word {
	return &word{
		charBuf:  make([]string, 0),
		asString: "",
		length:   0,
		delim:    "",
	}
}
