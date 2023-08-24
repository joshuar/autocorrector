// Copyright (c) 2023 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package app

import (
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
)

const (
	issueURL          = "https://github.com/joshuar/autocorrector/issues/new?assignees=joshuar&labels=&template=bug_report.md&title=%5BBUG%5D"
	featureRequestURL = "https://github.com/joshuar/autocorrector/issues/new?assignees=&labels=&template=feature_request.md&title="
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
	a.tray = a.app.NewWindow("System Tray")
	a.tray.SetMaster()
	if desk, ok := a.app.(desktop.App); ok {
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
					keyTracker.Toggle()
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
			NewMenuItem("Show Stats", a.statsWindow)
		menu := fyne.NewMenu(a.Name,
			menuItemAbout,
			menuItemSettings,
			menuItemStats,
			menuItemToggleNotifications,
			menuItemToggleKeyTracker,
			menuItemIssue,
			menuItemFeatureRequest)
		desk.SetSystemTrayMenu(menu)
	}
	a.tray.Hide()
}

func (a *App) settingsWindow() {
	w := a.app.NewWindow("Fyne Settings")
	w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
	w.Show()
}

func (a *App) statsWindow() {
	lifetimeStatsLabel := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		widget.NewLabel("Lifetime Stats"),
		layout.NewSpacer())
	lifetimeStatsChecked := widget.NewLabel(fmt.Sprintf("Checked: %d", stats.GetCheckedTotal()))
	lifetimeStatsCorrected := widget.NewLabel(fmt.Sprintf("Corrected: %d", stats.GetCorrectedTotal()))
	lifetimeStatsAccuracy := widget.NewLabel(fmt.Sprintf("Accuracy: %.2f%%", stats.CalcAccuracy()))
	lifetimeStatsGrid := container.New(layout.NewGridLayout(3),
		lifetimeStatsChecked,
		lifetimeStatsCorrected,
		lifetimeStatsAccuracy)
	w := a.app.NewWindow("Stats")
	content := container.New(layout.NewVBoxLayout(),
		lifetimeStatsLabel,
		lifetimeStatsGrid)
	w.SetContent(content)
	w.Resize(fyne.NewSize(164, 144))
	w.Show()
}
