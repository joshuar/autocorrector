package wordstats

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	log "github.com/sirupsen/logrus"

	"github.com/xujiajun/nutsdb"
)

const (
	countersBucket    = "counters"
	correctionsBucket = "correctionsLog"
	dbFileSuffix      = "autocorrector/stats.nutsdb"
)

// WordStats stores counters for words checked and words corrected
type WordStats struct {
	db        *nutsdb.DB
	Checked   chan string
	Corrected chan [2]string
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
		log.Warnf("Couldn't get value for key %s from bucket %s: %v", key, bucket, err)
	}
	return value
}

func (w *WordStats) set(key, bucket string, i interface{}) {
	var value []byte
	switch v := i.(type) {
	case uint64:
		value = make([]byte, binary.MaxVarintLen64)
		binary.PutUvarint(value, v)
	case wordAction:
		value = encode(&v)
	}
	if err := w.db.Update(
		func(tx *nutsdb.Tx) error {
			if err := tx.Put(bucket, []byte(key), value, 0); err != nil {
				return err
			}
			return nil
		}); err != nil {
		log.Warnf("Couldn't set value for key %s from bucket %s: %v", key, bucket, err)
	}
}

func (w *WordStats) addChecked() {
	for range w.Checked {
		checkedTotal, _ := binary.Uvarint(w.get("checkedTotal", countersBucket))
		w.set("checkedTotal", countersBucket, checkedTotal+1)
	}
}

func (w *WordStats) addCorrected() {
	for c := range w.Corrected {
		correctedTotal, _ := binary.Uvarint(w.get("correctedTotal", countersBucket))
		w.set("correctedTotal", countersBucket, correctedTotal+1)
		corrected := newWordAction(c[0], "corrected", c[1])
		w.set(corrected.Timestamp, correctionsBucket, corrected)
	}
}

// CalcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
func (w *WordStats) CalcAccuracy() float64 {
	checkedTotal, _ := binary.Uvarint(w.get("checkedTotal", countersBucket))
	correctedTotal, _ := binary.Uvarint(w.get("correctedTotal", countersBucket))
	return (1 - float64(correctedTotal)/float64(checkedTotal)) * 100
}

// GetCheckedTotal fetches the total number of checked words from the database
func (w *WordStats) GetCheckedTotal() uint64 {
	v, _ := binary.Uvarint(w.get("checkedTotal", countersBucket))
	return v
}

// GetCorrectedTotal fetches the total number of corrected words from the database
func (w *WordStats) GetCorrectedTotal() uint64 {
	v, _ := binary.Uvarint(w.get("correctedTotal", countersBucket))
	return v
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

	// var fullLog string
	// fullLog = "Correction Log:\n"
	// w.db.View(func(tx *bolt.Tx) error {
	// 	// Assume bucket exists and has keys
	// 	b := tx.Bucket([]byte(correctionsBucket))
	// 	b.ForEach(func(k, v []byte) error {
	// 		logEntry := decode(v)
	// 		fullLog += fmt.Sprintf("%s: replaced %s with %s\n", k, logEntry.Word, logEntry.Correction)
	// 		return nil
	// 	})
	// 	return nil
	// })
	log.Info("not implemented yet")
}

// OpenWordStats creates a new wordStats struct
func OpenWordStats() *WordStats {
	// open the on-disk database
	statsDbFile, err := xdg.DataFile(dbFileSuffix)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Using statsdb at %s", statsDbFile)

	opt := nutsdb.DefaultOptions
	opt.Dir = statsDbFile
	db, err := nutsdb.Open(opt)
	if err != nil {
		log.Fatal(fmt.Errorf("error reading stats database: %s", err))
		os.Exit(1)
	}

	w := &WordStats{
		db:        db,
		Checked:   make(chan string, 1),
		Corrected: make(chan [2]string, 1),
	}
	go w.addChecked()
	go w.addCorrected()
	return w
}
