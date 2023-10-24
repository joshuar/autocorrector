// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

import (
	_ "embed"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/joshuar/autocorrector/internal/db"
	"github.com/rs/zerolog/log"
)

//go:embed assets/urls/issueURL.txt
var issueURL string

//go:embed assets/urls/featureRequestURL.txt
var featureRequestURL string

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

func (a *App) setupSystemTray(stats *db.Stats) {
	a.tray = a.app.NewWindow("System Tray")
	a.tray.SetMaster()
	if desk, ok := a.app.(desktop.App); ok {
		menuItemQuit := fyne.NewMenuItem("Quit", func() {
			close(a.Done)
		})
		menuItemQuit.IsQuit = true
		menuItemAbout := fyne.
			NewMenuItem("About",
				func() {
					w := a.app.NewWindow(fmt.Sprintf("About %s", a.Name))
					w.SetContent(container.New(layout.NewVBoxLayout(),
						widget.NewLabel(fmt.Sprintf("App Version: %s", a.Version)),
						widget.NewButton("Ok", func() {
							w.Close()
						}),
					))
					w.Show()
				})
		menuItemToggleNotifications := fyne.
			NewMenuItem("Toggle Notifications",
				func() {
					a.showNotifications = !a.showNotifications
				})
		menuItemToggleKeyTracker := fyne.
			NewMenuItem("Toggle Corrections",
				func() {
					log.Debug().Msg("Toggling corrections.")
					a.Toggle()
				})
		menuItemIssue := fyne.
			NewMenuItem("Report Issue",
				func() {
					url, _ := url.Parse(issueURL)
					a.app.OpenURL(url)
				})
		menuItemFeatureRequest := fyne.
			NewMenuItem("Request Feature",
				func() {
					url, _ := url.Parse(featureRequestURL)
					a.app.OpenURL(url)
				})
		menuItemSettings := fyne.
			NewMenuItem("Settings", a.settingsWindow)
		menuItemStats := fyne.
			NewMenuItem("Show Stats", func() {
				a.statsWindow(stats)
			})
		menu := fyne.NewMenu(a.Name,
			menuItemAbout,
			menuItemSettings,
			menuItemStats,
			menuItemToggleNotifications,
			menuItemToggleKeyTracker,
			menuItemIssue,
			menuItemFeatureRequest,
			menuItemQuit)
		desk.SetSystemTrayMenu(menu)
	}
	a.tray.Hide()
}

func (a *App) settingsWindow() {
	w := a.app.NewWindow("Fyne Settings")
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	w.Show()
}

func (a *App) statsWindow(stats *db.Stats) {
	w := a.app.NewWindow("Stats")
	content := container.New(layout.NewVBoxLayout(),
		container.New(layout.NewHBoxLayout(),
			layout.NewSpacer(),
			widget.NewLabel("Lifetime Stats"),
			layout.NewSpacer()),
		container.New(layout.NewGridLayout(3),
			widget.NewLabel(fmt.Sprintf("Checked: %d", stats.GetCheckedTotal())),
			widget.NewLabel(fmt.Sprintf("Corrected: %d", stats.GetCorrectedTotal())),
			widget.NewLabel(fmt.Sprintf("Accuracy: %.2f%%", stats.GetAccuracy()))),
		container.New(layout.NewGridLayout(3),
			widget.NewLabel(fmt.Sprintf("Keys Pressed: %d", stats.GetKeysPressed())),
			widget.NewLabel(fmt.Sprintf("Backspace Pressed: %d", stats.GetBackspacePressed())),
			widget.NewLabel(fmt.Sprintf("Correction Rate: %.2f%%", stats.GetEfficiency()))))
	w.SetContent(content)
	w.Resize(fyne.NewSize(164, 144))
	w.Show()
}
