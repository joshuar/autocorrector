package wordstats

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/adrg/xdg"
	log "github.com/sirupsen/logrus"

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
	Word       string
	action     string
	Correction string
	Timestamp  string
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
		err := b.Put([]byte(corrected.Timestamp), encode(corrected))
		return err
	})
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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

// ShowStats logs top-level statistics about corrections
func (w *WordStats) ShowStats() {
	var statsDetails string
	statsDetails += fmt.Sprintf("%v words checked. ", w.GetCheckedTotal())
	statsDetails += fmt.Sprintf("%v words corrected. ", w.GetCorrectedTotal())
	statsDetails += fmt.Sprintf("Accuracy is: %.2f %%.", w.CalcAccuracy())
	log.Info(statsDetails)
}

// GetStats returns top-level statistics about corrections as a formatted string
func (w *WordStats) GetStats() string {
	var statsDetails string
	statsDetails += fmt.Sprintf("%v words checked. ", w.GetCheckedTotal())
	statsDetails += fmt.Sprintf("%v words corrected. ", w.GetCorrectedTotal())
	statsDetails += fmt.Sprintf("Accuracy is: %.2f %%.", w.CalcAccuracy())
	return statsDetails
}

// ShowLog prints out the full log of corrections history
func (w *WordStats) ShowLog() {
	var fullLog string
	fullLog = "Correction Log:\n"
	w.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(correctionsBucket))
		b.ForEach(func(k, v []byte) error {
			logEntry := decode(v)
			fullLog += fmt.Sprintf("%s: replaced %s with %s\n", k, logEntry.Word, logEntry.Correction)
			return nil
		})
		return nil
	})
	log.Info(fullLog)
}

func encode(logEntry *wordAction) []byte {
	encoded, err := json.Marshal(logEntry)
	if err != nil {
		log.Error(err)
	}
	return encoded
}

func decode(blob []byte) *wordAction {
	var logEntry wordAction
	err := json.Unmarshal(blob, &logEntry)
	if err != nil {
		log.Error(err)
	}
	return &logEntry
}

// OpenWordStats creates a new wordStats struct
func OpenWordStats() *WordStats {
	// open the on-disk database
	statsDbFile, err := xdg.DataFile("autocorrector/stats.db")
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Using statsdb at %s", statsDbFile)
	db, err := bolt.Open(statsDbFile, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(fmt.Errorf("error reading stats database (is autocorrector still running?): %s", err))
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
	return &wordAction{
		Word:       word,
		action:     action,
		Correction: correction,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}
}
