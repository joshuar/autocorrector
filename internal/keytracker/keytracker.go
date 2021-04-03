package keytracker

import (
	"bytes"
	"fmt"
	"time"
	"unicode"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/joshuar/autocorrector/internal/wordstats"
	"github.com/joshuar/go-linuxkeyboard/pkg/LinuxKeyboard"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type State string

const (
	start     State = "started"
	pause     State = "paused"
	resume    State = "resumed"
	terminate State = "terminated"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	events          *LinuxKeyboard.LinuxKeyboard
	ShowCorrections bool
	Signaller       chan State
}

func (kt *KeyTracker) EventWatcher(wordStats *wordstats.WordStats) {
	go kt.handler(kt.Signaller, wordStats)
	kt.Signaller <- start
}

func (kt *KeyTracker) handler(signaller chan State, wordStats *wordstats.WordStats) {
	done := make(chan struct{})

	for {
		signal := <-signaller

		switch signal {
		case resume:
			log.Debug("Resuming...")
			go kt.snoopKeys(wordStats, done)
		case pause:
			log.Debug("Pausing...")
			done <- struct{}{}
		case start:
			log.Debugf("Starting")
			go kt.snoopKeys(wordStats, done)
		case terminate:
			log.Debug("Terminating...")
			done <- struct{}{}
			return
		default:
			log.Println("unknown signal")
			return
		}
	}
}

func (kt *KeyTracker) snoopKeys(stats *wordstats.WordStats, done <-chan struct{}) {
	word := newWord()
	for {
		select {
		case <-done:
			log.Debug("Stop snooping keys...")
			kt.events.StopSnooping()
			return
		default:
			log.Debug("Start snooping keys...")
			ev := kt.events.StartSnooping()
			for e := range ev {
				switch {
				case e.Key.IsKeyPress():
					log.Debugf("Pressed key -- value: %d code: %d type: %d string: %s rune: %d (%c)", e.Key.Value, e.Key.Code, e.Key.Type, e.AsString, e.AsRune, e.AsRune)
					switch {
					case e.AsRune == rune('\b'):
						word.removeBuf()
					case unicode.IsDigit(e.AsRune) || unicode.IsLetter(e.AsRune):
						word.appendBuf(e.AsRune)
					case unicode.IsPunct(e.AsRune) || unicode.IsSymbol(e.AsRune) || unicode.IsSpace(e.AsRune):
						if word.getLength() > 0 {
							typo := word.getString()
							punct := e.AsRune
							correction := kt.checkWord(typo, stats)
							if correction != "" {
								stats.AddCorrected(typo, correction)
								kt.correctWord(typo, correction, punct)
							} else {
								stats.AddChecked(typo)
							}
							word.clear()
						}
					case e.AsString == "L_CTRL" || e.AsString == "R_CTRL" || e.AsString == "L_ALT" || e.AsString == "R_ALT" || e.AsString == "L_META" || e.AsString == "R_META":
						<-ev
						word.clear()
					case e.AsString == "L_SHIFT" || e.AsString == "R_SHIFT":
					// case unicode.IsPrint(e.AsRune):
					// 	log.Debugf("Got unhandled printable character: (rune %d, string %c, unicode %U", e.AsRune, e.AsRune, e.AsRune)
					default:
						word.clear()
					}
				case e.Key.IsKeyRelease():
					log.Debugf("Released key -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
				default:
					log.Debugf("Other event -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
				}
			}
		}
	}
}

func (kt *KeyTracker) checkWord(word string, stats *wordstats.WordStats) string {
	corrections := newCorrections()
	return corrections.findCorrection(word)
}

func (kt *KeyTracker) correctWord(word string, correction string, punctuation rune) {
	// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
	// TODO: use an accurate number for the latency
	time.Sleep(15 * time.Millisecond)
	// Erase the existing word.
	// Effectively, hit backspace key for the length of the word plus the punctuation mark.
	log.Debugf("Making correction %s to %s", word, correction)
	for i := 0; i <= len(word); i++ {
		kt.events.TypeBackSpace()
	}
	// Insert the replacement.
	// Type out the replacement and whatever delimiter was after it.
	kt.events.TypeString(correction)
	kt.events.TypeString(string(punctuation))
	if kt.ShowCorrections {
		beeep.Notify("Correction!", fmt.Sprintf("Replaced %s with %s", word, correction), "")
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	ch := make(chan struct{})
	close(ch)
	return &KeyTracker{
		events:          LinuxKeyboard.NewLinuxKeyboard(LinuxKeyboard.FindKeyboardDevice()),
		ShowCorrections: false,
		Signaller:       make(chan State),
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
			log.Fatal(fmt.Errorf("fatal error config file: %s", err))
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
			switch {
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
