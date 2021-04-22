package keytracker

import (
	"bytes"
	"time"
	"unicode"

	"github.com/joshuar/autocorrector/internal/control"
	"github.com/joshuar/go-linuxkeyboard/pkg/LinuxKeyboard"

	log "github.com/sirupsen/logrus"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd                       *LinuxKeyboard.LinuxKeyboard
	kbdEvents                 chan LinuxKeyboard.KeyboardEvent
	typedWord, wordCorrection chan wordDetails
	paused                    bool
}

// EventWatcher opens the stats database, starts a goroutine to "slurp" words,
// starts a goroutine to check for corrections and opens a socket for server
// control
func (kt *KeyTracker) EventWatcher(manager *control.ConnManager) {
	go kt.kbd.Snoop(kt.kbdEvents)
	go kt.slurpWords()
	go kt.correctWords()
	manager.SendState(&control.StateMsg{Start: true})
	for {
		select {
		case msg := <-manager.Data:
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
			case *control.WordMsg:
				w := newWord(t.Word, t.Correction, t.Punct)
				kt.wordCorrection <- *w
			default:
				log.Debugf("Unhandled message recieved: %v", msg)
			}
		case w := <-kt.typedWord:
			manager.SendWord(w.word, "", w.punct)
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
			// don't do anything when we aren't tracking keys, just reset the buffer
			charBuf.Reset()
		case e.Key.IsKeyPress():
			// don't act on key presses, just key releases
			log.Debugf("Pressed key -- value: %d code: %d type: %d string: %s rune: %d (%c)", e.Key.Value, e.Key.Code, e.Key.Type, e.AsString, e.AsRune, e.AsRune)
		case e.Key.IsKeyRelease():
			log.Debugf("Released key -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
			switch {
			case e.AsRune == rune('\b'):
				// backspace key
				if charBuf.Len() > 0 {
					charBuf.Truncate(charBuf.Len() - 1)
				}
			case unicode.IsDigit(e.AsRune) || unicode.IsLetter(e.AsRune):
				// a letter or number
				charBuf.WriteRune(e.AsRune)
			case unicode.IsPunct(e.AsRune) || unicode.IsSymbol(e.AsRune) || unicode.IsSpace(e.AsRune):
				// a punctuation mark, which would indicate a word has been typed, so handle that
				if charBuf.Len() > 0 {
					w := newWord(charBuf.String(), "", e.AsRune)
					charBuf.Reset()
					kt.typedWord <- *w
				}
			case e.AsString == "L_CTRL" || e.AsString == "R_CTRL" || e.AsString == "L_ALT" || e.AsString == "R_ALT" || e.AsString == "L_META" || e.AsString == "R_META":
			case e.AsString == "L_SHIFT" || e.AsString == "R_SHIFT":
			default:
				// for all other keys, including Ctrl, Meta, Alt, Shift, ignore
				charBuf.Reset()
			}
		default:
			log.Debugf("Other event -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
		}

	}
}

func (kt *KeyTracker) correctWords() {
	for w := range kt.wordCorrection {
		kt.pause()
		// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
		// TODO: use an accurate number for the latency
		time.Sleep(60 * time.Millisecond)
		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word plus the punctuation mark.
		log.Debugf("Making correction %s to %s", w.word, w.correction)
		for i := 0; i <= len(w.word); i++ {
			kt.kbd.TypeBackSpace()
		}
		// Insert the replacement.
		// Type out the replacement and whatever punctuation/delimiter was after it.
		kt.kbd.TypeString(w.correction + string(w.punct))
		kt.resume()
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	return &KeyTracker{
		kbd:            LinuxKeyboard.NewLinuxKeyboard(LinuxKeyboard.FindKeyboardDevice()),
		kbdEvents:      make(chan LinuxKeyboard.KeyboardEvent),
		wordCorrection: make(chan wordDetails),
		typedWord:      make(chan wordDetails),
		paused:         true,
	}
}

type wordDetails struct {
	word, correction string
	punct            rune
}

func newWord(w string, c string, p rune) *wordDetails {
	return &wordDetails{
		word:       w,
		correction: c,
		punct:      p,
	}
}
