// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"fyne.io/fyne/v2"
	"github.com/joshuar/autocorrector/internal/db"
	"github.com/joshuar/autocorrector/internal/keytracker"
	"github.com/rs/zerolog/log"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var debugAppID = ""

var configPath = filepath.Join(os.Getenv("HOME"), ".config", "autocorrector")

const (
	Name      = "autocorrector"
	fyneAppID = "com.github.joshuar.autocorrector"
)

type App struct {
	app               fyne.App
	tray              fyne.Window
	Name, Version     string
	showNotifications bool
	notificationsCh   chan *keytracker.Correction
	paused            bool
	toggleCh          chan bool
	Done              chan struct{}
}

func (a *App) NotificationCh() chan *keytracker.Correction {
	return a.notificationsCh
}

func (a *App) Toggle() {
	a.paused = !a.paused
	a.toggleCh <- a.paused
}

func New() *App {
	return &App{
		app:               newUI(),
		Name:              Name,
		Version:           Version,
		showNotifications: false,
		notificationsCh:   make(chan *keytracker.Correction),
		toggleCh:          make(chan bool),
		Done:              make(chan struct{}),
	}
}

func (a *App) Run() {
	var wg sync.WaitGroup
	appCtx, cancelfunc := context.WithCancel(context.Background())
	if err := createDir(configPath); err != nil {
		log.Fatal().Err(err).Msg("Could not create config directory.")
	}

	stats, err := db.RunStats(appCtx, configPath)
	defer close(stats.Done)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start stats tracking.")
	}

	keyTracker, err := keytracker.NewKeyTracker(a, stats)
	defer close(keyTracker.ToggleCh)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start keytracker.")
	}

	wg.Add(1)
	go func() {
		for {
			select {
			case <-appCtx.Done():
				return
			case n := <-a.notificationsCh:
				if a.showNotifications {
					a.app.SendNotification(&fyne.Notification{
						Title:   "Correction!",
						Content: fmt.Sprintf("Corrected %s with %s", n.Word, n.Correction),
					})
				}
			case v := <-a.toggleCh:
				keyTracker.ToggleCh <- v
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Debug().Msg("Ctrl-C pressed.")
		close(a.Done)
	}()
	go func() {
		<-a.Done
		stats.Save()
		log.Debug().Msg("Agent done.")
		os.Exit(0)
	}()
	go func() {
		<-appCtx.Done()
		log.Debug().Msg("Context canceled.")
		os.Exit(1)
	}()

	a.setupSystemTray(stats)
	a.app.Run()
	wg.Wait()
	cancelfunc()
}

func (a *App) Stop() {
	close(a.Done)
}

func createDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		log.Debug().Str("directory", path).Msg("No config directory, creating new one.")
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
