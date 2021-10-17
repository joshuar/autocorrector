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

func (kt *KeyTracker) Start() error {
	log.Debug("Starting keytracker...")
	kt.paused = false
	// kt.kbdEvents = make(chan LinuxKeyboard.KeyboardEvent)
	// go kt.slurpWords()
	// go kt.kbd.Snoop(kt.kbdEvents)
	return nil
}

func (kt *KeyTracker) Pause() error {
	log.Debug("Pausing keytracker...")
	kt.paused = true
	// close(kt.kbdEvents)
	return nil
}

func (kt *KeyTracker) Resume() error {
	log.Debug("Resuming keytracker...")
	kt.paused = false
	return nil
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	for k := range kt.kbdEvents {
		if k.Value == 1 && k.TypeName == "EV_KEY" {
			log.Debugf("Key pressed: %s %s %d %c\n", k.TypeName, k.EventName, k.Value, k.AsRune)
		}
		if k.Value == 0 && k.TypeName == "EV_KEY" {
			log.Debugf("Key released: %s %s %d\n", k.TypeName, k.EventName, k.Value)
			switch {
			case k.AsRune == rune('\b'):
				// backspace key
				if charBuf.Len() > 0 {
					charBuf.Truncate(charBuf.Len() - 1)
				}
			case unicode.IsDigit(k.AsRune) || unicode.IsLetter(k.AsRune):
				// a letter or number
				charBuf.WriteRune(k.AsRune)
			case unicode.IsPunct(k.AsRune) || unicode.IsSymbol(k.AsRune) || unicode.IsSpace(k.AsRune):
				// a punctuation mark, which would indicate a word has been typed, so handle that
				if charBuf.Len() > 0 {
					w := NewWord(charBuf.String(), "", k.AsRune)
					charBuf.Reset()
					kt.TypedWord <- *w
				}
			case k.EventName == "KEY_LEFTCTRL" || k.EventName == "KEY_RIGHTCTRL" || k.EventName == "KEY_LEFTALT" || k.EventName == "KEY_RIGHTALT" || k.EventName == "KEY_LEFTMETA" || k.EventName == "KEY_RIGHTMETA":
			case k.EventName == "KEY_LEFTSHIFT" || k.EventName == "KEY_RIGHTSHIFT":
			default:
				// for all other keys, including Ctrl, Meta, Alt, Shift, ignore
				charBuf.Reset()
			}
		}
		if k.Value == 2 && k.TypeName == "EV_KEY" {
			log.Debugf("Key held: %s %s %d %c\n", k.TypeName, k.EventName, k.Value, k.AsRune)
		}
	}

	// for {
	// 	e, ok := <-kt.kbdEvents
	// 	if !ok {
	// 		charBuf.Reset()
	// 		return
	// 	}
	// 	switch {
	// 	case kt.paused:
	// 		// don't do anything when we aren't tracking keys, just reset the buffer
	// 		charBuf.Reset()
	// 	case e.Key.IsKeyPress():
	// 		// don't act on key presses, just key releases
	// 		log.Debugf("Pressed key -- value: %d code: %d type: %d string: %s rune: %d (%c)", e.Key.Value, e.Key.Code, e.Key.Type, e.AsString, e.AsRune, e.AsRune)
	// 	case e.Key.IsKeyRelease():
	// 		log.Debugf("Released key -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
	// 		switch {
	// 		case e.AsRune == rune('\b'):
	// 			// backspace key
	// 			if charBuf.Len() > 0 {
	// 				charBuf.Truncate(charBuf.Len() - 1)
	// 			}
	// 		case unicode.IsDigit(e.AsRune) || unicode.IsLetter(e.AsRune):
	// 			// a letter or number
	// 			charBuf.WriteRune(e.AsRune)
	// 		case unicode.IsPunct(e.AsRune) || unicode.IsSymbol(e.AsRune) || unicode.IsSpace(e.AsRune):
	// 			// a punctuation mark, which would indicate a word has been typed, so handle that
	// 			if charBuf.Len() > 0 {
	// 				w := NewWord(charBuf.String(), "", e.AsRune)
	// 				charBuf.Reset()
	// 				kt.TypedWord <- *w
	// 			}
	// 		case e.AsString == "L_CTRL" || e.AsString == "R_CTRL" || e.AsString == "L_ALT" || e.AsString == "R_ALT" || e.AsString == "L_META" || e.AsString == "R_META":
	// 		case e.AsString == "L_SHIFT" || e.AsString == "R_SHIFT":
	// 		default:
	// 			// for all other keys, including Ctrl, Meta, Alt, Shift, ignore
	// 			charBuf.Reset()
	// 		}
	// 	default:
	// 		log.Debugf("Other event -- value: %d code: %d type: %d", e.Key.Value, e.Key.Code, e.Key.Type)
	// 	}

	// }
}

func (kt *KeyTracker) correctWords() {
	for w := range kt.WordCorrection {
		kt.Pause()
		// Before making a correction, add some artificial latency, to ensure the user has actually finished typing
		// TODO: use an accurate number for the latency
		// time.Sleep(60 * time.Millisecond)
		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word plus the punctuation mark.
		log.Debugf("Making correction %s to %s", w.Word, w.Correction)
		for i := 0; i <= len(w.Word); i++ {
			kt.kbd.TypeBackspace()
		}
		// Insert the replacement.
		// Type out the replacement and whatever punctuation/delimiter was after it.
		kt.kbd.TypeString(w.Correction + string(w.Punct))
		kt.Resume()
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	kt.kbd.Close()
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	kt := &KeyTracker{
		kbd:            kbd.NewVirtualKeyboard(),
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

func NewWord(w string, c string, p rune) *wordDetails {
	return &wordDetails{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
}
