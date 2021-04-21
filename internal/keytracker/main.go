package keytracker

import (
	"bytes"
	"fmt"
	"time"
	"unicode"

	"github.com/fsnotify/fsnotify"
	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/go-linuxkeyboard/pkg/LinuxKeyboard"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd             *LinuxKeyboard.LinuxKeyboard
	kbdEvents       chan LinuxKeyboard.KeyboardEvent
	typedWord       chan typed
	paused          bool
	ShowCorrections bool
	corrections     *corrections
	statsCh         chan *control.StatsMsg
}

// EventWatcher opens the stats database, starts a goroutine to "slurp" words,
// starts a goroutine to check for corrections and opens a socket for server
// control
func (kt *KeyTracker) EventWatcher(manager *control.ConnManager) {
	go kt.kbd.Snoop(kt.kbdEvents)
	go kt.slurpWords()
	go kt.checkWord()
	go sendStats(manager, kt.statsCh)
	// go sendNotifications(manager, kt.notifications)
	manager.SendState(&control.StateMsg{Start: true})
	for msg := range manager.Data {
		switch t := msg.(type) {
		case *control.StateMsg:
			switch {
			case t.Start:
				kt.start()
			case t.Pause:
				kt.pause()
			case t.Resume:
				kt.resume()
			}
		default:
			log.Debugf("Unhandled message recieved: %v", msg)
		}
	}
}

func (kt *KeyTracker) start() error {
	log.Info("Started checking words")
	kt.paused = false
	// kt.kbdEvents = make(chan LinuxKeyboard.KeyboardEvent)
	// go kt.slurpWords()
	// go kt.kbd.Snoop(kt.kbdEvents)
	return nil
}

func (kt *KeyTracker) pause() error {
	kt.paused = true
	log.Info("Pausing checking words...")
	// close(kt.kbdEvents)
	return nil
}

func (kt *KeyTracker) resume() error {
	return kt.start()
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	log.Debug("Started slurping words")
	for {
		e, ok := <-kt.kbdEvents
		if !ok {
			log.Debug("Stopped slurping words")
			charBuf.Reset()
			return
		}
		switch {
		case kt.paused:
			charBuf.Reset()
		case e.Key.IsKeyPress():
			log.Debugf("Pressed key -- value: %d code: %d type: %d string: %s rune: %d (%c)", e.Key.Value, e.Key.Code, e.Key.Type, e.AsString, e.AsRune, e.AsRune)
		case e.Key.IsKeyRelease():
			log.Debugf("Released key -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
			switch {
			case e.AsRune == rune('\b'):
				if charBuf.Len() > 0 {
					charBuf.Truncate(charBuf.Len() - 1)
				}
			case unicode.IsDigit(e.AsRune) || unicode.IsLetter(e.AsRune):
				charBuf.WriteRune(e.AsRune)
			case unicode.IsPunct(e.AsRune) || unicode.IsSymbol(e.AsRune) || unicode.IsSpace(e.AsRune):
				if charBuf.Len() > 0 {
					t := &typed{
						word:  charBuf.String(),
						punct: e.AsRune,
					}
					charBuf.Reset()
					kt.typedWord <- *t
				}
			case e.AsString == "L_CTRL" || e.AsString == "R_CTRL" || e.AsString == "L_ALT" || e.AsString == "R_ALT" || e.AsString == "L_META" || e.AsString == "R_META":
				// absorb the ctrl/alt/me	ta key and then reset the buffer
				<-kt.kbdEvents
				charBuf.Reset()
			case e.AsString == "L_SHIFT" || e.AsString == "R_SHIFT":
			default:
				charBuf.Reset()
			}
		default:
			log.Debugf("Other event -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
		}

	}
}

type typed struct {
	word  string
	punct rune
}

func (kt *KeyTracker) checkWord() {

	for typed := range kt.typedWord {
		correction := kt.corrections.findCorrection(typed.word)
		if correction != "" {
			kt.pause()
			// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
			// TODO: use an accurate number for the latency
			time.Sleep(60 * time.Millisecond)
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word plus the punctuation mark.
			log.Debugf("Making correction %s to %s", typed.word, correction)
			for i := 0; i <= len(typed.word); i++ {
				kt.kbd.TypeBackSpace()
			}
			// Insert the replacement.
			// Type out the replacement and whatever punctuation/delimiter was after it.
			kt.kbd.TypeString(correction + string(typed.punct))
			// if kt.ShowCorrections {
			// 	notificationData := &control.NotificationData{
			// 		Title:   "Correction!",
			// 		Message: fmt.Sprintf("Replaced %s with %s", typed.word, correction),
			// 	}
			// }
			kt.resume()
		}
		stats := &control.StatsMsg{
			Word:       typed.word,
			Correction: correction,
		}
		kt.statsCh <- stats
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
		kbd:             LinuxKeyboard.NewLinuxKeyboard(LinuxKeyboard.FindKeyboardDevice()),
		kbdEvents:       make(chan LinuxKeyboard.KeyboardEvent),
		typedWord:       make(chan typed),
		paused:          true,
		ShowCorrections: false,
		corrections:     newCorrections(),
		statsCh:         make(chan *control.StatsMsg),
	}
}

func sendStats(manager *control.ConnManager, statsCh chan *control.StatsMsg) {
	for s := range statsCh {
		manager.SendStats(s)
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
