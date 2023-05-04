// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

import (
	"context"
	_ "embed"

	"fyne.io/fyne/v2"
	"github.com/joshuar/autocorrector/internal/client"
)

//go:generate sh -c "printf %s $(git tag | tail -1) > VERSION"
//go:embed VERSION
var Version string

var debugAppID = ""

const (
	Name      = "autocorrector"
	fyneAppID = "com.github.joshuar.autocorrector"
)

type App struct {
	app           fyne.App
	tray          fyne.Window
	Name, Version string
}

func New() *App {
	return &App{
		app:     newUI(),
		Name:    Name,
		Version: Version,
	}
}

func (a *App) Run() {
	_, cancelfunc := context.WithCancel(context.Background())
	client.Start()
	a.setupSystemTray()
	a.app.Run()
	cancelfunc()
}
