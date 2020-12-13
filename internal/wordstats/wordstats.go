package wordstats

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"
	"github.com/prologic/bitcask"
)

// WordStats stores counters for words checked and words corrected
type WordStats struct {
	db *bitcask.Bitcask
}

type wordAction struct {
	word       string
	action     string
	correction string
	timestamp  time.Time
}

// AddChecked will increment the words checked counter in a wordStats struct
func (w *WordStats) AddChecked(word string) {
	checked := newWordAction(word, "checked", "")
	w.db.Put([]byte(checked.timestamp.String()), []byte(checked.encode()))
	log.Debugf("Add checked word %v to database", word)

}

// AddCorrected will increment the words corrected counter in a wordStats struct
func (w *WordStats) AddCorrected(word string, correction string) {
	corrected := newWordAction(word, "corrected", correction)
	w.db.Put([]byte(corrected.timestamp.String()), []byte(corrected.encode()))
	log.Debugf("Added correction %v for %v to database", correction, word)
}

// CalcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
// func (w *WordStats) CalcAccuracy() float32 {
// 	return (1 - float32(w.wordsCorrected)/float32(w.wordsChecked)) * 100
// }

func (wa *wordAction) encode() []byte {
	encoded, _ := json.Marshal(wa)
	// if err != nil {
	// 	return err
	// }
	return encoded

}

// NewWordStats creates a new wordStats struct
func NewWordStats() *WordStats {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal finding home directory: %s", err))
	}

	db, _ := bitcask.Open(strings.Join([]string{home, "/.config/autocorrector/stats.db"}, ""))
	w := WordStats{
		db: db,
	}
	return &w
}

func newWordAction(word string, action string, correction string) *wordAction {
	timeNow := time.Now()
	wa := wordAction{
		word:       word,
		action:     action,
		correction: correction,
		timestamp:  timeNow,
	}
	return &wa
}
