package wordstats

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

type wordAction struct {
	Timestamp  time.Time
	Word       string
	Action     string
	Correction string
}

func encode(logEntry *wordAction) []byte {
	encoded, err := json.Marshal(logEntry)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not encode.")
	}
	return encoded
}

func decode(blob []byte) *wordAction {
	var logEntry wordAction
	err := json.Unmarshal(blob, &logEntry)
	if err != nil {
		log.Debug().Caller().Err(err).
			Msg("Could not decode.")
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
