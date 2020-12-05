package wordstats

// WordStats stores counters for words checked and words corrected
type WordStats struct {
	wordsChecked   int
	wordsCorrected int
}

// AddChecked will increment the words checked counter in a wordStats struct
func (w *WordStats) AddChecked() {
	w.wordsChecked++
}

// AddCorrected will increment the words corrected counter in a wordStats struct
func (w *WordStats) AddCorrected() {
	w.wordsCorrected++
}

// CalcAccuracy returns the "accuracy" for the current session
// accuracy is measured as how close to not correcting any words
func (w *WordStats) CalcAccuracy() float32 {
	return (1 - float32(w.wordsCorrected)/float32(w.wordsChecked)) * 100
}

// NewWordStats creates a new wordStats struct
func NewWordStats() *WordStats {
	w := WordStats{
		wordsChecked:   0,
		wordsCorrected: 0,
	}
	return &w
}
