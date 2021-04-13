package keytracker

import (
	"bytes"
	"fmt"
	"time"
	"unicode"

	"github.com/fsnotify/fsnotify"
	ipc "github.com/james-barrow/golang-ipc"
	"github.com/joshuar/autocorrector/internal/wordstats"
	"github.com/joshuar/go-linuxkeyboard/pkg/LinuxKeyboard"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type State string

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	events          *LinuxKeyboard.LinuxKeyboard
	typedWord       chan typed
	Pause           bool
	ShowCorrections bool
	corrections     *corrections
	Signaller       chan State
	ipcServer       *ipc.Server
}

func (kt *KeyTracker) EventWatcher(wordStats *wordstats.WordStats) {
	go kt.slurpWords()
	go kt.checkWord(wordStats)
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	log.Debug("Start snooping keys...")
	ev := kt.events.StartSnooping()
	for {
		for e := range ev {
			switch {
			case kt.Pause:
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
					<-ev
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
}

type typed struct {
	word  string
	punct rune
}

func (kt *KeyTracker) checkWord(stats *wordstats.WordStats) {
	for typed := range kt.typedWord {
		correction := kt.corrections.findCorrection(typed.word)
		if correction != "" {
			stats.AddCorrected(typed.word, correction)
			kt.Pause = true
			// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
			// TODO: use an accurate number for the latency
			time.Sleep(60 * time.Millisecond)
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word plus the punctuation mark.
			log.Debugf("Making correction %s to %s", typed.word, correction)
			for i := 0; i <= len(typed.word); i++ {
				kt.events.TypeBackSpace()
			}
			// Insert the replacement.
			// Type out the replacement and whatever delimiter was after it.
			log.Debugf("punctuation is '%v'", string(typed.punct))
			kt.events.TypeString(correction + string(typed.punct))
			// kt.events.TypeString(string(punctuation))
			if kt.ShowCorrections {
				kt.ipcServer.Write(22, []byte(fmt.Sprintf("Replaced %s with %s", typed.word, correction)))
				// beeep.Notify("Correction!", fmt.Sprintf("Replaced %s with %s", typed.word, correction), "")
			}
			kt.Pause = false
		} else {
			stats.AddChecked(typed.word)
		}
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker(sc *ipc.Server) *KeyTracker {
	ch := make(chan struct{})
	close(ch)
	return &KeyTracker{
		events:          LinuxKeyboard.NewLinuxKeyboard(LinuxKeyboard.FindKeyboardDevice()),
		typedWord:       make(chan typed),
		Pause:           true,
		ShowCorrections: true,
		corrections:     newCorrections(),
		Signaller:       make(chan State),
		ipcServer:       sc,
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
