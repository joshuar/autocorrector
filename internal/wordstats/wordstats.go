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
	bolt "go.etcd.io/bbolt"
)

const (
	countersBucket    = "counters"
	correctionsBucket = "correctionsLog"
)

// WordStats stores counters for words checked and words corrected
type WordStats struct {
	db *bolt.DB
}

type wordAction struct {
	word       string
	action     string
	correction string
	timestamp  time.Time
	ID         uint64
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
	err := w.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(correctionsBucket))
		id, _ := b.NextSequence()
		corrected.ID = id
		idBuf := make([]byte, binary.MaxVarintLen64)
		binary.PutUvarint(idBuf, corrected.ID)
		err := b.Put(idBuf, []byte(corrected.encode()))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (w *WordStats) readAsInt(key string) uint64 {
	var valueAsBuf []byte
	var valueAsInt uint64
	w.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(countersBucket))
		valueAsBuf = b.Get([]byte(key))
		return nil
	})
	valueAsInt, _ = binary.Uvarint(valueAsBuf)
	return valueAsInt
}

func (w *WordStats) writeAsInt(key string, value uint64) {
	valueBuf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(valueBuf, value)
	err := w.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(countersBucket))
		err := b.Put([]byte(key), valueBuf)
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

// CalcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
func (w *WordStats) CalcAccuracy() float64 {
	checkedTotal := w.readAsInt("checkedTotal")
	correctedTotal := w.readAsInt("correctedTotal")
	return (1 - float64(correctedTotal)/float64(checkedTotal)) * 100
}

// GetCheckedTotal fetches the total number of checked words from the database
func (w *WordStats) GetCheckedTotal() uint64 {
	return w.readAsInt("checkedTotal")
}

// GetCorrectedTotal fetches the total number of corrected words from the database
func (w *WordStats) GetCorrectedTotal() uint64 {
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

// OpenWordStats creates a new wordStats struct
func OpenWordStats() *WordStats {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal(fmt.Errorf("Fatal finding home directory: %s", err))
	}

	// open the on-disk database
	db, err := bolt.Open(strings.Join([]string{home, "/.config/autocorrector/stats.db"}, ""), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(fmt.Errorf("Error reading stats database (is autocorrector still running?): %s", err))
		os.Exit(1)
	}
	// make sure the top-level buckets exist
	db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(countersBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(correctionsBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	return &WordStats{
		db: db,
	}
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
