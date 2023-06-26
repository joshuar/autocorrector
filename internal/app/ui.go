// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newUI() fyne.App {
	var a fyne.App
	if debugAppID != "" {
		a = app.NewWithID(debugAppID)
		a.SetIcon(theme.FyneLogo())

	} else {
		a = app.NewWithID(fyneAppID)
		a.SetIcon(trayIcon{})
	}
	return a
}

func (a *App) setupSystemTray() {
	// a.hassConfig = hass.GetConfig(a.config.RestAPIURL)
	a.tray = a.app.NewWindow("System Tray")
	a.tray.SetMaster()
	if desk, ok := a.app.(desktop.App); ok {
		menuItemAbout := fyne.NewMenuItem("About", func() {
			w := a.app.NewWindow(fmt.Sprintf("About %s", a.Name))
			w.SetContent(container.New(layout.NewVBoxLayout(),
				widget.NewLabel(fmt.Sprintf("App Version: %s", a.Version)),
				widget.NewButton("Ok", func() {
					w.Close()
				}),
			))
			w.Show()
		})
		menuItemToggleNotifications := fyne.NewMenuItem("Toggle Notifications", func() {
			a.showNotifications = !a.showNotifications
		})
		menuItemToggleKeyTracker := fyne.NewMenuItem("Toggle Corrections", func() {
			keyTracker.Toggle()
		})
		menu := fyne.NewMenu(a.Name,
			menuItemAbout,
			menuItemToggleNotifications,
			menuItemToggleKeyTracker)
		desk.SetSystemTrayMenu(menu)
	}
	a.tray.Hide()
}
