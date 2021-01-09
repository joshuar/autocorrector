package keytracker

import (
	"bytes"
	"fmt"
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
	events          chan hook.Event
	wordChar        chan rune
	punctChar       chan rune
	controlChar     chan bool
	backspaceChar   chan bool
	Disabled        bool
	ShowCorrections bool
}

// SnoopKeys listens for key presses and fires on the appropriate channel
func (kt *KeyTracker) SnoopKeys() {

	// here we listen for key presses and match the key pressed against the regex patterns or raw keycodes above
	// depending on what key was pressed, we fire on the appropriate channel to do something about it
	for e := range kt.events {
		if !kt.Disabled {
			switch {
			case e.Keychar == rune('\b'):
				kt.backspaceChar <- true
			case unicode.IsDigit(e.Keychar) || unicode.IsLetter(e.Keychar):
				kt.wordChar <- e.Keychar
			case unicode.IsPunct(e.Keychar) || unicode.IsSpace(e.Keychar):
				kt.punctChar <- e.Keychar
			case unicode.IsControl(e.Keychar):
				kt.controlChar <- true
			default:
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
		case key := <-kt.wordChar:
			w.appendBuf(key)
		// got the backspace key, remove last character from the buffer
		case <-kt.backspaceChar:
			w.removeBuf()
		// got a word delim key, we've got a word, find a replacement
		case punct := <-kt.punctChar:
			w.delim = string(punct)
			w.correctWord(st, kt.ShowCorrections)
		// got the line delim or navigational key, clear the current word
		case <-kt.controlChar:
			w.clearBuf()
		}
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	close(kt.wordChar)
	close(kt.punctChar)
	close(kt.controlChar)
	close(kt.backspaceChar)
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	return &KeyTracker{
		events:          robotgo.EventStart(),
		wordChar:        make(chan rune),
		punctChar:       make(chan rune),
		controlChar:     make(chan bool),
		backspaceChar:   make(chan bool),
		Disabled:        false,
		ShowCorrections: false,
	}
}

type word struct {
	charBuf    *bytes.Buffer
	asString   string
	length     int
	delim      string
	correction string
}

func (w *word) clearBuf() {
	w.charBuf.Reset()
}

func (w *word) appendBuf(char rune) {
	w.charBuf.WriteRune(char)
}

func (w *word) removeBuf() {
	if w.charBuf.Len() > 0 {
		w.charBuf.Truncate(w.charBuf.Len() - 1)
	}
}

func (w *word) extract() {
	corrections := newCorrections()
	w.asString = w.charBuf.String()
	w.length = w.charBuf.Len()
	w.correction = corrections.findCorrection(w.asString)
	w.clearBuf()
}

func (w *word) correctWord(stats *wordstats.WordStats, showCorrections bool) {
	if w.charBuf.Len() > 0 {
		w.extract()
		if w.correction != "" {
			// A replacement was found!
			log.Debug("Found replacement for ", w.asString, ": ", w.correction)
			// Update our stats.
			go stats.AddCorrected(w.asString, w.correction)
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word.
			for i := 0; i <= w.length; i++ {
				robotgo.KeyTap("backspace")
			}
			// Insert the replacement.
			// Type out the replacement and whatever delimiter was after it.
			robotgo.TypeStr(w.correction)
			robotgo.KeyTap(w.delim)
			if showCorrections {
				beeep.Alert("Correction!", fmt.Sprintf("Replaced %s with %s", w.asString, w.correction), "")
			}
		}
	}
}

func newWord() *word {
	return &word{
		charBuf:    new(bytes.Buffer),
		asString:   "",
		length:     0,
		delim:      "",
		correction: "",
	}
}

type corrections struct {
	correctionList map[string]string
}

func (c *corrections) findCorrection(mispelling string) string {
	return c.correctionList[mispelling]
}

func newCorrections() *corrections {
	c := make(map[string]string)
	viper.Unmarshal(&c)
	return &corrections{correctionList: c}
}
