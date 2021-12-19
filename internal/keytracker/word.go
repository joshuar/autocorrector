package keytracker

type WordDetails struct {
	Word, Correction string
	Punct            rune
}

// NewWord creates a struct to hold a word, its correction and the
// punctuation mark that follows it
func NewWord(w string, c string, p rune) *WordDetails {
	return &WordDetails{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
}
