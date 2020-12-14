package wordstats

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
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
	checkedTotal := w.readAsInt("checkedTotal")
	w.writeAsInt("checkedTotal", checkedTotal+1)
}

// AddCorrected will increment the words corrected counter in a wordStats struct
func (w *WordStats) AddCorrected(word string, correction string) {
	correctedTotal := w.readAsInt("correctedTotal")
	w.writeAsInt("correctedTotal", correctedTotal+1)
	corrected := newWordAction(word, "corrected", correction)
	w.db.Put([]byte(corrected.timestamp.String()), []byte(corrected.encode()))
	log.Debugf("Added correction %v for %v to database", correction, word)
}

func (w *WordStats) readAsInt(key string) uint64 {
	if w.db.Has([]byte(key)) {
		valueAsBuf, _ := w.db.Get([]byte(key))
		valueAsInt, _ := binary.Uvarint(valueAsBuf)
		return valueAsInt
	} else {
		return 0
	}
}

func (w *WordStats) writeAsInt(key string, value uint64) {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, value)
	w.db.Put([]byte(key), buf)
}

// CalcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
func (w *WordStats) CalcAccuracy() float64 {
	checkedTotal := w.readAsInt("checkedTotal")
	correctedTotal := w.readAsInt("checkedTotal")
	return (1 - float64(correctedTotal)/float64(checkedTotal)) * 100
}

func (w *WordStats) ShowCheckedTotal() uint64 {
	return w.readAsInt("checkedTotal")
}

func (w *WordStats) ShowCorrectedTotal() uint64 {
	return w.readAsInt("correctedTotal")
}

// CloseWordStats closes the stats database cleanly
func (w *WordStats) CloseWordStats() {
	w.db.Close()
}

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

	db, err := bitcask.Open(strings.Join([]string{home, "/.config/autocorrector/stats.db"}, ""))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
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
