package wordstats

import (
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"

	"github.com/nutsdb/nutsdb"
)

const (
	countersBucket    = "counters"
	correctedKey      = "correctedTotal"
	checkedKey        = "checkedTotal"
	correctionsBucket = "correctionsLog"
	dbFileSuffix      = "autocorrector/stats.nutsdb"
)

// WordStats stores counters for words checked and words corrected
type WordStats struct {
	db        *nutsdb.DB
	Checked   chan string
	Corrected chan [2]string
}

func openDB() *nutsdb.DB {
	// open the on-disk database
	statsDbFile, err := xdg.DataFile(dbFileSuffix)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open stats db.")
	}
	log.Debug().Caller().
		Msgf("Using statsdb at %s", statsDbFile)
	opt := nutsdb.DefaultOptions
	opt.Dir = statsDbFile
	db, err := nutsdb.Open(opt)
	if err != nil {
		log.Fatal().Err(err).
			Msg("Error reading stats database.")
		os.Exit(1)
	}
	return db
}

func (w *WordStats) get(key, bucket string) []byte {
	var value []byte
	if err := w.db.View(
		func(tx *nutsdb.Tx) error {
			if e, err := tx.Get(bucket, []byte(key)); err != nil {
				return err
			} else {
				value = e.Value
			}
			return nil
		}); err != nil {
		log.Debug().Caller().
			Msgf("Couldn't get value for key %s from bucket %s: %v", key, bucket, err)
	}
	return value
}

func (w *WordStats) set(bucket, key string, value interface{}) {
	log.Debug().Caller().
		Msgf("Writing to bucket %s: %s = %v", bucket, key, value)
	var valueBuf []byte
	switch v := value.(type) {
	case uint64:
		valueBuf = make([]byte, binary.MaxVarintLen64)
		binary.PutUvarint(valueBuf, v)
	case *wordAction:
		valueBuf = encode(v)
	}
	if err := w.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(bucket, []byte(key), valueBuf, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Debug().Caller().Err(err).
			Msgf("Couldn't set value for key %s from bucket %s.", key, bucket)
	}
}

func (w *WordStats) addChecked() {
	for range w.Checked {
		checkedTotal, _ := binary.Uvarint(w.get(checkedKey, countersBucket))
		w.set(countersBucket, checkedKey, checkedTotal+1)
	}
}

func (w *WordStats) addCorrected() {
	for c := range w.Corrected {
		correctedTotal, _ := binary.Uvarint(w.get(correctedKey, countersBucket))
		w.set(countersBucket, correctedKey, correctedTotal+1)
		corrected := newWordAction(c[0], "corrected", c[1])
		w.set(correctionsBucket, corrected.Timestamp.Format(time.RFC3339), corrected)
	}
}

// CalcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
func (w *WordStats) CalcAccuracy() float64 {
	checkedTotal, _ := binary.Uvarint(w.get(checkedKey, countersBucket))
	correctedTotal, _ := binary.Uvarint(w.get(correctedKey, countersBucket))
	return (1 - float64(correctedTotal)/float64(checkedTotal)) * 100
}

// GetCheckedTotal fetches the total number of checked words from the database
func (w *WordStats) GetCheckedTotal() uint64 {
	v, _ := binary.Uvarint(w.get(checkedKey, countersBucket))
	return v
}

// GetCorrectedTotal fetches the total number of corrected words from the database
func (w *WordStats) GetCorrectedTotal() uint64 {
	v, _ := binary.Uvarint(w.get(correctedKey, countersBucket))
	return v
}

// CloseWordStats closes the stats database cleanly
func (w *WordStats) CloseWordStats() {
	w.db.Close()
}

// GetStats returns top-level statistics about corrections as a formatted string
func (w *WordStats) GetStats() string {
	var statsDetails string
	statsDetails += fmt.Sprintf("%v words checked. ", w.GetCheckedTotal())
	statsDetails += fmt.Sprintf("%v words corrected. ", w.GetCorrectedTotal())
	statsDetails += fmt.Sprintf("Accuracy is: %.2f%%.", w.CalcAccuracy())
	return statsDetails
}

// ShowStats logs top-level statistics about corrections
func (w *WordStats) ShowStats() {
	log.Info().Msg(w.GetStats())
}

// ShowLog prints out the full log of corrections history
func (w *WordStats) ShowLog() {
	w.db.View(
		func(tx *nutsdb.Tx) error {
			entries, err := tx.GetAll(correctionsBucket)
			if err != nil {
				log.Warn().Err(err).Msg("Couldn't fetch entries from corrections log.")
			}

			for _, entry := range entries {
				w := decode(entry.Value)
				log.Info().Msgf("%s %s %s → %s", string(entry.Key), w.Action, w.Word, w.Correction)
			}
			return nil
		})
}

// OpenWordStats creates a new wordStats struct
func RunStats() *WordStats {
	w := &WordStats{
		db:        openDB(),
		Checked:   make(chan string, 1),
		Corrected: make(chan [2]string, 1),
	}
	go w.addChecked()
	go w.addCorrected()
	return w
}
