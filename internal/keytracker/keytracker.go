package keytracker

import (
	"bytes"
	"fmt"
	"unicode"

	"github.com/fsnotify/fsnotify"
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
	StopSnooping    chan bool
	StartSnooping   chan bool
	ShowCorrections bool
}

// SnoopKeys listens for key presses and fires on the appropriate channel
func (kt *KeyTracker) SnoopKeys() {
	for {
		select {
		case <-kt.StopSnooping:
			log.Debug("Stopping event tracking...")
			hook.End()
		case <-kt.StartSnooping:
			log.Debug("Starting event tracking...")
			kt.events = hook.Start()
			go func() {
				for e := range kt.events {
					switch {
					case e.Keychar == rune('\b'):
						kt.backspaceChar <- true
					case unicode.IsDigit(e.Keychar) || unicode.IsLetter(e.Keychar):
						kt.wordChar <- e.Keychar
					case unicode.IsPunct(e.Keychar) || unicode.IsSymbol(e.Keychar) || unicode.IsSpace(e.Keychar):
						kt.punctChar <- e.Keychar
					case unicode.IsControl(e.Keychar):
						kt.controlChar <- true
					case unicode.IsPrint(e.Keychar):
						log.Debugf("Got unhandled printable character: %v", string(e.Keychar))
					default:
					}
				}
			}()
		}
	}
}

// SlurpWord slurps up key presses into words where appropriate
func (kt *KeyTracker) SlurpWord(stats *wordstats.WordStats) {
	corrections := newCorrections()
	w := newWord()
	for {
		select {
		// got a letter or apostrophe key, append to create a word
		case key := <-kt.wordChar:
			w.appendBuf(key)
		// got the backspace key, remove last character from the buffer
		case <-kt.backspaceChar:
			log.Debug("removing char")
			w.removeBuf()
		// got a word delim key, we've got a word, find a replacement
		case punct := <-kt.punctChar:
			// don't do anything if it's just a punctuation mark only
			if w.getLength() > 0 {
				log.Debugf("Checking word %v for corrections...", w.getString())
				correction := corrections.findCorrection(w.getString())
				kt.correctWord(stats, w.getString(), correction, w.getLength(), string(punct))
				w.clear()
			}
		// got the line delim or navigational key, clear the current word
		case <-kt.controlChar:
			w.clear()
		}
	}
}

func (kt *KeyTracker) correctWord(stats *wordstats.WordStats, word string, correction string, length int, punctuation string) {
	if correction != "" {
		// Update our stats.
		go stats.AddCorrected(word, correction)
		// stop key snooping
		kt.StopSnooping <- true
		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word plus the punctuation mark.
		log.Debugf("Making correction %v to %v", word, correction)
		for i := 0; i <= length; i++ {
			robotgo.KeyTap("backspace")
		}
		// Insert the replacement.
		// Type out the replacement and whatever delimiter was after it.
		robotgo.TypeStr(correction)
		robotgo.KeyTap(punctuation)
		// restart key snooping
		kt.StartSnooping <- true
		if kt.ShowCorrections {
			beeep.Notify("Correction!", fmt.Sprintf("Replaced %s with %s", word, correction), "")
		}
	} else {
		go stats.AddChecked(word)
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
		events:          make(chan hook.Event),
		wordChar:        make(chan rune),
		punctChar:       make(chan rune),
		controlChar:     make(chan bool),
		backspaceChar:   make(chan bool),
		StartSnooping:   make(chan bool),
		StopSnooping:    make(chan bool),
		ShowCorrections: false,
	}
}

type word struct {
	charBuf *bytes.Buffer
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

func (w *word) getString() string {
	return w.charBuf.String()
}

func (w *word) getLength() int {
	return w.charBuf.Len()
}

func (w *word) clear() {
	w.clearBuf()
}

func newWord() *word {
	return &word{charBuf: new(bytes.Buffer)}
}

type corrections struct {
	correctionList    map[string]string
	updateCorrections chan bool
}

func (c *corrections) findCorrection(mispelling string) string {
	return c.correctionList[mispelling]
}

func (c *corrections) checkConfig() {
	// check if any value is also a key
	// in this case, we'd end up with replacing the typo then replacing the replacement
	configMap := make(map[string]string)
	viper.Unmarshal(&configMap)
	for _, v := range configMap {
		found := viper.GetString(v)
		if found != "" {
			log.Fatalf("A replacement in the config is also listed as a typo (%v)  This won't work.", v)
		}
	}
	log.Debug("Config looks okay.")
}

func newCorrections() *corrections {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Could not find config file: ", viper.ConfigFileUsed())
		} else {
			log.Fatal(fmt.Errorf("Fatal error config file: %s", err))
		}
	}
	corrections := &corrections{
		correctionList:    make(map[string]string),
		updateCorrections: make(chan bool),
	}
	corrections.checkConfig()
	viper.Unmarshal(&corrections.correctionList)
	go func() {
		for {
			select {
			case <-corrections.updateCorrections:
				corrections.checkConfig()
				viper.Unmarshal(&corrections.correctionList)
				log.Debug("Updated corrections from config file.")
			}
		}

	}()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debug("Config file has changed.")
		corrections.updateCorrections <- true
	})
	viper.WatchConfig()
	return corrections
}
