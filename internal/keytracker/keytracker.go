package keytracker

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/joshuar/autocorrector/internal/control"
	kbd "github.com/joshuar/gokbd"
	log "github.com/sirupsen/logrus"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd              *kbd.VirtualKeyboardDevice
	kbdEvents        <-chan kbd.KeyEvent
	ctrlCh           chan interface{}
	wordToCheck      chan string
	correctionToMake chan *control.WordMsg
	paused           bool
}

func (kt *KeyTracker) controller() {
	for d := range kt.ctrlCh {
		switch d := d.(type) {
		case bool:
			log.Debugf("Keytracker is paused? %v", d)
			kt.paused = d
		default:
			log.Debug("Unexpected data %T on notification channel: %v", d, d)
		}
	}
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	patternBuf := newPatternBuf(3)
	for k := range kt.kbdEvents {
		if k.IsKeyRelease() {
			patternBuf.write(k.AsRune)
			if patternBuf.match(".  ") {
				kt.kbd.TypeBackspace()
			}
			switch {
			case k.IsBackspace():
				// backspace key
				if charBuf.Len() > 0 {
					charBuf.Truncate(charBuf.Len() - 1)
				}
			case unicode.IsDigit(k.AsRune), unicode.IsLetter(k.AsRune):
				// a letter or number
				charBuf.WriteRune(k.AsRune)
			case unicode.IsPunct(k.AsRune), unicode.IsSymbol(k.AsRune), unicode.IsSpace(k.AsRune):
				// a punctuation mark, which would indicate a word has been typed, so handle that
				//
				// but if the punctuation is a newline, just reset as it would
				// be difficult to correct a command entered and already
				// swallowed by something...
				if k.AsRune == '\n' {
					charBuf.Reset()
				}
				// most other punctuation should indicate end of word, so
				// handle that
				if charBuf.Len() > 0 {
					log.Debugf("character buffer is '%s'", charBuf.String())
					go kt.checkWord(charBuf.String(), k.AsRune)
					charBuf.Reset()
				}
			default:
				// for all other keys, including Ctrl, Meta, Alt, Shift, ignore
				charBuf.Reset()
			}
		}
	}
	kt.CloseKeyTracker()
}

func (kt *KeyTracker) checkWord(w string, p rune) {
	kt.wordToCheck <- w
	c := <-kt.correctionToMake
	details := NewWord(c.Word, c.Correction, p)
	if details.Correction != "" {
		go kt.correctWord(details)
	}
}

func (kt *KeyTracker) correctWord(w *WordDetails) {
	// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
	if !kt.paused {
		log.Debugf("Making correction %s to %s", w.Word, w.Correction)
		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word plus the punctuation mark.
		for i := 0; i <= utf8.RuneCountInString(w.Word); i++ {
			kt.kbd.TypeBackspace()
		}
		// Insert the replacement.
		// Type out the replacement and whatever punctuation/delimiter was after it.
		kt.kbd.TypeString(w.Correction + string(w.Punct))
		w.Word = ""
		w.Correction = ""
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	kt.kbd.Close()
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() (chan interface{}, chan string, chan *control.WordMsg) {
	kt := KeyTracker{
		kbd:              kbd.NewVirtualKeyboard("autocorrector"),
		kbdEvents:        kbd.SnoopAllKeyboards(kbd.OpenKeyboardDevices()),
		ctrlCh:           make(chan interface{}),
		wordToCheck:      make(chan string),
		correctionToMake: make(chan *control.WordMsg),
		paused:           true,
	}
	go kt.controller()
	go kt.slurpWords()
	return kt.ctrlCh, kt.wordToCheck, kt.correctionToMake
}
