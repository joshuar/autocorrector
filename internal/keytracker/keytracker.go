// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package keytracker

import (
	"bytes"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/joshuar/autocorrector/internal/corrections"
	kbd "github.com/joshuar/gokbd"
	"github.com/rs/zerolog/log"
)

type stats interface {
	IncKeyCounter()
	IncBackspaceCounter()
	IncCheckedCounter()
	IncCorrectedCounter()
}

type agent interface {
	NotificationCh() chan *Correction
}

type Correction struct {
	Word, Correction string
	Punct            rune
}

func NewCorrection(word, correction string, punct rune) *Correction {
	return &Correction{
		Word:       word,
		Correction: correction,
		Punct:      punct,
	}
}

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	kbd       *kbd.VirtualKeyboardDevice
	kbdEvents <-chan kbd.KeyEvent
	paused    bool
	ToggleCh  chan bool
}

func (kt *KeyTracker) slurpWords(wordCh chan *Correction, stats stats) {
	charBuf := new(bytes.Buffer)
	log.Debug().Msg("Slurping words...")
	for k := range kt.kbdEvents {
		if kt.paused {
			continue
		}
		if k.IsKeyRelease() {
			switch {
			case k.IsBackspace():
				// backspace key
				stats.IncBackspaceCounter()
				if charBuf.Len() > 0 {
					charBuf.Truncate(charBuf.Len() - 1)
				}
			case k.AsRune == '\n' || unicode.IsControl(k.AsRune):
				stats.IncKeyCounter()
				// newline or control character, reset the buffer
				charBuf.Reset()
			case unicode.IsPunct(k.AsRune), unicode.IsSymbol(k.AsRune), unicode.IsSpace(k.AsRune):
				stats.IncKeyCounter()
				// a punctuation mark, which would indicate a word has been typed, so handle that
				//
				// most other punctuation should indicate end of word, so
				// handle that
				if charBuf.Len() > 0 {
					wordCh <- NewCorrection(charBuf.String(), "", k.AsRune)
					charBuf.Reset()
				}
			default:
				stats.IncKeyCounter()
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

func (kt *KeyTracker) checkWord(wordCh chan *Correction, correctionCh chan *Correction, corrections *corrections.Corrections, stats stats) {
	for w := range wordCh {
		log.Debug().Msgf("Checking word: %s", w.Word)
		stats.IncCheckedCounter()
		var ok bool
		if w.Correction, ok = corrections.CheckWord(w.Word); ok {
			correctionCh <- w
		}
	}
}

func (kt *KeyTracker) correctWord(correctionCh chan *Correction, agent agent, stats stats) {
	for correction := range correctionCh {
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
		stats.IncCorrectedCounter()
		agent.NotificationCh() <- correction
	}
}

// CloseKeyTracker shuts down the channels used for key tracking
func (kt *KeyTracker) CloseKeyTracker() {
	kt.kbd.Close()
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker(agent agent, stats stats) (*KeyTracker, error) {
	vKbd, err := kbd.NewVirtualKeyboard("autocorrector")
	if err != nil {
		return nil, err
	}
	kt := &KeyTracker{
		kbd:       vKbd,
		kbdEvents: kbd.SnoopAllKeyboards(kbd.OpenAllKeyboardDevices()),
		paused:    false,
		ToggleCh:  make(chan bool),
	}
	correctionsList, err := corrections.NewCorrections()
	if err != nil {
		return nil, err
	}

	go func() {
		correctionCh := make(chan *Correction)
		defer close(correctionCh)
		wordCh := make(chan *Correction)
		defer close(wordCh)

		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			kt.slurpWords(wordCh, stats)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			kt.checkWord(wordCh, correctionCh, correctionsList, stats)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			kt.correctWord(correctionCh, agent, stats)
		}()
		wg.Add(1)
		go func() {
			for v := range kt.ToggleCh {
				kt.paused = v
				log.Debug().Msgf("Keytracker paused: %t", kt.paused)
			}
		}()
		wg.Wait()
	}()
	return kt, nil
}
