package keytracker

import (
	"bytes"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/go-vgo/robotgo"
	"github.com/joshuar/autocorrector/internal/wordstats"
	hook "github.com/robotn/gohook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	letters     = [...]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	numbers     = [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	punctuation = [...]string{"-", "+", ",", ".", "/", "\\", "[", "]", "`", ";", "'", "space"}
	controls    = [...]string{"tab", "ctrl", "alt", "ralt", "shift", "rshift", "cmd", "rcmd", "enter", "up", "down", "left", "right"}
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
			robotgo.EventEnd()
		case <-kt.StartSnooping:
			log.Debug("Starting event tracking...")
			kt.events = robotgo.EventStart()
			kt.setupSnooping()
			go func() {
				<-robotgo.EventProcess(kt.events)
			}()
		}
	}
}

func (kt *KeyTracker) setupSnooping() {
	// listen for letter keys and pass them to the wordChar channel
	log.Debug("Adding hook for letter keys...")
	for i := 0; i < len(letters); i++ {
		robotgo.EventHook(hook.KeyDown, []string{letters[i]}, func(e hook.Event) {
			kt.wordChar <- e.Keychar
		})
	}
	// listen for number keys and pass them to the wordChar channel
	log.Debug("Adding hook for number keys...")
	for i := 0; i < len(numbers); i++ {
		robotgo.EventHook(hook.KeyDown, []string{numbers[i]}, func(e hook.Event) {
			kt.wordChar <- e.Keychar
		})
	}
	// listen for punctuation keys and pass them to the punctChar channel
	log.Debug("Adding hook for punctuation keys...")
	for i := 0; i < len(punctuation); i++ {
		robotgo.EventHook(hook.KeyDown, []string{punctuation[i]}, func(e hook.Event) {
			kt.punctChar <- e.Keychar
		})
	}
	// listen for control keys and pass them to the punctChar channel
	log.Debug("Adding hook for control keys...")
	for i := 0; i < len(controls); i++ {
		robotgo.EventHook(hook.KeyDown, []string{controls[i]}, func(e hook.Event) {
			kt.controlChar <- true
		})
	}
	// listen for backspace/delete and handle that
	log.Debug("Adding hook for delete/backspace keys...")
	robotgo.EventHook(hook.KeyDown, []string{"delete"}, func(e hook.Event) {
		kt.backspaceChar <- true
	})
}

// SlurpWords listens for key press events and handles appropriately
// func slurpWords(kt *keyTracker, replacements *viper.Viper) {
func (kt *KeyTracker) SlurpWords(stats *wordstats.WordStats) {
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
			w.delim = string(punct)
			w.correctWord(stats, corrections, kt.ShowCorrections, kt.StartSnooping, kt.StopSnooping)
		// got the line delim or navigational key, clear the current word
		case <-kt.controlChar:
			w.clear()
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
	charBuf    *bytes.Buffer
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

func (w *word) extract(corrections *corrections) {
	w.correction = corrections.findCorrection(w.charBuf.String())
}

func (w *word) clear() {
	w.clearBuf()
	w.delim = ""
	w.correction = ""
}

func (w *word) correctWord(stats *wordstats.WordStats, corrections *corrections, showCorrections bool, startSnooping chan bool, stopSnooping chan bool) {
	if w.charBuf.Len() > 0 {
		w.extract(corrections)
		if w.correction != "" {
			// Update our stats.
			go stats.AddCorrected(w.charBuf.String(), w.correction)
			// stop key snooping
			stopSnooping <- true
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word plus the punctuation mark.
			log.Debugf("Making correction %v -> %v", w.charBuf.String(), w.correction)
			for i := 0; i <= w.charBuf.Len(); i++ {
				robotgo.KeyTap("backspace")
			}
			// Insert the replacement.
			// Type out the replacement and whatever delimiter was after it.
			robotgo.TypeStr(w.correction)
			robotgo.KeyTap(w.delim)
			// restart key snooping
			startSnooping <- true
			if showCorrections {
				beeep.Notify("Correction!", fmt.Sprintf("Replaced %s with %s", w.charBuf.String(), w.correction), "")
			}
		} else {
			go stats.AddChecked(w.charBuf.String())
		}
		w.clear()
	}
}

func newWord() *word {
	return &word{
		charBuf:    new(bytes.Buffer),
		delim:      "",
		correction: "",
	}
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
