package keytracker

type wordDetails struct {
	Word, Correction string
	Punct            rune
}

// NewWord creates a struct to hold a word, its correction and the
// punctuation mark that follows it
func NewWord(w string, c string, p rune) *wordDetails {
	return &wordDetails{
		Word:       w,
		Correction: c,
		Punct:      p,
	}
}
