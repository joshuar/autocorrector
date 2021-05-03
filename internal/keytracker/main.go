package keytracker

import (
	"bytes"
	"time"
	"unicode"

	"github.com/joshuar/go-linuxkeyboard/pkg/LinuxKeyboard"

	log "github.com/sirupsen/logrus"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd                       *LinuxKeyboard.LinuxKeyboard
	kbdEvents                 chan LinuxKeyboard.KeyboardEvent
	TypedWord, WordCorrection chan wordDetails
	paused                    bool
}

// EventWatcher opens the stats database, starts a goroutine to "slurp" words,
// starts a goroutine to check for corrections and opens a socket for server
// control
func (kt *KeyTracker) StartEvents() {
	go kt.kbd.Snoop(kt.kbdEvents)
	go kt.slurpWords()
	go kt.correctWords()
}

func (kt *KeyTracker) Start() error {
	kt.paused = false
	// kt.kbdEvents = make(chan LinuxKeyboard.KeyboardEvent)
	// go kt.slurpWords()
	// go kt.kbd.Snoop(kt.kbdEvents)
	return nil
}

func (kt *KeyTracker) Pause() error {
	kt.paused = true
	// close(kt.kbdEvents)
	return nil
}

func (kt *KeyTracker) Resume() error {
	return kt.Start()
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	for {
		e, ok := <-kt.kbdEvents
		if !ok {
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
					w := NewWord(charBuf.String(), "", e.AsRune)
					charBuf.Reset()
					kt.TypedWord <- *w
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
	for w := range kt.WordCorrection {
		kt.Pause()
		// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
		// TODO: use an accurate number for the latency
		time.Sleep(60 * time.Millisecond)
		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word plus the punctuation mark.
		log.Debugf("Making correction %s to %s", w.Word, w.Correction)
		for i := 0; i <= len(w.Word); i++ {
			kt.kbd.TypeBackSpace()
		}
		// Insert the replacement.
		// Type out the replacement and whatever punctuation/delimiter was after it.
		kt.kbd.TypeString(w.Correction + string(w.Punct))
		kt.Resume()
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
		WordCorrection: make(chan wordDetails),
		TypedWord:      make(chan wordDetails),
		paused:         true,
	}
}

type wordDetails struct {
	Word, Correction string
	Punct            rune
}

func NewWord(w string, c string, p rune) *wordDetails {
	return &wordDetails{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
}
