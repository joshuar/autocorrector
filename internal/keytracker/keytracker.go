// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package keytracker

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/joshuar/autocorrector/internal/stats"
	"github.com/joshuar/autocorrector/internal/word"
	kbd "github.com/joshuar/gokbd"
	"github.com/rs/zerolog/log"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd       *kbd.VirtualKeyboardDevice
	kbdEvents <-chan kbd.KeyEvent
	paused    bool
}

func (kt *KeyTracker) slurpWords(wordCh chan word.WordDetails, stats *stats.Stats) {
	charBuf := new(bytes.Buffer)
	patternBuf := newPatternBuf(3)
	log.Debug().Msg("Slurping words...")
	for k := range kt.kbdEvents {
		if kt.paused {
			continue
		}
		if k.IsKeyRelease() {
			patternBuf.write(k.AsRune)
			if patternBuf.match(".  ") {
				kt.kbd.TypeBackspace()
			}
			switch {
			case k.IsBackspace():
				// backspace key
				stats.BackSpacePressed.Inc()
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
				stats.KeysPressed.Inc()
				if charBuf.Len() > 0 {
					wordCh <- word.WordDetails{
						Word:  charBuf.String(),
						Punct: k.AsRune,
					}
					charBuf.Reset()
				}
			default:
				// case unicode.IsDigit(k.AsRune), unicode.IsLetter(k.AsRune):
				// a letter or number
				stats.KeysPressed.Inc()
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

func (kt *KeyTracker) CorrectWord(correction word.WordDetails) {
	if !kt.paused {
		log.Debug().Msgf("Making correction %s to %s", correction.Word, correction.Correction)

		// Erase the existing word.
		// Effectively, hit backspace key for the length of the word plus the punctuation mark.
		for i := 0; i <= utf8.RuneCountInString(correction.Word); i++ {
			kt.kbd.TypeBackspace()
		}
		// Insert the replacement.
		// Type out the replacement and whatever punctuation/delimiter was after it.
		kt.kbd.TypeString(correction.Correction + string(correction.Punct))
	}
}

func (kt *KeyTracker) Toggle() {
	kt.paused = !kt.paused
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	kt.kbd.Close()
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker(wordCh chan word.WordDetails, stats *stats.Stats) *KeyTracker {
	vKbd, err := kbd.NewVirtualKeyboard("autocorrector")
	if err != nil {
		log.Error().Err(err).Msg("Could not open a new virtual keyboard.")
		return nil
	}
	kt := &KeyTracker{
		kbd:       vKbd,
		kbdEvents: kbd.SnoopAllKeyboards(kbd.OpenAllKeyboardDevices()),
		paused:    false,
	}
	go kt.slurpWords(wordCh, stats)
	return kt
}
