package wordstats

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type wordAction struct {
	Word       string
	Action     string
	Correction string
	Timestamp  time.Time
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

func newWordAction(word string, action string, correction string) *wordAction {
	return &wordAction{
		Word:       word,
		Action:     action,
		Correction: correction,
		Timestamp:  time.Now().UTC(),
	}
}
