package keytracker

import (
	"bytes"
	"unicode"

	kbd "github.com/joshuar/gokbd"
	log "github.com/sirupsen/logrus"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd                       *kbd.VirtualKeyboardDevice
	kbdEvents                 chan kbd.KeyEvent
	TypedWord, WordCorrection chan wordDetails
	paused                    bool
}

// StartEvents opens the stats database, starts a goroutine to "slurp" words,
// starts a goroutine to check for corrections
func (kt *KeyTracker) StartEvents() {
	go kt.slurpWords()
	go kt.correctWords()
}

// Start will start corrections
func (kt *KeyTracker) Start() error {
	log.Debug("Starting keytracker...")
	kt.paused = false
	return nil
}

// Pause will stop corrections
func (kt *KeyTracker) Pause() error {
	log.Debug("Pausing keytracker...")
	kt.paused = true
	// close(kt.kbdEvents)
	return nil
}

// Resume will resume corrections
func (kt *KeyTracker) Resume() error {
	log.Debug("Resuming keytracker...")
	kt.paused = false
	return nil
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	for k := range kt.kbdEvents {
		switch {
		case k.IsKeyRelease():
			log.Debugf("Key released: %s %s %d\n", k.TypeName, k.EventName, k.Value)
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
				if charBuf.Len() > 0 {
					w := NewWord(charBuf.String(), "", k.AsRune)
					charBuf.Reset()
					kt.TypedWord <- *w
				}
			default:
				// for all other keys, including Ctrl, Meta, Alt, Shift, ignore
				charBuf.Reset()
			}
		case k.IsKeyPress():
			log.Debugf("Key pressed: %s %s %d %c\n", k.TypeName, k.EventName, k.Value, k.AsRune)
		case k.Value == 2 && k.TypeName == "EV_KEY":
			log.Debugf("Key held: %s %s %d %c\n", k.TypeName, k.EventName, k.Value, k.AsRune)
		}
	}
}

func (kt *KeyTracker) correctWords() {
	for w := range kt.WordCorrection {
		// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
		// TODO: use an accurate number for the latency
		// time.Sleep(60 * time.Millisecond)
		if !kt.paused {
			log.Debugf("Making correction %s to %s", w.Word, w.Correction)
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word plus the punctuation mark.
			for i := 0; i <= len(w.Word); i++ {
				kt.kbd.TypeBackspace()
			}
			// Insert the replacement.
			// Type out the replacement and whatever punctuation/delimiter was after it.
			kt.kbd.TypeString(w.Correction + string(w.Punct))
		}
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	kt.kbd.Close()
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	kt := &KeyTracker{
		kbd:            kbd.NewVirtualKeyboard("autocorrector"),
		kbdEvents:      make(chan kbd.KeyEvent),
		WordCorrection: make(chan wordDetails),
		TypedWord:      make(chan wordDetails),
		paused:         true,
	}
	err := kbd.SnoopAllKeyboards(kt.kbdEvents)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	return kt
}

type wordDetails struct {
	Word, Correction string
	Punct            rune
}

// NewWord creates a struct to hold a word, its correction and the
// punctuation mark that follows it
func NewWord(w string, c string, p rune) *wordDetails {
	return &wordDetails{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
}
