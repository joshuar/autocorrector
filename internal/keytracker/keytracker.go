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
	kbdEvents                 <-chan kbd.KeyEvent
	TypedWord, WordCorrection chan wordDetails
	paused                    bool
}

// StartEvents opens the stats database, starts a goroutine to "slurp" words,
// starts a goroutine to check for corrections
func (kt *KeyTracker) StartEvents() {
	go kt.slurpWords()
	go kt.correctWords()
}

// Pause will stop corrections
func (kt *KeyTracker) Pause() error {
	log.Debug("Pausing keytracker...")
	kt.paused = true
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
				if charBuf.Len() > 0 {
					w := NewWord(charBuf.String(), "", k.AsRune)
					charBuf.Reset()
					kt.TypedWord <- *w
				}
			default:
				// for all other keys, including Ctrl, Meta, Alt, Shift, ignore
				charBuf.Reset()
			}
		}
	}
	kt.CloseKeyTracker()
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
	return &KeyTracker{
		kbd:            kbd.NewVirtualKeyboard("autocorrector"),
		kbdEvents:      kbd.SnoopAllKeyboards(kbd.OpenKeyboardDevices()),
		WordCorrection: make(chan wordDetails),
		TypedWord:      make(chan wordDetails),
		paused:         true,
	}
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

type patternBuf struct {
	buf      *bytes.Buffer
	len, idx int
}

func newPatternBuf(size int) *patternBuf {
	return &patternBuf{
		buf: new(bytes.Buffer),
		len: size,
		idx: 0,
	}
}

func (pb *patternBuf) write(r rune) {
	if pb.idx == 3 {
		pb.buf.Reset()
		pb.idx = 0
	}
	pb.buf.WriteRune(r)
}

func (pb *patternBuf) match(s string) bool {
	return bytes.HasSuffix(pb.buf.Bytes(), []byte(s))
}
