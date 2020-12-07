package keytracker

import (
	"regexp"
	"sort"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	log "github.com/sirupsen/logrus"
)

// KeyTracker holds the channels for handling key presses and
// indicating when word/line delimiter characters are encountered or
// backspace is pressed
type KeyTracker struct {
	Key       chan rune
	WordDelim chan bool
	LineDelim chan bool
	Backspace chan bool
	Disabled  chan bool
	events    chan hook.Event
}

// SnoopKeys listens for key presses and fires on the appropriate channel
func (kt *KeyTracker) SnoopKeys() {
	// wordChar represents any standard character that would make up part of a word
	wordChar, _ := regexp.Compile("[[:alnum:]']")
	// wordDelim represents punctauation and space characters that indicate the end of a word
	wordDelim, _ := regexp.Compile("[[:punct:][:blank:]]")
	// lineDeline are linefeed/return characters indicating a new line was started
	lineDelim, _ := regexp.Compile("[\n\r\f]")
	// otherControlKey are the raw keycodes for various navigational keys like home, end, pgup, pgdown
	// and the arrow keys.
	otherControlKey := []int{65360, 65361, 65362, 65363, 65364, 65367, 65365, 65366}

	kt.events = robotgo.EventStart()

	// here we listen for key presses and match the key pressed against the regex patterns or raw keycodes above
	// depending on what key was pressed, we fire on the appropriate channel to do something about it
	log.Info("Listening for keypresses...")
	for e := range kt.events {
		log.Debug("Got keypress: ", e.Keychar, " : ", string(e.Keychar))
		switch {
		case wordChar.MatchString(string(e.Keychar)):
			kt.Key <- e.Keychar
		case wordDelim.MatchString(string(e.Keychar)):
			kt.Key <- e.Keychar
			kt.WordDelim <- true
		case lineDelim.MatchString(string(e.Keychar)):
			kt.LineDelim <- true
		case e.Keychar == 8:
			kt.Backspace <- true
		case sort.SearchInts(otherControlKey, int(e.Rawcode)) > 0:
			kt.LineDelim <- true
		default:
			log.Debugf("Unknown key pressed: %v", e)
		}
	}
}

// NewKeyTracker creates a new keyTracker struct
func NewKeyTracker() *KeyTracker {
	k := make(chan rune)
	w := make(chan bool)
	l := make(chan bool)
	b := make(chan bool)
	d := make(chan bool)
	kt := KeyTracker{
		Key:       k,
		WordDelim: w,
		LineDelim: l,
		Backspace: b,
		Disabled:  d,
	}
	return &kt
}
