package keytracker

import (
	"bytes"
	"context"
	"unicode"
	"unicode/utf8"

	"github.com/joshuar/autocorrector/internal/corrections"
	kbd "github.com/joshuar/gokbd"
	"github.com/rs/zerolog/log"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd          *kbd.VirtualKeyboardDevice
	kbdEvents    <-chan kbd.KeyEvent
	ControlCh    chan interface{}
	WordCh       chan WordDetails
	CorrectionCh chan WordDetails
	corrections  *corrections.Corrections
	paused       bool
}

func (kt *KeyTracker) controller() {
	for d := range kt.ControlCh {
		switch d := d.(type) {
		case bool:
			log.Debug().Caller().
				Msgf("Keytracker is paused? %v", d)
			kt.paused = d
		default:
			log.Debug().Caller().
				Msgf("Unexpected data %T on notification channel: %v", d, d)
		}
	}
}

func (kt *KeyTracker) slurpWords() {
	charBuf := new(bytes.Buffer)
	patternBuf := newPatternBuf(3)
	log.Debug().Msg("Slurping words...")
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
			case k.AsRune == '\n' || unicode.IsControl(k.AsRune):
				// newline or control character, reset the buffer
				charBuf.Reset()
			case unicode.IsPunct(k.AsRune), unicode.IsSymbol(k.AsRune), unicode.IsSpace(k.AsRune):
				// a punctuation mark, which would indicate a word has been typed, so handle that
				//
				// most other punctuation should indicate end of word, so
				// handle that
				if charBuf.Len() > 0 {
					kt.WordCh <- WordDetails{
						Word:  charBuf.String(),
						Punct: k.AsRune,
					}
					charBuf.Reset()
				}
			default:
				// case unicode.IsDigit(k.AsRune), unicode.IsLetter(k.AsRune):
				// a letter or number
				_, err := charBuf.WriteRune(k.AsRune)
				if err != nil {
					log.Debug().Caller().Err(err).
						Msgf("Failed to write %v to character buffer.", k.AsRune)
				}
			}
		}
	}
	kt.CloseKeyTracker()
}

func (kt *KeyTracker) checkWords() {
	for w := range kt.WordCh {
		log.Debug().Msgf("Checking word: %s", w.Word)
		if correction, ok := kt.corrections.CheckWord(w.Word); ok {
			kt.CorrectionCh <- WordDetails{
				Word:       w.Word,
				Correction: correction,
				Punct:      w.Punct,
			}
		}
	}
}

func (kt *KeyTracker) correctWords() {
	for c := range kt.CorrectionCh {
		if !kt.paused {
			log.Debug().Msgf("Making correction %s to %s", c.Word, c.Correction)
			// Erase the existing word.
			// Effectively, hit backspace key for the length of the word plus the punctuation mark.
			for i := 0; i <= utf8.RuneCountInString(c.Word); i++ {
				kt.kbd.TypeBackspace()
			}
			// Insert the replacement.
			// Type out the replacement and whatever punctuation/delimiter was after it.
			kt.kbd.TypeString(c.Correction + string(c.Punct))
		}
	}
}

func (kt *KeyTracker) Paused() bool {
	return kt.paused
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	kt.kbd.Close()
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker(ctx context.Context) *KeyTracker {
	vKbd, err := kbd.NewVirtualKeyboard("autocorrector")
	if err != nil {
		log.Error().Err(err).Msg("Could not open a new virtual keyboard.")
		return nil
	}
	kt := &KeyTracker{
		kbd:          vKbd,
		kbdEvents:    kbd.SnoopAllKeyboards(kbd.OpenAllKeyboardDevices()),
		ControlCh:    make(chan interface{}),
		WordCh:       make(chan WordDetails),
		CorrectionCh: make(chan WordDetails),
		paused:       false,
		corrections:  corrections.NewCorrections(),
	}
	go kt.controller()
	go kt.slurpWords()
	go kt.checkWords()
	go kt.correctWords()
	return kt
}
