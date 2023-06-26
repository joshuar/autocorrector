// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

import (
	"context"
	_ "embed"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/joshuar/autocorrector/internal/corrections"
	"github.com/joshuar/autocorrector/internal/handler"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/joshuar/autocorrector/internal/word"
	"github.com/joshuar/autocorrector/internal/wordstats"
	"github.com/rs/zerolog/log"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var debugAppID = ""

var keyTracker *keytracker.KeyTracker

const (
	Name      = "autocorrector"
	fyneAppID = "com.github.joshuar.autocorrector"
)

type App struct {
	app               fyne.App
	tray              fyne.Window
	Name, Version     string
	showNotifications bool
}

func New() *App {
	return &App{
		app:               newUI(),
		Name:              Name,
		Version:           Version,
		showNotifications: false,
	}
}

func (a *App) Run() {
	appCtx, cancelfunc := context.WithCancel(context.Background())
	handler := handler.NewHandler()
	keyTracker = keytracker.NewKeyTracker(handler.WordCh)
	corrections := corrections.NewCorrections()
	stats := wordstats.RunStats()

	go func() {
		for {
			select {
			case <-appCtx.Done():
				return
			case notification := <-handler.NotificationsCh:
				if a.showNotifications {
					a.app.SendNotification(&notification)
				}
			}
		}
	}()

	go func() {
		for newWord := range handler.WordCh {
			log.Debug().Msgf("Checking word: %s", newWord.Word)
			stats.Checked <- newWord.Word
			if correction, ok := corrections.CheckWord(newWord.Word); ok {
				handler.CorrectionCh <- word.WordDetails{
					Word:       newWord.Word,
					Correction: correction,
					Punct:      newWord.Punct,
				}
			}
		}
	}()

	go func() {
		for correction := range handler.CorrectionCh {
			keyTracker.CorrectWord(correction)
			stats.Corrected <- [2]string{correction.Word, correction.Correction}
			handler.NotificationsCh <- fyne.Notification{
				Title:   "Correction!",
				Content: fmt.Sprintf("Corrected %s with %s", correction.Word, correction.Correction),
			}
		}
	}()

	a.setupSystemTray()
	a.app.Run()
	cancelfunc()
}
